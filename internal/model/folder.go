package model

import "time"

// Folder is a logical grouping of assets (not a file-system directory).
type Folder struct {
	ID          string    `json:"id"`
	ParentID    *string   `json:"parent_id,omitempty"` // nil = root folder
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Path        string    `json:"path,omitempty"`     // full path: /parent/child/name
	Description string    `json:"description,omitempty"`
	AssetCount  int       `json:"asset_count"`
	TotalBytes  int64     `json:"total_bytes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StorageStats is returned by the /stats endpoint.
type StorageStats struct {
	TotalAssets   int   `json:"total_assets"`
	TotalBytes    int64 `json:"total_bytes"`
	TotalFolders  int   `json:"total_folders"`
	ImageCount    int   `json:"image_count"`
	VideoCount    int   `json:"video_count"`
	DocumentCount int   `json:"document_count"`
	FontCount     int   `json:"font_count"`
}
