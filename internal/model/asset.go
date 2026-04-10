package model

import (
	"time"
)

// Asset represents a single stored file.
type Asset struct {
	ID           string    `json:"id"`
	FolderID     *string   `json:"folder_id,omitempty"`
	Name         string    `json:"name"`
	OriginalName string    `json:"original_name"`
	MIMEType     string    `json:"mime_type"`
	SizeBytes    int64     `json:"size_bytes"`
	Width        *int      `json:"width,omitempty"` // nil for non-image types
	Height       *int      `json:"height,omitempty"`
	StorageKey   string    `json:"-"` // internal storage path, never exposed
	PublicURL    string    `json:"url"`
	Tags         []string  `json:"tags"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AssetType classifies the file category from its MIME type.
type AssetType string

const (
	AssetTypeImage    AssetType = "image"
	AssetTypeVideo    AssetType = "video"
	AssetTypeDocument AssetType = "document"
	AssetTypeFont     AssetType = "font"
	AssetTypeOther    AssetType = "other"
)

func (a *Asset) Type() AssetType {
	switch {
	case len(a.MIMEType) >= 5 && a.MIMEType[:5] == "image":
		return AssetTypeImage
	case len(a.MIMEType) >= 5 && a.MIMEType[:5] == "video":
		return AssetTypeVideo
	case a.MIMEType == "application/pdf":
		return AssetTypeDocument
	case len(a.MIMEType) >= 4 && a.MIMEType[:4] == "font":
		return AssetTypeFont
	default:
		return AssetTypeOther
	}
}

// ListAssetParams defines filters and pagination for listing assets.
type ListAssetParams struct {
	FolderID *string
	Type     *AssetType
	Search   string
	Tags     []string
	Page     int
	PageSize int
	SortBy   string // "created_at" | "name" | "size_bytes"
	SortDir  string // "asc" | "desc"
}

// AssetListResult wraps a page of assets with total count metadata.
type AssetListResult struct {
	Assets   []*Asset `json:"assets"`
	Total    int      `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
	HasMore  bool     `json:"has_more"`
}
