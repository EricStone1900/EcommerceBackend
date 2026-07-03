package port

import (
	"context"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
)

// FileRepository defines the interface for file metadata persistence.
type FileRepository interface {
	// Create saves a new file record.
	Create(ctx context.Context, file *entity.File) error

	// GetByID retrieves a file by its ID.
	GetByID(ctx context.Context, id uint) (*entity.File, error)

	// UpdateStatus updates the processing status of a file.
	UpdateStatus(ctx context.Context, id uint, status entity.FileStatus) error

	// ListByOwner retrieves all files uploaded by a specific user.
	ListByOwner(ctx context.Context, ownerID uint) ([]*entity.File, error)
}
