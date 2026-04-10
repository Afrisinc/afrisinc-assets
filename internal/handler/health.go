package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/afrisinc/assets/pkg/response"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

// Live handles GET /health/live — liveness probe (always 200 if process is up).
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Ready handles GET /health/ready — readiness probe (checks DB connectivity).
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		response.JSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "degraded",
			"detail": "database unreachable",
		})
		return
	}

	stat := h.db.Stat()
	response.JSON(w, http.StatusOK, map[string]any{
		"status":            "ok",
		"db_total_conns":    stat.TotalConns(),
		"db_idle_conns":     stat.IdleConns(),
		"db_acquired_conns": stat.AcquiredConns(),
	})
}
