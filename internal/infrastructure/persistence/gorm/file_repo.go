package gorm

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
)

// FileRepository implements port.FileRepository using GORM.
type FileRepository struct {
	db *gorm.DB
}

// NewFileRepository creates a new GORM-backed FileRepository.
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create persists a new file record and populates the ID and timestamps.
func (r *FileRepository) Create(ctx context.Context, file *entity.File) error {
	model := toFileModel(file)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	// Copy generated fields back
	*file = *model.toEntity()
	return nil
}

// GetByID retrieves a file by its primary key. Returns nil if not found.
func (r *FileRepository) GetByID(ctx context.Context, id uint) (*entity.File, error) {
	var model fileModel
	err := r.db.WithContext(ctx).First(&model, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file by id: %w", err)
	}
	return model.toEntity(), nil
}

// UpdateStatus updates the processing status of a file.
func (r *FileRepository) UpdateStatus(ctx context.Context, id uint, status entity.FileStatus) error {
	result := r.db.WithContext(ctx).Model(&fileModel{}).
		Where("id = ?", id).
		Update("status", string(status))
	if result.Error != nil {
		return fmt.Errorf("failed to update file status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListByOwner retrieves all files uploaded by a specific user.
func (r *FileRepository) ListByOwner(ctx context.Context, ownerID uint) ([]*entity.File, error) {
	var models []fileModel
	if err := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list files by owner: %w", err)
	}

	files := make([]*entity.File, len(models))
	for i := range models {
		files[i] = models[i].toEntity()
	}
	return files, nil
}
