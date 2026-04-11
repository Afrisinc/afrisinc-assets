package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"
	"strings"
	"time"

	"github.com/afrisinc/assets/internal/model"
	"github.com/afrisinc/assets/internal/repository"
	"github.com/afrisinc/assets/internal/storage"
	"github.com/afrisinc/assets/pkg/fileutil"
)

// UploadInput carries all data needed for a single upload operation.
type UploadInput struct {
	FolderID     *string
	OriginalName string
	MIMEType     string
	SizeBytes    int64
	Reader       io.Reader
	Tags         []string
}

// AssetService orchestrates upload, retrieval, deletion and listing.
type AssetService struct {
	assets  *repository.AssetRepo
	store   storage.Store
	baseURL string
}

func NewAssetService(repo *repository.AssetRepo, store storage.Store, baseURL string) *AssetService {
	return &AssetService{assets: repo, store: store, baseURL: baseURL}
}

// Upload stores the file and records its metadata in the database.
func (s *AssetService) Upload(ctx context.Context, in *UploadInput) (*model.Asset, error) {
	id := fileutil.NewID()
	ext := path.Ext(in.OriginalName)
	safeExt := fileutil.SafeExtension(in.MIMEType, ext)
	storageKey := buildKey(in.MIMEType, id, safeExt)

	// For images we decode dimensions before storing.
	// Disabled: dimension extraction caused race condition with concurrent pipe writes
	// when storage implementation uses background goroutines.
	// TODO: Re-implement by buffering to memory first or using post-storage extraction
	var width, height *int

	// Upload the file to storage
	obj, err := s.store.Put(ctx, storageKey, in.Reader, in.SizeBytes, in.MIMEType)
	if err != nil {
		return nil, fmt.Errorf("asset upload: put object: %w", err)
	}

	asset := &model.Asset{
		ID:           id,
		FolderID:     in.FolderID,
		Name:         fileutil.Slugify(strings.TrimSuffix(in.OriginalName, ext)) + safeExt,
		OriginalName: in.OriginalName,
		MIMEType:     in.MIMEType,
		SizeBytes:    obj.SizeBytes,
		Width:        width,
		Height:       height,
		StorageKey:   storageKey,
		Tags:         in.Tags,
	}

	// Ensure tags is never nil for database insert (column is NOT NULL DEFAULT '{}')
	if asset.Tags == nil {
		asset.Tags = []string{}
	}

	if err := s.assets.Create(ctx, asset); err != nil {
		// Roll back the stored object to avoid orphans
		if delErr := s.store.Delete(ctx, storageKey); delErr != nil {
			slog.Error("failed to delete orphan after DB error",
				"key", storageKey, "error", delErr)
		}
		return nil, fmt.Errorf("asset upload: save metadata: %w", err)
	}

	publicURL, _ := s.store.URL(ctx, storageKey, 24*time.Hour)
	asset.PublicURL = publicURL

	return asset, nil
}

func (s *AssetService) GetByID(ctx context.Context, id string) (*model.Asset, error) {
	a, err := s.assets.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	url, _ := s.store.URL(ctx, a.StorageKey, 24*time.Hour)
	a.PublicURL = url
	return a, nil
}

// Download returns a readable stream directly from storage.
func (s *AssetService) Download(ctx context.Context, id string) (*model.Asset, io.ReadCloser, error) {
	a, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	rc, _, err := s.store.Get(ctx, a.StorageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("asset download: %w", err)
	}
	return a, rc, nil
}

func (s *AssetService) List(ctx context.Context, p model.ListAssetParams) (*model.AssetListResult, error) {
	result, err := s.assets.List(ctx, p)
	if err != nil {
		return nil, err
	}
	// Hydrate public URLs
	for _, a := range result.Assets {
		url, _ := s.store.URL(ctx, a.StorageKey, 24*time.Hour)
		a.PublicURL = url
	}
	return result, nil
}

func (s *AssetService) Delete(ctx context.Context, id string) error {
	a, err := s.assets.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	// Soft-delete metadata first
	if err := s.assets.SoftDelete(ctx, id); err != nil {
		return err
	}

	// Best-effort storage delete (log failure but don't fail the request)
	if err := s.store.Delete(ctx, a.StorageKey); err != nil {
		slog.Warn("storage delete failed after soft-delete",
			"id", id, "key", a.StorageKey, "error", err)
	}
	return nil
}

func (s *AssetService) Stats(ctx context.Context) (*model.StorageStats, error) {
	return s.assets.Stats(ctx)
}

// buildKey constructs a deterministic, time-bucketed storage path.
// Example: "images/2024/01/abc123.webp"
func buildKey(mimeType, id, ext string) string {
	prefix := "files"
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		prefix = "images"
	case strings.HasPrefix(mimeType, "video/"):
		prefix = "videos"
	case mimeType == "application/pdf":
		prefix = "documents"
	case strings.HasPrefix(mimeType, "font/"):
		prefix = "fonts"
	}
	now := time.Now()
	return fmt.Sprintf("%s/%d/%02d/%s%s", prefix, now.Year(), now.Month(), id, ext)
}
