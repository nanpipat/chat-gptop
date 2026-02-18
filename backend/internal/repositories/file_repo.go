package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"rag-chat-system/internal/models"
)

type FileRepo struct {
	db *pgxpool.Pool
}

func NewFileRepo(db *pgxpool.Pool) *FileRepo {
	return &FileRepo{db: db}
}

func (r *FileRepo) Create(ctx context.Context, f *models.File) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO files (id, project_id, parent_id, name, path, is_dir) VALUES ($1, $2, $3, $4, $5, $6)`,
		f.ID, f.ProjectID, f.ParentID, f.Name, f.Path, f.IsDir,
	)
	return err
}

func (r *FileRepo) ListByProject(ctx context.Context, projectID string) ([]models.File, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, project_id, parent_id, name, path, is_dir, created_at FROM files WHERE project_id=$1 ORDER BY is_dir DESC, name ASC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var f models.File
		if err := rows.Scan(&f.ID, &f.ProjectID, &f.ParentID, &f.Name, &f.Path, &f.IsDir, &f.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func (r *FileRepo) GetByID(ctx context.Context, id string) (*models.File, error) {
	var f models.File
	err := r.db.QueryRow(ctx,
		`SELECT id, project_id, parent_id, name, path, is_dir, created_at FROM files WHERE id=$1`, id,
	).Scan(&f.ID, &f.ProjectID, &f.ParentID, &f.Name, &f.Path, &f.IsDir, &f.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FileRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM files WHERE id=$1`, id)
	return err
}

func (r *FileRepo) DeleteByProjectID(ctx context.Context, projectID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM files WHERE project_id=$1`, projectID)
	return err
}

func (r *FileRepo) GetChildren(ctx context.Context, parentID string) ([]models.File, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, project_id, parent_id, name, path, is_dir, created_at FROM files WHERE parent_id=$1 ORDER BY is_dir DESC, name ASC`,
		parentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.File
	for rows.Next() {
		var f models.File
		if err := rows.Scan(&f.ID, &f.ProjectID, &f.ParentID, &f.Name, &f.Path, &f.IsDir, &f.CreatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func (r *FileRepo) FindByProjectAndPath(ctx context.Context, projectID, path string) (*models.File, error) {
	var f models.File
	err := r.db.QueryRow(ctx,
		`SELECT id, project_id, parent_id, name, path, is_dir, created_at FROM files WHERE project_id=$1 AND path=$2`,
		projectID, path,
	).Scan(&f.ID, &f.ProjectID, &f.ParentID, &f.Name, &f.Path, &f.IsDir, &f.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &f, nil
}
