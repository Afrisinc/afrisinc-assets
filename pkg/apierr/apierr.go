package apierr

import (
	"errors"
	"net/http"

	"github.com/afrisinc/assets/internal/repository"
	"github.com/afrisinc/assets/internal/service"
	"github.com/afrisinc/assets/pkg/response"
)

// Handle maps well-known sentinel errors to appropriate HTTP responses.
// Returns true if the error was handled, false if the caller should treat it
// as an unexpected internal error.
func Handle(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, service.ErrNotFound), errors.Is(err, repository.ErrNotFound):
		response.NotFound(w)
		return true
	case errors.Is(err, service.ErrInvalidInput):
		response.BadRequest(w, err.Error())
		return true
	}
	return false
}
