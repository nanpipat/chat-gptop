package models

import "time"

type File struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	ParentID  *string   `json:"parent_id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsDir     bool      `json:"is_dir"`
	CreatedAt time.Time `json:"created_at"`
	Children  []File    `json:"children,omitempty"`
}
