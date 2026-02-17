package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"rag-chat-system/internal/models"
	"rag-chat-system/internal/repositories"
)

type FileService struct {
	fileRepo      *repositories.FileRepo
	chunkRepo     *repositories.ChunkRepo
	ingestService *IngestService
	storagePath   string
}

func NewFileService(
	fileRepo *repositories.FileRepo,
	chunkRepo *repositories.ChunkRepo,
	ingestService *IngestService,
	storagePath string,
) *FileService {
	return &FileService{
		fileRepo:      fileRepo,
		chunkRepo:     chunkRepo,
		ingestService: ingestService,
		storagePath:   storagePath,
	}
}

func (s *FileService) UploadFile(ctx context.Context, projectID string, parentID *string, filename string, reader io.Reader) (*models.File, error) {
	projectDir := filepath.Join(s.storagePath, "projects", projectID)
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	fileID := uuid.New().String()
	diskPath := filepath.Join(projectDir, fileID+"_"+filename)

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	if err := os.WriteFile(diskPath, content, 0644); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	f := &models.File{
		ID:        fileID,
		ProjectID: projectID,
		ParentID:  parentID,
		Name:      filename,
		Path:      diskPath,
		IsDir:     false,
	}

	if err := s.fileRepo.Create(ctx, f); err != nil {
		return nil, fmt.Errorf("insert file: %w", err)
	}

	if err := s.ingestService.IngestContent(ctx, projectID, fileID, string(content)); err != nil {
		return nil, fmt.Errorf("ingest: %w", err)
	}

	return f, nil
}

func (s *FileService) EnsureDir(ctx context.Context, projectID string, parentID *string, name, relPath string) (*models.File, error) {
	existing, err := s.fileRepo.FindByProjectAndPath(ctx, projectID, relPath)
	if err == nil {
		return existing, nil
	}

	dirID := uuid.New().String()
	dir := &models.File{
		ID:        dirID,
		ProjectID: projectID,
		ParentID:  parentID,
		Name:      name,
		Path:      relPath,
		IsDir:     true,
	}

	if err := s.fileRepo.Create(ctx, dir); err != nil {
		return nil, fmt.Errorf("create dir record: %w", err)
	}

	return dir, nil
}

func (s *FileService) ListByProject(ctx context.Context, projectID string) ([]models.File, error) {
	files, err := s.fileRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return buildTree(files), nil
}

func (s *FileService) DeleteFile(ctx context.Context, fileID string) error {
	f, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	if f.IsDir {
		if err := s.deleteRecursive(ctx, fileID); err != nil {
			return err
		}
	} else {
		_ = s.chunkRepo.DeleteByFileID(ctx, fileID)
		_ = os.Remove(f.Path)
	}

	return s.fileRepo.Delete(ctx, fileID)
}

func (s *FileService) deleteRecursive(ctx context.Context, parentID string) error {
	children, err := s.fileRepo.GetChildren(ctx, parentID)
	if err != nil {
		return err
	}

	for _, child := range children {
		if child.IsDir {
			if err := s.deleteRecursive(ctx, child.ID); err != nil {
				return err
			}
		} else {
			_ = s.chunkRepo.DeleteByFileID(ctx, child.ID)
			_ = os.Remove(child.Path)
		}
		_ = s.fileRepo.Delete(ctx, child.ID)
	}

	return nil
}

func buildTree(files []models.File) []models.File {
	byID := make(map[string]*models.File)
	for i := range files {
		files[i].Children = []models.File{}
		byID[files[i].ID] = &files[i]
	}

	var roots []models.File
	for i := range files {
		if files[i].ParentID == nil {
			roots = append(roots, files[i])
		} else if parent, ok := byID[*files[i].ParentID]; ok {
			parent.Children = append(parent.Children, files[i])
		}
	}

	// Re-attach children from map back to roots
	var result []models.File
	for _, root := range roots {
		if mapped, ok := byID[root.ID]; ok {
			result = append(result, *mapped)
		}
	}

	if result == nil {
		result = []models.File{}
	}
	return result
}
