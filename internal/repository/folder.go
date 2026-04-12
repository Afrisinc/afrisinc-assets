package repository

import (
	"context"

	"github.com/afrisinc/assets/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FolderRepo struct {
	db *pgxpool.Pool
}

func NewFolderRepo(db *pgxpool.Pool) *FolderRepo {
	return &FolderRepo{db: db}
}

func (r *FolderRepo) Create(ctx context.Context, f *model.Folder) error {
	query := `
		INSERT INTO folders (id, parent_id, name, slug, path, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`
	return r.db.QueryRow(ctx, query, f.ID, f.ParentID, f.Name, f.Slug, f.Path, f.Description).
		Scan(&f.CreatedAt, &f.UpdatedAt)
}

func (r *FolderRepo) GetByID(ctx context.Context, id string) (*model.Folder, error) {
	f := &model.Folder{}
	err := r.db.QueryRow(ctx, `
		SELECT f.id, f.parent_id, f.name, f.slug, f.path, f.description,
		       COUNT(a.id) AS asset_count,
		       COALESCE(SUM(a.size_bytes), 0) AS total_bytes,
		       f.created_at, f.updated_at
		FROM folders f
		LEFT JOIN assets a ON a.folder_id = f.id AND a.deleted_at IS NULL
		WHERE f.id = $1
		GROUP BY f.id`, id,
	).Scan(&f.ID, &f.ParentID, &f.Name, &f.Slug, &f.Path, &f.Description,
		&f.AssetCount, &f.TotalBytes, &f.CreatedAt, &f.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return f, err
}

func (r *FolderRepo) List(ctx context.Context) ([]*model.Folder, error) {
	rows, err := r.db.Query(ctx, `
		SELECT f.id, f.parent_id, f.name, f.slug, f.path, f.description,
		       COUNT(a.id) AS asset_count,
		       COALESCE(SUM(a.size_bytes), 0) AS total_bytes,
		       f.created_at, f.updated_at
		FROM folders f
		LEFT JOIN assets a ON a.folder_id = f.id AND a.deleted_at IS NULL
		GROUP BY f.id
		ORDER BY f.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []*model.Folder
	for rows.Next() {
		f := &model.Folder{}
		if err := rows.Scan(&f.ID, &f.ParentID, &f.Name, &f.Slug, &f.Path, &f.Description,
			&f.AssetCount, &f.TotalBytes, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		folders = append(folders, f)
	}
	return folders, nil
}

func (r *FolderRepo) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM folders WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
