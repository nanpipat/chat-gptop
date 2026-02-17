package models

type DocumentChunk struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	FileID    string    `json:"file_id"`
	Content   string    `json:"content"`
	Embedding []float32 `json:"-"`
}
