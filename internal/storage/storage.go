package storage

import (
	"context"
	"io"
	"time"
)

// Object is a lightweight descriptor returned after a successful Put.
type Object struct {
	Key       string
	SizeBytes int64
	ETag      string
}

// Store is the single interface both local-disk and S3 backends implement.
// All keys are slash-separated paths, e.g. "images/2024/01/uuid.webp".
type Store interface {
	// Put stores a reader under the given key.
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (*Object, error)

	// Get returns a ReadCloser for the given key. Caller must close it.
	Get(ctx context.Context, key string) (io.ReadCloser, *Object, error)

	// Delete removes a key. Returns nil if the key did not exist.
	Delete(ctx context.Context, key string) error

	// Exists reports whether a key is present.
	Exists(ctx context.Context, key string) (bool, error)

	// URL returns a public-facing URL for the key.
	// For local storage this is a path; for S3 it is a pre-signed URL.
	URL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// Close releases any resources held by the store (e.g. connection pools).
	Close() error
}
