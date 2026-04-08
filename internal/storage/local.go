package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalStore saves files to the server's local filesystem.
type LocalStore struct {
	root    string
	baseURL string
}

// NewLocal creates a LocalStore rooted at dir.
func NewLocal(root, baseURL string) (*LocalStore, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, fmt.Errorf("storage: create root %q: %w", root, err)
	}
	return &LocalStore{root: root, baseURL: baseURL}, nil
}

func (s *LocalStore) Put(_ context.Context, key string, r io.Reader, _ int64, _ string) (*Object, error) {
	dest := filepath.Join(s.root, filepath.FromSlash(key))

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return nil, fmt.Errorf("storage: mkdir: %w", err)
	}

	f, err := os.Create(dest)
	if err != nil {
		return nil, fmt.Errorf("storage: create file: %w", err)
	}
	defer f.Close()

	n, err := io.Copy(f, r)
	if err != nil {
		return nil, fmt.Errorf("storage: write file: %w", err)
	}

	return &Object{Key: key, SizeBytes: n}, nil
}

func (s *LocalStore) Get(_ context.Context, key string) (io.ReadCloser, *Object, error) {
	path := filepath.Join(s.root, filepath.FromSlash(key))

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, fmt.Errorf("storage: stat: %w", err)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("storage: open: %w", err)
	}

	return f, &Object{Key: key, SizeBytes: info.Size()}, nil
}

func (s *LocalStore) Delete(_ context.Context, key string) error {
	path := filepath.Join(s.root, filepath.FromSlash(key))
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: delete: %w", err)
	}
	return nil
}

func (s *LocalStore) Exists(_ context.Context, key string) (bool, error) {
	_, err := os.Stat(filepath.Join(s.root, filepath.FromSlash(key)))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// URL returns a direct HTTP URL via the base URL + key.
// No expiry logic for local storage (files are always public via Nginx).
func (s *LocalStore) URL(_ context.Context, key string, _ time.Duration) (string, error) {
	return s.baseURL + "/files/" + key, nil
}

func (s *LocalStore) Close() error { return nil }
