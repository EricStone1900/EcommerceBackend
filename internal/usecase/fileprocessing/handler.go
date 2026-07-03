package fileprocessing

import (
	"context"
	"reflect"

	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/internal/domain/port"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/upload"
)

// Handler processes file.uploaded events.
// In the future, the image handler here would call an actual OCR microservice.
type Handler struct {
	fileRepo port.FileRepository
	logger   *zap.Logger
}

// NewHandler creates a new file processing event handler.
func NewHandler(fileRepo port.FileRepository, logger *zap.Logger) *Handler {
	return &Handler{
		fileRepo: fileRepo,
		logger:   logger,
	}
}

// HandleFileUploaded handles the "file.uploaded" event.
// For image files, it simulates OCR processing by logging and updating the file status.
func (h *Handler) HandleFileUploaded(ctx context.Context, payload any) error {
	event, ok := payload.(upload.FileUploadedEvent)
	if !ok {
		h.logger.Warn("fileprocessing: unexpected payload type",
			zap.String("type", reflect.TypeOf(payload).String()),
		)
		return nil
	}

	h.logger.Info("file.uploaded event received",
		zap.Uint("file_id", event.FileID),
		zap.String("type", event.Type),
		zap.String("url", event.URL),
		zap.Uint("owner_id", event.OwnerID),
	)

	// Only process images (future OCR use case)
	if event.Type != string(entity.FileTypeImage) {
		h.logger.Debug("fileprocessing: skipping non-image file",
			zap.String("type", event.Type),
		)
		return nil
	}

	h.logger.Info("fileprocessing: simulating OCR processing for image",
		zap.Uint("file_id", event.FileID),
		zap.String("note", "此处未来会调用 OCR 微服务"),
	)

	// Update file status to processed
	if err := h.fileRepo.UpdateStatus(ctx, event.FileID, entity.FileStatusProcessed); err != nil {
		h.logger.Error("fileprocessing: failed to update file status", zap.Error(err))
		return err
	}

	h.logger.Info("fileprocessing: file status updated to processed",
		zap.Uint("file_id", event.FileID),
	)
	return nil
}
