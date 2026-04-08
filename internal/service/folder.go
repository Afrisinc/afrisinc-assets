package service

import (
	"context"
	"errors"
	"fmt"

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
}

func (s *FolderService) Create(ctx context.Context, in CreateFolderInput) (*model.Folder, error) {
	f := &model.Folder{
		ID:          fileutil.NewID(),
		Name:        in.Name,
		Slug:        fileutil.Slugify(in.Name),
		Description: in.Description,
	}
	if err := s.folders.Create(ctx, f); err != nil {
		return nil, fmt.Errorf("folder create: %w", err)
	}
	return f, nil
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
