package handler

import (
	"net/http"

	"github.com/afrisinc/assets/pkg/response"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

// Live handles GET /health/live — liveness probe (always up if the process can respond).
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	response.Raw(w, http.StatusOK, map[string]string{"status": "up"})
}

// Ready handles GET /health/ready — readiness probe (checks DB connectivity).
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	dbResult := checkDBHealth(r.Context(), h.db)

	allUp := dbResult.statusCode == http.StatusOK

	status := "healthy"
	statusCode := http.StatusOK
	if !allUp {
		status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	response.Raw(w, statusCode, map[string]any{
		"status": status,
		"db":     dbResult.db,
	})
}
