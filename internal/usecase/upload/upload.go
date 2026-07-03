package upload

import (
	"crypto/rand"
	"encoding/hex"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/internal/domain/port"
)

// --- File type configuration ---

// fileTypeConfig defines the MaxSize (in bytes) and allowed extensions for each file type.
var fileTypeConfig = map[entity.FileType]struct {
	MaxSize    int64
	Extensions []string
}{
	entity.FileTypeImage: {
		MaxSize:    10 * 1024 * 1024, // 10MB
		Extensions: []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"},
	},
	entity.FileTypeDocument: {
		MaxSize:    50 * 1024 * 1024, // 50MB
		Extensions: []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".txt", ".md"},
	},
	entity.FileTypeVideo: {
		MaxSize:    500 * 1024 * 1024, // 500MB
		Extensions: []string{".mp4", ".mov", ".avi", ".mkv"},
	},
}

// --- DTOs ---

// UploadResponse is the representation of an uploaded file returned to clients.
type UploadResponse struct {
	ID           uint   `json:"id"`
	Type         string `json:"type"`
	OriginalName string `json:"original_name"`
	URL          string `json:"url"`
	Size         int64  `json:"size"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

// FileUploadedEvent is the structured payload for the "file.uploaded" event.
type FileUploadedEvent struct {
	FileID  uint   `json:"file_id"`
	Type    string `json:"type"`
	URL     string `json:"url"`
	OwnerID uint   `json:"owner_id"`
}

// fileToResponse converts a domain entity to the response DTO.
func fileToResponse(f *entity.File) *UploadResponse {
	if f == nil {
		return nil
	}
	return &UploadResponse{
		ID:           f.ID,
		Type:         f.Type,
		OriginalName: f.OriginalName,
		URL:          f.URL,
		Size:         f.Size,
		Status:       string(f.Status),
		CreatedAt:    f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// generateUniqueFilename creates a unique filename using crypto/rand.
func generateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes) + ext
}

// UploadUseCase implements file upload business logic.
type UploadUseCase struct {
	fileRepo port.FileRepository
	storage  port.Storage
	eventBus port.EventPublisher
	logger   *zap.Logger
}

// NewUploadUseCase creates a new UploadUseCase.
func NewUploadUseCase(
	fileRepo port.FileRepository,
	storage port.Storage,
	eventBus port.EventPublisher,
	logger *zap.Logger,
) *UploadUseCase {
	return &UploadUseCase{
		fileRepo: fileRepo,
		storage:  storage,
		eventBus: eventBus,
		logger:   logger,
	}
}
