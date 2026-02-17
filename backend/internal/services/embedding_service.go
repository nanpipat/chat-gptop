package services

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type EmbeddingService struct {
	openai *OpenAIService
}

func NewEmbeddingService(openaiSvc *OpenAIService) *EmbeddingService {
	return &EmbeddingService{openai: openaiSvc}
}

func (s *EmbeddingService) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	resp, err := s.openai.Client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.SmallEmbedding3,
	})
	if err != nil {
		return nil, err
	}
	return resp.Data[0].Embedding, nil
}
