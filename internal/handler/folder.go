package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/afrisinc/assets/internal/service"
	"github.com/afrisinc/assets/pkg/apierr"
	"github.com/afrisinc/assets/pkg/response"
	"github.com/afrisinc/assets/pkg/validator"
)

type FolderHandler struct {
	svc *service.FolderService
}

func NewFolderHandler(svc *service.FolderService) *FolderHandler {
	return &FolderHandler{svc: svc}
}

// Create handles POST /api/v1/folders
// Supports creating a single folder or a nested path
func (h *FolderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		ParentID    *string `json:"parent_id"` // for nested folders
		Path        string  `json:"path"`      // for creating full hierarchy (e.g., "marketplace/account-123/templates")
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid JSON body")
		return
	}

	vld := validator.New()

	// If path is provided, create nested hierarchy
	if body.Path != "" {
		vld.Required(body.Path, "path")
		if !vld.Valid() {
			response.ValidationErrors(w, vld.Errors())
			return
		}

		folder, err := h.svc.CreateNested(r.Context(), body.Path)
		if err != nil {
			response.InternalError(w, err)
			return
		}
		response.Created(w, folder)
		return
	}

	// Otherwise create single folder
	vld.Required(body.Name, "name")
	vld.MaxLen(body.Name, 100, "name")
	vld.MaxLen(body.Description, 500, "description")
	if !vld.Valid() {
		response.ValidationErrors(w, vld.Errors())
		return
	}

	folder, err := h.svc.Create(r.Context(), service.CreateFolderInput{
		Name:        body.Name,
		Description: body.Description,
		ParentID:    body.ParentID,
	})
	if err != nil {
		response.InternalError(w, err)
		return
	}
	response.Created(w, folder)
}

// List handles GET /api/v1/folders
func (h *FolderHandler) List(w http.ResponseWriter, r *http.Request) {
	folders, err := h.svc.List(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]any{"folders": folders})
}

// Get handles GET /api/v1/folders/{id}
func (h *FolderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	folder, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if apierr.Handle(w, err) {
			return
		}
		response.InternalError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, folder)
}

// Delete handles DELETE /api/v1/folders/{id}
func (h *FolderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		if apierr.Handle(w, err) {
			return
		}
		response.InternalError(w, err)
		return
	}
	response.NoContent(w)
}
