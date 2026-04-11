package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/afrisinc/assets/internal/config"
	"github.com/afrisinc/assets/internal/model"
	"github.com/afrisinc/assets/internal/service"
	"github.com/afrisinc/assets/pkg/apierr"
	"github.com/afrisinc/assets/pkg/response"
	"github.com/afrisinc/assets/pkg/validator"
)

// AssetHandler wires asset service methods to HTTP endpoints.
type AssetHandler struct {
	svc *service.AssetService
	cfg *config.UploadConfig
}

func NewAssetHandler(svc *service.AssetService, cfg *config.UploadConfig) *AssetHandler {
	return &AssetHandler{svc: svc, cfg: cfg}
}

// Upload handles POST /api/v1/assets
// Accepts multipart/form-data with field "file" and optional "folder_id", "tags".
func (h *AssetHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if e := r.ParseMultipartForm(h.cfg.MaxFileSizeBytes); e != nil {
		response.BadRequest(w, "unable to parse form: "+e.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "field 'file' is required")
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	// Keep only the base type (strip any boundary parameter)
	if i := strings.Index(mimeType, ";"); i != -1 {
		mimeType = strings.TrimSpace(mimeType[:i])
	}

	if !h.isMIMEAllowed(mimeType) {
		response.BadRequest(w, fmt.Sprintf("file type %q is not allowed", mimeType))
		return
	}

	// Buffer file content to avoid streaming/closure race conditions
	// This ensures the file is completely read from the multipart form
	// before being written to storage
	fileBuffer := new(bytes.Buffer)
	if _, err := io.Copy(fileBuffer, file); err != nil {
		response.BadRequest(w, "unable to read file: "+err.Error())
		return
	}

	var folderID *string
	if fid := r.FormValue("folder_id"); fid != "" {
		folderID = &fid
	}

	var tags []string
	if raw := r.FormValue("tags"); raw != "" {
		for _, t := range strings.Split(raw, ",") {
			if t = strings.TrimSpace(t); t != "" {
				tags = append(tags, t)
			}
		}
	}

	in := &service.UploadInput{
		FolderID:     folderID,
		OriginalName: header.Filename,
		MIMEType:     mimeType,
		SizeBytes:    int64(fileBuffer.Len()),
		Reader:       fileBuffer,
		Tags:         tags,
	}

	asset, err := h.svc.Upload(r.Context(), in)
	if err != nil {
		response.InternalError(w, err)
		return
	}

	response.Created(w, asset)
}

// Get handles GET /api/v1/assets/{id}
func (h *AssetHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	asset, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if apierr.Handle(w, err) {
			return
		}
		response.InternalError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, asset)
}

// List handles GET /api/v1/assets
// Query params: folder_id, type, search, tags, page, page_size, sort_by, sort_dir
func (h *AssetHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	params := model.ListAssetParams{
		Search:   q.Get("search"),
		Page:     queryInt(q.Get("page"), 1),
		PageSize: queryInt(q.Get("page_size"), 50),
		SortBy:   q.Get("sort_by"),
		SortDir:  q.Get("sort_dir"),
	}

	if fid := q.Get("folder_id"); fid != "" {
		params.FolderID = &fid
	}
	if t := q.Get("type"); t != "" {
		at := model.AssetType(t)
		params.Type = &at
	}
	if rawTags := q.Get("tags"); rawTags != "" {
		for _, tag := range strings.Split(rawTags, ",") {
			if tag = strings.TrimSpace(tag); tag != "" {
				params.Tags = append(params.Tags, tag)
			}
		}
	}

	result, err := h.svc.List(r.Context(), params)
	if err != nil {
		response.InternalError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, result)
}

// Download handles GET /api/v1/assets/{id}/download
// Streams the raw file bytes with the correct Content-Type and
// Content-Disposition headers so browsers trigger a save dialog.
func (h *AssetHandler) Download(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	asset, rc, err := h.svc.Download(r.Context(), id)
	if err != nil {
		if apierr.Handle(w, err) {
			return
		}
		response.InternalError(w, err)
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", asset.MIMEType)
	w.Header().Set("Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, asset.OriginalName))
	w.Header().Set("Content-Length", strconv.FormatInt(asset.SizeBytes, 10))
	w.Header().Set("Cache-Control", "private, max-age=3600")

	// Stream file directly to avoid seeking issues
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}

// Delete handles DELETE /api/v1/assets/{id}
func (h *AssetHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

// BulkDelete handles DELETE /api/v1/assets with a JSON body {"ids":["a","b"]}
func (h *AssetHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid JSON body")
		return
	}

	vld := validator.New()
	vld.Check(len(body.IDs) > 0, "ids", "ids must not be empty")
	vld.Check(len(body.IDs) <= 100, "ids", "cannot delete more than 100 assets at once")
	if !vld.Valid() {
		response.ValidationErrors(w, vld.Errors())
		return
	}

	var failed []string
	for _, id := range body.IDs {
		if err := h.svc.Delete(r.Context(), id); err != nil {
			if !errors.Is(err, service.ErrNotFound) {
				failed = append(failed, id)
			}
		}
	}

	if len(failed) > 0 {
		response.JSON(w, http.StatusMultiStatus, map[string]any{
			"deleted": len(body.IDs) - len(failed),
			"failed":  failed,
		})
		return
	}

	response.NoContent(w)
}

// Stats handles GET /api/v1/assets/stats
func (h *AssetHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.Stats(r.Context())
	if err != nil {
		response.InternalError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, stats)
}

// isMIMEAllowed checks the uploaded file type against the allow-list.
func (h *AssetHandler) isMIMEAllowed(mimeType string) bool {
	for _, allowed := range h.cfg.AllowedMIMETypes {
		if allowed == mimeType {
			return true
		}
	}
	return false
}

func queryInt(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return fallback
}

