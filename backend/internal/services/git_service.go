package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"rag-chat-system/internal/crypto"
	"rag-chat-system/internal/repositories"
)

type syncState struct {
	Status string // "syncing", "done", "error"
	Error  string
}

type GitService struct {
	projectRepo   *repositories.ProjectRepo
	fileRepo      *repositories.FileRepo
	chunkRepo     *repositories.ChunkRepo
	fileService   *FileService
	storagePath   string
	encryptionKey string

	mu         sync.Mutex
	syncStatus map[string]*syncState // projectID -> state
}

func NewGitService(
	projectRepo *repositories.ProjectRepo,
	fileRepo *repositories.FileRepo,
	chunkRepo *repositories.ChunkRepo,
	fileService *FileService,
	storagePath string,
	encryptionKey string,
) *GitService {
	return &GitService{
		projectRepo:   projectRepo,
		fileRepo:      fileRepo,
		chunkRepo:     chunkRepo,
		fileService:   fileService,
		storagePath:   storagePath,
		encryptionKey: encryptionKey,
		syncStatus:    make(map[string]*syncState),
	}
}

// ConfigureGit saves the git configuration for a project, encrypting the PAT if provided.
func (s *GitService) ConfigureGit(ctx context.Context, projectID, gitURL, branch, pat string) error {
	if branch == "" {
		branch = "main"
	}

	var encryptedToken *string
	if pat != "" {
		encrypted, err := crypto.Encrypt(pat, s.encryptionKey)
		if err != nil {
			return fmt.Errorf("encrypt token: %w", err)
		}
		encryptedToken = &encrypted
	}

	return s.projectRepo.UpdateGitConfig(ctx, projectID, gitURL, branch, encryptedToken)
}

// GetConfig returns the git configuration for a project (never exposes the token).
type GitConfig struct {
	GitURL       string  `json:"git_url"`
	GitBranch    string  `json:"git_branch"`
	HasToken     bool    `json:"has_token"`
	LastSyncedAt *string `json:"last_synced_at,omitempty"`
	SyncStatus   string  `json:"sync_status,omitempty"` // "syncing", "done", "error"
	SyncError    string  `json:"sync_error,omitempty"`
}

func (s *GitService) GetConfig(ctx context.Context, projectID string) (*GitConfig, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}

	if project.GitURL == nil || *project.GitURL == "" {
		return nil, nil
	}

	config := &GitConfig{
		GitURL:   *project.GitURL,
		HasToken: project.GitTokenEncrypted != nil && *project.GitTokenEncrypted != "",
	}

	if project.GitBranch != nil {
		config.GitBranch = *project.GitBranch
	} else {
		config.GitBranch = "main"
	}

	if project.LastSyncedAt != nil {
		ts := project.LastSyncedAt.Format("2006-01-02T15:04:05Z")
		config.LastSyncedAt = &ts
	}

	// Include sync status
	s.mu.Lock()
	if st, ok := s.syncStatus[projectID]; ok {
		config.SyncStatus = st.Status
		config.SyncError = st.Error
	}
	s.mu.Unlock()

	return config, nil
}

// RemoveConfig clears the git configuration from a project.
func (s *GitService) RemoveConfig(ctx context.Context, projectID string) error {
	return s.projectRepo.ClearGitConfig(ctx, projectID)
}

// SyncAsync kicks off sync in a background goroutine and returns immediately.
func (s *GitService) SyncAsync(projectID string) error {
	s.mu.Lock()
	if st, ok := s.syncStatus[projectID]; ok && st.Status == "syncing" {
		s.mu.Unlock()
		return fmt.Errorf("sync already in progress")
	}
	s.syncStatus[projectID] = &syncState{Status: "syncing"}
	s.mu.Unlock()

	go func() {
		err := s.Sync(context.Background(), projectID)
		s.mu.Lock()
		if err != nil {
			log.Printf("[GitSync] Background sync failed for %s: %v", projectID, err)
			s.syncStatus[projectID] = &syncState{Status: "error", Error: err.Error()}
		} else {
			s.syncStatus[projectID] = &syncState{Status: "done"}
		}
		s.mu.Unlock()
	}()

	return nil
}

