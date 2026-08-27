package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/afrisinc/assets/internal/config"
	"github.com/afrisinc/assets/internal/server"
)

func main() {
	healthCheck := flag.Bool("health", false, "check the running server's /health/live endpoint and exit 0/1 (used by Docker HEALTHCHECK — the production image has no shell/wget)")
	flag.Parse()

	if *healthCheck {
		os.Exit(runHealthCheck())
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	srv, err := server.New(cfg)
	if err != nil {
		slog.Error("failed to create server", "error", err)
		os.Exit(1)
	}

	// Graceful shutdown on SIGINT / SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("server starting", "addr", cfg.Server.Addr)
		if err := srv.Start(); err != nil {
			slog.Error("server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped cleanly")
}

// runHealthCheck hits the local server's liveness endpoint and returns a
// process exit code (0/1) for use as `ENTRYPOINT ["/server"]`'s Docker
// HEALTHCHECK. It reads SERVER_ADDR directly rather than going through
// config.Load, since the production image has no shell to run wget/curl.
func runHealthCheck() int {
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		port = addr
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://127.0.0.1:%s/health/live", port), nil)
	if err != nil {
		return 1
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
