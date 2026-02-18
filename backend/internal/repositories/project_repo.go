package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"rag-chat-system/internal/models"
)

type ProjectRepo struct {
	db *pgxpool.Pool
}

func NewProjectRepo(db *pgxpool.Pool) *ProjectRepo {
	return &ProjectRepo{db: db}
}

func (r *ProjectRepo) Create(ctx context.Context, p *models.Project) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO projects (id, name) VALUES ($1, $2)`,
		p.ID, p.Name,
	)
	return err
}

func (r *ProjectRepo) List(ctx context.Context) ([]models.Project, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, name, git_url, git_branch, last_synced_at, created_at FROM projects ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.GitURL, &p.GitBranch, &p.LastSyncedAt, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (r *ProjectRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM projects WHERE id=$1`, id)
	return err
}

func (r *ProjectRepo) GetByID(ctx context.Context, id string) (*models.Project, error) {
	var p models.Project
	err := r.db.QueryRow(ctx,
		`SELECT id, name, git_url, git_branch, git_token_encrypted, last_synced_at, created_at FROM projects WHERE id=$1`, id,
	).Scan(&p.ID, &p.Name, &p.GitURL, &p.GitBranch, &p.GitTokenEncrypted, &p.LastSyncedAt, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepo) UpdateGitConfig(ctx context.Context, id, gitURL, branch string, encryptedToken *string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE projects SET git_url=$2, git_branch=$3, git_token_encrypted=$4 WHERE id=$1`,
		id, gitURL, branch, encryptedToken,
	)
	return err
}

func (r *ProjectRepo) ClearGitConfig(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE projects SET git_url=NULL, git_branch='main', git_token_encrypted=NULL, last_synced_at=NULL WHERE id=$1`,
		id,
	)
	return err
}

func (r *ProjectRepo) UpdateLastSyncedAt(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE projects SET last_synced_at=NOW() WHERE id=$1`,
		id,
	)
	return err
}
