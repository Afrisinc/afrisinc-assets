package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/afrisinc/assets/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AssetRepo handles all database operations for assets.
type AssetRepo struct {
	db *pgxpool.Pool
}

func NewAssetRepo(db *pgxpool.Pool) *AssetRepo {
	return &AssetRepo{db: db}
}

const assetColumns = `
	id, folder_id, name, original_name, mime_type, size_bytes,
	width, height, storage_key, tags, created_at, updated_at`

func scanAsset(row pgx.Row) (*model.Asset, error) {
	a := &model.Asset{}
	err := row.Scan(
		&a.ID, &a.FolderID, &a.Name, &a.OriginalName,
		&a.MIMEType, &a.SizeBytes, &a.Width, &a.Height,
		&a.StorageKey, &a.Tags, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AssetRepo) Create(ctx context.Context, a *model.Asset) error {
	query := `
		INSERT INTO assets
			(id, folder_id, name, original_name, mime_type, size_bytes,
			 width, height, storage_key, tags)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		a.ID, a.FolderID, a.Name, a.OriginalName,
		a.MIMEType, a.SizeBytes, a.Width, a.Height,
		a.StorageKey, a.Tags,
	).Scan(&a.CreatedAt, &a.UpdatedAt)
}

func (r *AssetRepo) GetByID(ctx context.Context, id string) (*model.Asset, error) {
	query := `SELECT ` + assetColumns + ` FROM assets WHERE id = $1 AND deleted_at IS NULL`
	row := r.db.QueryRow(ctx, query, id)
	a, err := scanAsset(row)
	if err == pgx.ErrNoRows {
		return nil, ErrNotFound
	}
	return a, err
}

func (r *AssetRepo) List(ctx context.Context, p model.ListAssetParams) (*model.AssetListResult, error) {
	where := []string{"deleted_at IS NULL"}
	args := []any{}
	argN := 1

	if p.FolderID != nil {
		where = append(where, fmt.Sprintf("folder_id = $%d", argN))
		args = append(args, *p.FolderID)
		argN++
	}
	if p.Type != nil {
		where = append(where, fmt.Sprintf("mime_type LIKE $%d", argN))
		args = append(args, string(*p.Type)+"%")
		argN++
	}
	if p.Search != "" {
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR original_name ILIKE $%d)", argN, argN))
		args = append(args, "%"+p.Search+"%")
		argN++
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM assets ` + whereClause
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Sanitise sort fields to prevent injection
	sortBy := "created_at"
	allowed := map[string]bool{"created_at": true, "name": true, "size_bytes": true}
	if allowed[p.SortBy] {
		sortBy = p.SortBy
	}
	sortDir := "DESC"
	if strings.ToUpper(p.SortDir) == "ASC" {
		sortDir = "ASC"
	}

	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 200 {
		p.PageSize = 50
	}
	offset := (p.Page - 1) * p.PageSize

	dataQuery := fmt.Sprintf(
		`SELECT %s FROM assets %s ORDER BY %s %s LIMIT $%d OFFSET $%d`,
		assetColumns, whereClause, sortBy, sortDir, argN, argN+1,
	)
	args = append(args, p.PageSize, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := make([]*model.Asset, 0, p.PageSize)
	for rows.Next() {
		a, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}

	return &model.AssetListResult{
		Assets:   assets,
		Total:    total,
		Page:     p.Page,
		PageSize: p.PageSize,
		HasMore:  total > p.Page*p.PageSize,
	}, nil
}

func (r *AssetRepo) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE assets SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AssetRepo) Stats(ctx context.Context) (*model.StorageStats, error) {
	s := &model.StorageStats{}
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*)                                            AS total_assets,
			COALESCE(SUM(size_bytes), 0)                       AS total_bytes,
			COUNT(*) FILTER (WHERE mime_type LIKE 'image/%')   AS image_count,
			COUNT(*) FILTER (WHERE mime_type LIKE 'video/%')   AS video_count,
			COUNT(*) FILTER (WHERE mime_type = 'application/pdf') AS document_count,
			COUNT(*) FILTER (WHERE mime_type LIKE 'font/%')    AS font_count
		FROM assets WHERE deleted_at IS NULL`,
	).Scan(
		&s.TotalAssets, &s.TotalBytes,
		&s.ImageCount, &s.VideoCount, &s.DocumentCount, &s.FontCount,
	)
	return s, err
}
