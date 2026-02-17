package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"rag-chat-system/internal/models"
)

type ChatRepo struct {
	db *pgxpool.Pool
}

func NewChatRepo(db *pgxpool.Pool) *ChatRepo {
	return &ChatRepo{db: db}
}

func (r *ChatRepo) Create(ctx context.Context, c *models.Chat) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO chats (id, title) VALUES ($1, $2)`,
		c.ID, c.Title,
	)
	return err
}

func (r *ChatRepo) List(ctx context.Context) ([]models.Chat, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, title, created_at FROM chats ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var c models.Chat
		if err := rows.Scan(&c.ID, &c.Title, &c.CreatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, c)
	}
	return chats, nil
}

func (r *ChatRepo) UpdateTitle(ctx context.Context, id, title string) error {
	_, err := r.db.Exec(ctx, `UPDATE chats SET title=$1 WHERE id=$2`, title, id)
	return err
}
