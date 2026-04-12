package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/afrisinc/assets/internal/model"
	"github.com/afrisinc/assets/internal/repository"
	"github.com/afrisinc/assets/pkg/fileutil"
)

type FolderService struct {
	folders *repository.FolderRepo
}

func NewFolderService(repo *repository.FolderRepo) *FolderService {
	return &FolderService{folders: repo}
}

type CreateFolderInput struct {
	Name        string
	Description string
	ParentID    *string // nil = root folder
}

func (s *FolderService) Create(ctx context.Context, in CreateFolderInput) (*model.Folder, error) {
	var path string

	// If parent is specified, fetch parent and build path
	if in.ParentID != nil {
		parent, err := s.folders.GetByID(ctx, *in.ParentID)
		if err != nil {
			return nil, fmt.Errorf("folder create: parent not found: %w", err)
		}
		path = parent.Path + "/" + fileutil.Slugify(in.Name)
	} else {
		path = "/" + fileutil.Slugify(in.Name)
	}

	f := &model.Folder{
		ID:          fileutil.NewID(),
		ParentID:    in.ParentID,
		Name:        in.Name,
		Slug:        fileutil.Slugify(in.Name),
		Path:        path,
		Description: in.Description,
	}
	if err := s.folders.Create(ctx, f); err != nil {
		return nil, fmt.Errorf("folder create: %w", err)
	}
	return f, nil
}

// CreateNested creates a hierarchy of folders from a path like "marketplace/accountId/templates/templateId"
func (s *FolderService) CreateNested(ctx context.Context, path string) (*model.Folder, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var parent *model.Folder

	for _, part := range parts {
		folder, err := s.Create(ctx, CreateFolderInput{
			Name:     part,
			ParentID: func() *string { if parent == nil { return nil }; return &parent.ID }(),
		})
		if err != nil {
			return nil, err
		}
		parent = folder
	}

	return parent, nil
}

func (s *FolderService) GetByID(ctx context.Context, id string) (*model.Folder, error) {
	f, err := s.folders.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return f, err
}

func (s *FolderService) List(ctx context.Context) ([]*model.Folder, error) {
	return s.folders.List(ctx)
}

func (s *FolderService) Delete(ctx context.Context, id string) error {
	err := s.folders.Delete(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
