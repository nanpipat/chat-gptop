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
	}

	for _, q := range queries {
		if _, err := pool.Exec(ctx, q); err != nil {
			log.Fatalf("Migration failed: %v\nQuery: %s", err, q)
		}
	}

	fmt.Println("Migrations complete")
}
