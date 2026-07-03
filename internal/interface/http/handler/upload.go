package handler

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"

	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/middleware"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/upload"
	"github.com/EricStone1900/ecommerce-backend/pkg/errors"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
)

// UploadUseCase defines the interface for file upload operations.
// This abstraction enables testability via mock implementations.
type UploadUseCase interface {
	HandleUpload(ctx context.Context, ownerID uint, fileBytes []byte, filename string, fileType string) (*upload.UploadResponse, error)
}

// UploadHandler handles HTTP requests for file upload endpoints.
type UploadHandler struct {
	uploadUseCase UploadUseCase
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(uploadUseCase UploadUseCase) *UploadHandler {
	return &UploadHandler{uploadUseCase: uploadUseCase}
}

// Upload handles POST /api/v1/upload (protected, any role)
// Accepts multipart/form-data with "file" (the file) and "type" (image|document|video).
func (h *UploadHandler) Upload(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// Parse file type
	fileType := c.PostForm("type")
	if fileType == "" {
		c.JSON(400, response.Error(errors.CodeInvalidFileType, "file type is required"))
		return
	}

	// Parse uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(400, response.Error(errors.CodeInvalidRequest, "file is required"))
		return
	}
	defer file.Close()

	// Read file contents
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(500, response.Error(errors.CodeInternalError, "failed to read file"))
		return
	}

	// Process upload
	resp, err := h.uploadUseCase.HandleUpload(c.Request.Context(), userID, fileBytes, header.Filename, fileType)
	if err != nil {
		handleUploadError(c, err)
		return
	}

	c.JSON(201, response.Success(resp))
}

// handleUploadError converts a use case error into the appropriate HTTP response.
func handleUploadError(c *gin.Context, err error) {
	if bizErr, ok := err.(*errors.Error); ok {
		c.JSON(bizErr.HTTPCode, response.Error(bizErr.Code, bizErr.Message))
		return
	}
	c.JSON(500, response.Error(errors.CodeInternalError, "internal server error"))
}
