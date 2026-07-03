package gorm

import (
	"time"

	"gorm.io/gorm"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
)

// fileModel is the GORM-specific persistence model for uploaded files.
// It carries gorm.DeletedAt for soft delete — the domain entity does not.
type fileModel struct {
	ID           uint           `gorm:"primaryKey"`
	OwnerID      uint           `gorm:"not null;index"`
	Type         string         `gorm:"not null;size:20"`
	OriginalName string         `gorm:"not null;size:255"`
	URL          string         `gorm:"not null;size:512"`
	Size         int64          `gorm:"not null;default:0"`
	Status       string         `gorm:"not null;size:20;default:pending"`
	CreatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the default table name.
func (fileModel) TableName() string {
	return "files"
}

// toEntity converts the GORM model to a domain entity.
func (m *fileModel) toEntity() *entity.File {
	return &entity.File{
		ID:           m.ID,
		OwnerID:      m.OwnerID,
		Type:         m.Type,
		OriginalName: m.OriginalName,
		URL:          m.URL,
		Size:         m.Size,
		Status:       entity.FileStatus(m.Status),
		CreatedAt:    m.CreatedAt,
	}
}

// toFileModel converts a domain entity to the GORM model.
func toFileModel(f *entity.File) *fileModel {
	return &fileModel{
		ID:           f.ID,
		OwnerID:      f.OwnerID,
		Type:         f.Type,
		OriginalName: f.OriginalName,
		URL:          f.URL,
		Size:         f.Size,
		Status:       string(f.Status),
		CreatedAt:    f.CreatedAt,
	}
}
