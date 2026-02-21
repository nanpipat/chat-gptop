package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"rag-chat-system/internal/models"
	"rag-chat-system/internal/repositories"
	"rag-chat-system/internal/storage"
)

type FileService struct {
	fileRepo      *repositories.FileRepo
	chunkRepo     *repositories.ChunkRepo
	ingestService *IngestService
	storage       storage.Storage
	// Keep storagePath just in case for specialized operations, or remove it?
	// It's used for projectDir construction, but projectDir logic needs to change for S3
}

func NewFileService(
	fileRepo *repositories.FileRepo,
	chunkRepo *repositories.ChunkRepo,
	ingestService *IngestService,
	store storage.Storage,
) *FileService {
	return &FileService{
		fileRepo:      fileRepo,
		chunkRepo:     chunkRepo,
		ingestService: ingestService,
		storage:       store,
	}
}

// textExtensions lists file extensions that should be ingested as text
var textExtensions = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true, ".jsx": true,
	".html": true, ".css": true, ".scss": true, ".less": true,
	".json": true, ".yaml": true, ".yml": true, ".toml": true, ".xml": true,
	".md": true, ".txt": true, ".csv": true, ".log": true, ".env": true,
	".sh": true, ".bash": true, ".zsh": true, ".fish": true, ".bat": true, ".ps1": true,
	".sql": true, ".graphql": true, ".gql": true,
	".rs": true, ".rb": true, ".java": true, ".kt": true, ".scala": true,
	".c": true, ".cpp": true, ".h": true, ".hpp": true, ".cs": true,
	".php": true, ".swift": true, ".r": true, ".m": true,
	".dockerfile": true, ".makefile": true, ".gitignore": true,
	".tf": true, ".hcl": true, ".proto": true, ".prisma": true,
	".vue": true, ".svelte": true, ".astro": true,
	".conf": true, ".cfg": true, ".ini": true, ".properties": true,
}

func isTextFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		// Files without extension: check common names
		base := strings.ToLower(filepath.Base(filename))
		switch base {
		case "dockerfile", "makefile", "rakefile", "gemfile", "procfile",
			".gitignore", ".dockerignore", ".env", "readme", "license":
			return true
		}
		return false
	}
	return textExtensions[ext]
}

func isBinaryContent(data []byte) bool {
	return bytes.ContainsRune(data, 0)
}

func (s *FileService) UploadFile(ctx context.Context, projectID string, parentID *string, filename string, reader io.Reader) (*models.File, error) {
	fileID := uuid.New().String()

	// Create object key: projects/{projectID}/{fileID}_{filename}
	// Using fileID prefix prevents name collisions
	objectKey := fmt.Sprintf("projects/%s/%s_%s", projectID, fileID, filename)

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	if err := s.storage.Put(ctx, objectKey, bytes.NewReader(content)); err != nil {
		return nil, fmt.Errorf("write file: %w", err)
	}

	f := &models.File{
		ID:        fileID,
		ProjectID: projectID,
		ParentID:  parentID,
		Name:      filename,
		Path:      objectKey,
		IsDir:     false,
	}

	if err := s.fileRepo.Create(ctx, f); err != nil {
		return nil, fmt.Errorf("insert file: %w", err)
	}

	// Only ingest text files — skip binary files to avoid PostgreSQL UTF8 errors
	if isTextFile(filename) && !isBinaryContent(content) {
		if err := s.ingestService.IngestContent(ctx, projectID, fileID, string(content), filename); err != nil {
			return nil, fmt.Errorf("ingest: %w", err)
		}
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
		_ = s.storage.Delete(ctx, f.Path)
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
			_ = s.storage.Delete(ctx, child.Path)
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
