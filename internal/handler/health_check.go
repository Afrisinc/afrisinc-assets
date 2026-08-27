package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CheckResult is the outcome of a single dependency health check.
type CheckResult struct {
	Status    string `json:"status"`
	LatencyMs int64  `json:"latencyMs,omitempty"`
	Error     string `json:"error,omitempty"`
}

// dbHealthResult bundles the derived HTTP status alongside the check result.
type dbHealthResult struct {
	statusCode int
	db         CheckResult
}

// checkDBHealth pings the pool with a timeout and reports latency, mirroring
// the shape/semantics of checkDBHealth in gv4_ms_accounts/utils/health.check.ts.
func checkDBHealth(ctx context.Context, db *pgxpool.Pool) dbHealthResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		return dbHealthResult{
			statusCode: http.StatusServiceUnavailable,
			db:         CheckResult{Status: "down", Error: err.Error()},
		}
	}

	return dbHealthResult{
		statusCode: http.StatusOK,
		db:         CheckResult{Status: "up", LatencyMs: time.Since(start).Milliseconds()},
	}
}
