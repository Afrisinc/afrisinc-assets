package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

// envelope is the top-level JSON wrapper for every response.
type envelope struct {
	OK      bool   `json:"ok"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"` // development only
}

// JSON writes a 200 OK response with the given payload.
func JSON(w http.ResponseWriter, status int, data any) {
	write(w, status, envelope{OK: status < 400, Data: data})
}

// Created writes a 201 Created response.
func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Raw writes a response without the standard ok/data envelope — used for
// health probes and other endpoints that must match an external contract.
func Raw(w http.ResponseWriter, status int, v any) {
	write(w, status, v)
}

// Error writes a structured error response.
func Error(w http.ResponseWriter, status int, msg string) {
	write(w, status, envelope{OK: false, Error: msg})
}

// NotFound writes a 404 response.
func NotFound(w http.ResponseWriter) {
	Error(w, http.StatusNotFound, "resource not found")
}

// BadRequest writes a 400 response with the given message.
func BadRequest(w http.ResponseWriter, msg string) {
	Error(w, http.StatusBadRequest, msg)
}

// Unauthorized writes a 401 response.
func Unauthorized(w http.ResponseWriter) {
	Error(w, http.StatusUnauthorized, "authentication required")
}

// Forbidden writes a 403 response.
func Forbidden(w http.ResponseWriter) {
	Error(w, http.StatusForbidden, "permission denied")
}

// InternalError logs the error and returns a safe 500 to the client.
// In development mode (DEBUG=1), includes the actual error details.
func InternalError(w http.ResponseWriter, err error) {
	slog.Error("internal server error", "error", err)
	msg := "an unexpected error occurred"
	details := ""

	// Include error details in development mode
	if os.Getenv("DEBUG") == "1" {
		details = err.Error()
	}

	write(w, http.StatusInternalServerError, envelope{OK: false, Error: msg, Details: details})
}

// ValidationErrors writes a 422 with a map of field → message.
func ValidationErrors(w http.ResponseWriter, errs map[string]string) {
	write(w, http.StatusUnprocessableEntity, struct {
		OK     bool              `json:"ok"`
		Errors map[string]string `json:"errors"`
	}{OK: false, Errors: errs})
}

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("response encode failed", "error", err)
	}
}
