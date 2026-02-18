package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"rag-chat-system/internal/models"
)

type MessageRepo struct {
	db *pgxpool.Pool
}

func NewMessageRepo(db *pgxpool.Pool) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Create(ctx context.Context, m *models.Message) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO messages (id, chat_id, role, content) VALUES ($1, $2, $3, $4)`,
		m.ID, m.ChatID, m.Role, m.Content,
	)
	return err
}

func (r *MessageRepo) ListByChatID(ctx context.Context, chatID string) ([]models.Message, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, chat_id, role, content, created_at FROM messages WHERE chat_id=$1 ORDER BY created_at ASC`,
		chatID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.ChatID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, nil
}

func (r *MessageRepo) DeleteByChatID(ctx context.Context, chatID string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM messages WHERE chat_id=$1`, chatID)
	return err
}
