package services

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"

	"rag-chat-system/internal/rag"
	"rag-chat-system/internal/repositories"
)

type IngestService struct {
	chunkRepo        *repositories.ChunkRepo
	embeddingService *EmbeddingService
}

func NewIngestService(chunkRepo *repositories.ChunkRepo, embeddingService *EmbeddingService) *IngestService {
	return &IngestService{
		chunkRepo:        chunkRepo,
		embeddingService: embeddingService,
	}
}

func (s *IngestService) IngestContent(ctx context.Context, projectID, fileID, content, fileName string) error {
	fileExt := filepath.Ext(fileName)
	chunks := rag.ChunkText(content, 500, 100)

	for _, chunk := range chunks {
		embedding, err := s.embeddingService.CreateEmbedding(ctx, chunk)
		if err != nil {
			return fmt.Errorf("create embedding: %w", err)
		}

		chunkID := uuid.New().String()
		if err := s.chunkRepo.Create(ctx, chunkID, projectID, fileID, chunk, embedding, fileName, fileExt); err != nil {
			return fmt.Errorf("store chunk: %w", err)
		}
	}

	return nil
}
