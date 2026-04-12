package router

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/afrisinc/assets/internal/config"
	"github.com/afrisinc/assets/internal/handler"
	"github.com/afrisinc/assets/internal/middleware"
)

// Deps bundles all handler instances the router needs.
type Deps struct {
	Asset   *handler.AssetHandler
	Folder  *handler.FolderHandler
	Health  *handler.HealthHandler
	Docs    *handler.DocsHandler
	Config  *config.Config
	Limiter *middleware.RateLimiter
}

// New builds and returns the fully configured chi router.
func New(d *Deps) http.Handler {
	r := chi.NewRouter()

	// ── Global middleware stack ───────────────────────────────────────────
	r.Use(middleware.Recover)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.StripSlashes)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Tighten in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Content-Disposition", "Content-Length"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// ── Health probes (no auth — Nginx/load balancer hits these) ─────────
	r.Get("/health/live", d.Health.Live)
	r.Get("/health/ready", d.Health.Ready)

	// ── API Documentation (no auth) ─────────────────────────────────────
	r.Get("/api/docs", d.Docs.SwaggerHTML)
	r.Get("/api/docs/openapi.yaml", d.Docs.OpenAPI)

	// ── File serving for local development (in production, Nginx serves this) ──
	// Only enable in development mode
	if os.Getenv("ENVIRONMENT") != "production" {
		fileServer := http.FileServer(http.Dir(d.Config.Storage.LocalRoot))
		r.Get("/files/*", func(w http.ResponseWriter, r *http.Request) {
			// Strip /files prefix and serve
			filePath := strings.TrimPrefix(r.URL.Path, "/files")
			r.URL.Path = filePath
			fileServer.ServeHTTP(w, r)
		})
	}

	// ── API v1 (authenticated) ────────────────────────────────────────────
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(d.Limiter.Limit)
		r.Use(middleware.APIKeyAuth(d.Config.Auth.APIKey))

		// Assets
		r.Route("/assets", func(r chi.Router) {
			r.Get("/", d.Asset.List)          // GET    /api/v1/assets
			r.Post("/", d.Asset.Upload)       // POST   /api/v1/assets
			r.Get("/stats", d.Asset.Stats)    // GET    /api/v1/assets/stats
			r.Delete("/", d.Asset.BulkDelete) // DELETE /api/v1/assets  (body: {"ids":[…]})

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", d.Asset.Get)              // GET    /api/v1/assets/{id}
				r.Delete("/", d.Asset.Delete)        // DELETE /api/v1/assets/{id}
				r.Get("/download", d.Asset.Download) // GET    /api/v1/assets/{id}/download
			})
		})

		// Folders
		r.Route("/folders", func(r chi.Router) {
			r.Get("/", d.Folder.List)    // GET    /api/v1/folders
			r.Post("/", d.Folder.Create) // POST   /api/v1/folders

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", d.Folder.Get)       // GET    /api/v1/folders/{id}
				r.Delete("/", d.Folder.Delete) // DELETE /api/v1/folders/{id}
			})
		})
	})

	// ── 404 catch-all ─────────────────────────────────────────────────────
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"ok":false,"error":"route not found"}`)) //nolint:errcheck
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"ok":false,"error":"method not allowed"}`)) //nolint:errcheck
	})

	return r
}
