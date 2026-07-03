package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

// Storage implements port.Storage using the local filesystem.
// Files are stored under BasePath and served via Gin's Static file server.
type Storage struct {
	basePath string
	logger   *zap.Logger
}

// NewStorage creates a new local disk storage.
// It ensures the base directory exists.
func NewStorage(basePath string, logger *zap.Logger) (*Storage, error) {
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve storage path: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	logger.Info("local storage initialized", zap.String("path", absPath))
	return &Storage{basePath: absPath, logger: logger}, nil
}

// Upload writes the file data to the local disk and returns the URL path.
// The URL is a relative path (e.g., /uploads/<filename>) that can be served
// by Gin's Static file server. When switching to S3, this would return a full URL.
func (s *Storage) Upload(ctx context.Context, data []byte, filename string) (string, error) {
	filePath := filepath.Join(s.basePath, filename)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	s.logger.Debug("file uploaded to local storage",
		zap.String("path", filePath),
		zap.Int("bytes", len(data)),
	)

	// Return relative URL path for Gin static serving
	return fmt.Sprintf("/uploads/%s", filename), nil
}

// Delete removes the file from local disk.
// Returns nil if the file does not exist (idempotent delete).
func (s *Storage) Delete(ctx context.Context, url string) error {
	filename := filepath.Base(url)
	filePath := filepath.Join(s.basePath, filename)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			s.logger.Warn("file not found for deletion", zap.String("url", url))
			return nil
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	s.logger.Debug("file deleted from local storage", zap.String("path", filePath))
	return nil
}
