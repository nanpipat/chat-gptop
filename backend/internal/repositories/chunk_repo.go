package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	pgvector "github.com/pgvector/pgvector-go"
)

type ChunkRepo struct {
	db *pgxpool.Pool
}

func NewChunkRepo(db *pgxpool.Pool) *ChunkRepo {
	return &ChunkRepo{db: db}
}

func (r *ChunkRepo) Create(ctx context.Context, id, projectID, fileID, content string, embedding []float32) error {
	vec := pgvector.NewVector(embedding)
	_, err := r.db.Exec(ctx,
		`INSERT INTO document_chunks (id, project_id, file_id, content, embedding) VALUES ($1, $2, $3, $4, $5)`,
		id, projectID, fileID, content, vec,
	)
	return err
}

func (r *ChunkRepo) SearchByEmbedding(ctx context.Context, embedding []float32, projectIDs []string, limit int) ([]string, error) {
	vec := pgvector.NewVector(embedding)

	var query string
	var args []interface{}

	if len(projectIDs) > 0 {
		query = `SELECT content FROM document_chunks WHERE project_id = ANY($2) ORDER BY embedding <-> $1 LIMIT $3`
		args = []interface{}{vec, projectIDs, limit}
	} else {
		query = `SELECT content FROM document_chunks ORDER BY embedding <-> $1 LIMIT $2`
		args = []interface{}{vec, limit}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("chunk search: %w", err)
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			return nil, err
		}
		results = append(results, content)
	}
	return results, nil
}

func (r *ChunkRepo) DeleteByFileID(ctx context.Context, fileID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM document_chunks WHERE file_id=$1`, fileID)
	return err
}
