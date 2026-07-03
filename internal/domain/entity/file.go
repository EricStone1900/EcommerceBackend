package entity

import "time"

// FileStatus represents the processing status of an uploaded file.
type FileStatus string

const (
	FileStatusPending   FileStatus = "pending"
	FileStatusProcessed FileStatus = "processed"
)

// FileType represents the category of an uploaded file.
type FileType string

const (
	FileTypeImage    FileType = "image"
	FileTypeDocument FileType = "document"
	FileTypeVideo    FileType = "video"
)

// File represents an uploaded file in the ecommerce system.
// Note: DeletedAt is intentionally omitted — it's infra-only (gorm.DeletedAt).
type File struct {
	ID           uint       `json:"id"`
	OwnerID      uint       `json:"owner_id"`
	Type         string     `json:"type"`
	OriginalName string     `json:"original_name"`
	URL          string     `json:"url"`
	Size         int64      `json:"size"`
	Status       FileStatus `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}
