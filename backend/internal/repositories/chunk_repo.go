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

func (r *ChunkRepo) Create(ctx context.Context, id, projectID, fileID, content string, embedding []float32, fileName, fileExt string) error {
	vec := pgvector.NewVector(embedding)
	_, err := r.db.Exec(ctx,
		`INSERT INTO document_chunks (id, project_id, file_id, content, embedding, tsv, file_name, file_ext)
		 VALUES ($1, $2, $3, $4, $5, to_tsvector('english', $4), $6, $7)`,
		id, projectID, fileID, content, vec, fileName, fileExt,
	)
	return err
}

// HybridSearch combines vector similarity and full-text search using Reciprocal Rank Fusion (RRF).
func (r *ChunkRepo) HybridSearch(ctx context.Context, embedding []float32, query string, projectIDs []string, limit int) ([]string, error) {
	vec := pgvector.NewVector(embedding)

	var sql string
	var args []interface{}

	if len(projectIDs) > 0 {
		sql = `
		WITH vector_ranked AS (
			SELECT id, content, ROW_NUMBER() OVER (ORDER BY embedding <=> $1) AS rank
			FROM document_chunks
			WHERE project_id = ANY($2)
			ORDER BY embedding <=> $1
			LIMIT $3
		),
		fts_ranked AS (
			SELECT id, content, ROW_NUMBER() OVER (ORDER BY ts_rank(tsv, plainto_tsquery('english', $4)) DESC) AS rank
			FROM document_chunks
			WHERE project_id = ANY($2) AND tsv @@ plainto_tsquery('english', $4)
			LIMIT $3
		)
		SELECT COALESCE(v.content, f.content) AS content
		FROM vector_ranked v
		FULL OUTER JOIN fts_ranked f ON v.id = f.id
		ORDER BY
			COALESCE(1.0 / (60 + v.rank), 0) + COALESCE(1.0 / (60 + f.rank), 0) DESC
		LIMIT $3`
		args = []interface{}{vec, projectIDs, limit, query}
	} else {
		sql = `
		WITH vector_ranked AS (
			SELECT id, content, ROW_NUMBER() OVER (ORDER BY embedding <=> $1) AS rank
			FROM document_chunks
			ORDER BY embedding <=> $1
			LIMIT $2
		),
		fts_ranked AS (
			SELECT id, content, ROW_NUMBER() OVER (ORDER BY ts_rank(tsv, plainto_tsquery('english', $3)) DESC) AS rank
			FROM document_chunks
			WHERE tsv @@ plainto_tsquery('english', $3)
			LIMIT $2
		)
		SELECT COALESCE(v.content, f.content) AS content
		FROM vector_ranked v
		FULL OUTER JOIN fts_ranked f ON v.id = f.id
		ORDER BY
			COALESCE(1.0 / (60 + v.rank), 0) + COALESCE(1.0 / (60 + f.rank), 0) DESC
		LIMIT $2`
		args = []interface{}{vec, limit, query}
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("hybrid search: %w", err)
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

func (r *ChunkRepo) DeleteByProjectID(ctx context.Context, projectID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM document_chunks WHERE project_id=$1`, projectID)
	return err
}
