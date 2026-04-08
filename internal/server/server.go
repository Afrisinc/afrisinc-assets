package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/afrisinc/assets/internal/config"
	"github.com/afrisinc/assets/internal/handler"
	"github.com/afrisinc/assets/internal/middleware"
	"github.com/afrisinc/assets/internal/repository"
	"github.com/afrisinc/assets/internal/router"
	"github.com/afrisinc/assets/internal/service"
	"github.com/afrisinc/assets/internal/storage"
)

// Server wraps the HTTP server and all long-lived resources.
type Server struct {
	http *http.Server
	db   *pgxpool.Pool
	stor storage.Store
}

// New constructs the full dependency graph and returns a ready-to-start Server.
func New(cfg *config.Config) (*Server, error) {
	// ── Database ──────────────────────────────────────────────────────────
	poolCfg, err := pgxpool.ParseConfig(cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("server: parse db dsn: %w", err)
	}
	poolCfg.MaxConns = int32(cfg.DB.MaxConns)
	poolCfg.MinConns = int32(cfg.DB.MinConns)
	poolCfg.MaxConnLifetime = cfg.DB.MaxConnLife

	db, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("server: create db pool: %w", err)
	}
	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("server: db ping: %w", err)
	}

	// ── Storage backend ───────────────────────────────────────────────────
	var stor storage.Store
	switch cfg.Storage.Driver {
	case "local":
		stor, err = storage.NewLocal(cfg.Storage.LocalRoot, cfg.Storage.BaseURL)
		if err != nil {
			return nil, fmt.Errorf("server: init local storage: %w", err)
		}
	case "s3":
		// Swap in storage.NewS3(cfg) once fully implemented.
		return nil, fmt.Errorf("server: S3 storage driver not yet implemented")
	default:
		return nil, fmt.Errorf("server: unknown storage driver %q", cfg.Storage.Driver)
	}

	// ── Repositories ──────────────────────────────────────────────────────
	assetRepo := repository.NewAssetRepo(db)
	folderRepo := repository.NewFolderRepo(db)

	// ── Services ──────────────────────────────────────────────────────────
	assetSvc := service.NewAssetService(assetRepo, stor, cfg.Storage.BaseURL)
	folderSvc := service.NewFolderService(folderRepo)

	// ── Handlers ──────────────────────────────────────────────────────────
	assetH := handler.NewAssetHandler(assetSvc, &cfg.Upload)
	folderH := handler.NewFolderHandler(folderSvc)
	healthH := handler.NewHealthHandler(db)

	// ── Rate limiter: 60 req/s per IP, burst of 20 ───────────────────────
	limiter := middleware.NewRateLimiter(60, 20)

	// ── Router ────────────────────────────────────────────────────────────
	h := router.New(&router.Deps{
		Asset:   assetH,
		Folder:  folderH,
		Health:  healthH,
		Config:  cfg,
		Limiter: limiter,
	})

	httpSrv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      h,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &Server{http: httpSrv, db: db, stor: stor}, nil
}

// Start begins accepting connections. It blocks until the server stops.
func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

// Shutdown gracefully drains in-flight requests, then closes DB and storage.
func (s *Server) Shutdown(ctx context.Context) error {
	defer s.db.Close()
	defer s.stor.Close() //nolint:errcheck
	return s.http.Shutdown(ctx)
}
