package storage

import (
	"context"
	"io"
)

// Storage defines the interface for file storage operations
type Storage interface {
	// Put saves content to storage at the given path
	Put(ctx context.Context, path string, content io.Reader) error

	// Get retrieves content from storage at the given path
	Get(ctx context.Context, path string) (io.ReadCloser, error)

	// Delete removes content from storage at the given path
	Delete(ctx context.Context, path string) error
}
