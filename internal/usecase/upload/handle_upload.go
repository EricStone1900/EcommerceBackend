package upload

import (
	"context"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// ContainsExt checks if the extension is in the allowed list.
func containsExt(exts []string, ext string) bool {
	for _, e := range exts {
		if e == ext {
			return true
		}
	}
	return false
}

// HandleUpload processes a file upload: validates, stores, persists metadata, and publishes an event.
func (uc *UploadUseCase) HandleUpload(ctx context.Context, ownerID uint, fileBytes []byte, filename string, fileTypeStr string) (*UploadResponse, error) {
	// 1. Validate file type
	fileType := entity.FileType(fileTypeStr)
	config, ok := fileTypeConfig[fileType]
	if !ok {
		return nil, bizerr.ErrInvalidFileType
	}

	// 2. Validate extension
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" || !containsExt(config.Extensions, ext) {
		return nil, bizerr.ErrInvalidFileType
	}

	// 3. Validate size
	if int64(len(fileBytes)) > config.MaxSize {
		return nil, bizerr.ErrFileTooLarge
	}

	// 4. Generate unique filename
	uniqueName := generateUniqueFilename(filename)

	// 5. Upload to storage
	url, err := uc.storage.Upload(ctx, fileBytes, uniqueName)
	if err != nil {
		uc.logger.Error("storage upload failed", zap.Error(err))
		return nil, bizerr.ErrFileUploadFailed
	}

	// 6. Save file metadata
	file := &entity.File{
		OwnerID:      ownerID,
		Type:         string(fileType),
		OriginalName: filename,
		URL:          url,
		Size:         int64(len(fileBytes)),
		Status:       entity.FileStatusPending,
	}
	if err := uc.fileRepo.Create(ctx, file); err != nil {
		uc.logger.Error("file repo create failed", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}

	// 7. Publish event (async, log error but don't fail the request)
	eventPayload := FileUploadedEvent{
		FileID:  file.ID,
		Type:    string(fileType),
		URL:     url,
		OwnerID: ownerID,
	}
	if err := uc.eventBus.Publish(ctx, "file.uploaded", eventPayload); err != nil {
		uc.logger.Warn("failed to publish file.uploaded event", zap.Error(err))
	}

	return fileToResponse(file), nil
}
