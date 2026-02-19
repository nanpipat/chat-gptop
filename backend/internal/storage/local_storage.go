package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage implements Storage using the local filesystem
type LocalStorage struct {
	baseDir string
}

// NewLocalStorage creates a new LocalStorage instance
func NewLocalStorage(baseDir string) *LocalStorage {
	return &LocalStorage{baseDir: baseDir}
}

func (s *LocalStorage) Put(ctx context.Context, path string, content io.Reader) error {
	fullPath := filepath.Join(s.baseDir, path)
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, content)
	return err
}

func (s *LocalStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.baseDir, path)
	return os.Open(fullPath)
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.baseDir, path)
	return os.Remove(fullPath)
}
