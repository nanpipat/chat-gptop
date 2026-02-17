package services

import (
	"context"
	"strings"

	"rag-chat-system/internal/repositories"
)

type RAGService struct {
	chunkRepo       *repositories.ChunkRepo
	embeddingService *EmbeddingService
}

func NewRAGService(chunkRepo *repositories.ChunkRepo, embeddingService *EmbeddingService) *RAGService {
	return &RAGService{
		chunkRepo:       chunkRepo,
		embeddingService: embeddingService,
	}
}

func (s *RAGService) SearchRelevantChunks(ctx context.Context, query string, projectIDs []string) ([]string, error) {
	embedding, err := s.embeddingService.CreateEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	return s.chunkRepo.SearchByEmbedding(ctx, embedding, projectIDs, 10)
}

func (s *RAGService) BuildContext(chunks []string) string {
	return strings.Join(chunks, "\n\n---\n\n")
}
