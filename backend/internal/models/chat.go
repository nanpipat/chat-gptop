package models

import "time"

type Chat struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	ProjectIDs []string  `json:"project_ids"`
	CreatedAt  time.Time `json:"created_at"`
}
