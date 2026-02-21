package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(pool *pgxpool.Pool) {
	ctx := context.Background()

	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS vector`,

		`CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS files (
			id UUID PRIMARY KEY,
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			parent_id UUID REFERENCES files(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			is_dir BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS document_chunks (
			id UUID PRIMARY KEY,
			project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
			file_id UUID REFERENCES files(id) ON DELETE CASCADE,
			content TEXT,
			embedding VECTOR(1536)
		)`,

		`CREATE TABLE IF NOT EXISTS chats (
			id UUID PRIMARY KEY,
			title TEXT DEFAULT 'New Chat',
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS messages (
			id UUID PRIMARY KEY,
			chat_id UUID REFERENCES chats(id) ON DELETE CASCADE,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// Git sync columns
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS git_url TEXT`,
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS git_branch TEXT DEFAULT 'main'`,
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS git_token_encrypted TEXT`,
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMP`,

		// Chat project selection
		`ALTER TABLE chats ADD COLUMN IF NOT EXISTS project_ids TEXT[] DEFAULT '{}'`,

		// HNSW index for fast vector similarity search
		`CREATE INDEX IF NOT EXISTS idx_document_chunks_embedding ON document_chunks USING hnsw (embedding vector_cosine_ops)`,

		// Index on project_id for filtered queries
		`CREATE INDEX IF NOT EXISTS idx_document_chunks_project_id ON document_chunks (project_id)`,

		// Full-text search: tsvector column + GIN index + backfill
		`ALTER TABLE document_chunks ADD COLUMN IF NOT EXISTS tsv tsvector`,
		`CREATE INDEX IF NOT EXISTS idx_document_chunks_tsv ON document_chunks USING gin(tsv)`,
		`UPDATE document_chunks SET tsv = to_tsvector('english', content) WHERE tsv IS NULL`,

		// Metadata columns for source attribution
		`ALTER TABLE document_chunks ADD COLUMN IF NOT EXISTS file_name TEXT`,
		`ALTER TABLE document_chunks ADD COLUMN IF NOT EXISTS file_ext TEXT`,
	}

	for _, q := range queries {
		if _, err := pool.Exec(ctx, q); err != nil {
			log.Fatalf("Migration failed: %v\nQuery: %s", err, q)
		}
	}

	fmt.Println("Migrations complete")
}