// Sync clones the git repository and re-ingests all files.
func (s *GitService) Sync(ctx context.Context, projectID string) error {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return fmt.Errorf("get project: %w", err)
	}

	if project.GitURL == nil || *project.GitURL == "" {
		return fmt.Errorf("no git URL configured for this project")
	}

	gitURL := *project.GitURL
	branch := "main"
	if project.GitBranch != nil && *project.GitBranch != "" {
		branch = *project.GitBranch
	}

	// Decrypt PAT and embed in URL if present
	if project.GitTokenEncrypted != nil && *project.GitTokenEncrypted != "" {
		token, err := crypto.Decrypt(*project.GitTokenEncrypted, s.encryptionKey)
		if err != nil {
			return fmt.Errorf("decrypt token: %w", err)
		}
		gitURL, err = embedTokenInURL(gitURL, token)
		if err != nil {
			return fmt.Errorf("embed token in URL: %w", err)
		}
	}

	// Clone to temp dir
	tmpDir, err := os.MkdirTemp("", "git-sync-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	log.Printf("[GitSync] Cloning %s (branch: %s) for project %s", *project.GitURL, branch, projectID)

	cmd := exec.CommandContext(ctx, "git", "clone", "--branch", branch, "--depth", "1", gitURL, tmpDir)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s: %w", string(output), err)
	}

	// Delete all existing files and chunks for this project
	log.Printf("[GitSync] Clearing existing files for project %s", projectID)
	if err := s.chunkRepo.DeleteByProjectID(ctx, projectID); err != nil {
		return fmt.Errorf("delete chunks: %w", err)
	}
	if err := s.fileRepo.DeleteByProjectID(ctx, projectID); err != nil {
		return fmt.Errorf("delete files: %w", err)
	}

	// Clean up storage directory
	projectDir := filepath.Join(s.storagePath, "projects", projectID)
	_ = os.RemoveAll(projectDir)

	// Walk cloned files and ingest
	log.Printf("[GitSync] Ingesting files for project %s", projectID)
	dirCache := make(map[string]string) // relPath -> dirID

	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from clone root
		relPath, _ := filepath.Rel(tmpDir, path)
		if relPath == "." {
			return nil
		}

		// Skip .git directory
		if strings.HasPrefix(relPath, ".git") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if info.IsDir() {
			// Create directory entry
			dir := filepath.Dir(relPath)
			var parentID *string
			if dir != "." && dir != "" {
				parentID, err = s.ensureDirChain(ctx, projectID, dir, dirCache)
				if err != nil {
					return fmt.Errorf("ensure parent dir: %w", err)
				}
			}

			d, err := s.fileService.EnsureDir(ctx, projectID, parentID, info.Name(), relPath)
			if err != nil {
				return fmt.Errorf("ensure dir %s: %w", relPath, err)
			}
			dirCache[relPath] = d.ID
			return nil
		}

		// Upload file
		dir := filepath.Dir(relPath)
		var parentID *string
		if dir != "." && dir != "" {
			parentID, err = s.ensureDirChain(ctx, projectID, dir, dirCache)
			if err != nil {
				return fmt.Errorf("ensure parent dir: %w", err)
			}
		}

		file, openErr := os.Open(path)
		if openErr != nil {
			return fmt.Errorf("open file %s: %w", relPath, openErr)
		}
		defer file.Close()

		content, readErr := io.ReadAll(file)
		if readErr != nil {
			return fmt.Errorf("read file %s: %w", relPath, readErr)
		}

		_, uploadErr := s.fileService.UploadFile(ctx, projectID, parentID, info.Name(), bytes.NewReader(content))
		if uploadErr != nil {
			log.Printf("[GitSync] Warning: failed to upload %s: %v", relPath, uploadErr)
			// Don't fail the entire sync for a single file
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk files: %w", err)
	}

	// Update last_synced_at
	if err := s.projectRepo.UpdateLastSyncedAt(ctx, projectID); err != nil {
		return fmt.Errorf("update last_synced_at: %w", err)
	}

	log.Printf("[GitSync] Sync complete for project %s", projectID)
	return nil
}

// ensureDirChain creates all directories in a path chain, reusing cached IDs.
func (s *GitService) ensureDirChain(ctx context.Context, projectID, dirPath string, cache map[string]string) (*string, error) {
	if id, ok := cache[dirPath]; ok {
		return &id, nil
	}

	parts := strings.Split(filepath.ToSlash(dirPath), "/")
	var parentID *string
	accumulated := ""

	for _, part := range parts {
		if part == "" {
			continue
		}
		if accumulated == "" {
			accumulated = part
		} else {
			accumulated = accumulated + "/" + part
		}

		if id, ok := cache[accumulated]; ok {
			parentID = &id
			continue
		}

		dir, err := s.fileService.EnsureDir(ctx, projectID, parentID, part, accumulated)
		if err != nil {
			return nil, err
		}

		cache[accumulated] = dir.ID
		parentID = &dir.ID
	}

	return parentID, nil
}

// embedTokenInURL inserts a PAT token into an HTTPS git URL for authentication.
func embedTokenInURL(gitURL, token string) (string, error) {
	u, err := url.Parse(gitURL)
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}

	u.User = url.UserPassword(token, "")
	return u.String(), nil
}
