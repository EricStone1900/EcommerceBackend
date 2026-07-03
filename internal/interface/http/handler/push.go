package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/middleware"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/push"
	"github.com/EricStone1900/ecommerce-backend/pkg/errors"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
)

// PushUseCase defines the interface for push notification operations.
// This abstraction enables testability via mock implementations.
type PushUseCase interface {
	RegisterToken(ctx context.Context, userID uint, req push.RegisterTokenRequest) (*push.RegisterTokenResponse, error)
	DeleteToken(ctx context.Context, userID uint, req push.DeleteTokenRequest) error
	SendTest(ctx context.Context, userID uint) (*push.SendTestResponse, error)
}

// PushHandler handles HTTP requests for push notification endpoints.
type PushHandler struct {
	pushUseCase PushUseCase
}

// NewPushHandler creates a new PushHandler.
func NewPushHandler(pushUseCase PushUseCase) *PushHandler {
	return &PushHandler{pushUseCase: pushUseCase}
}

// RegisterToken handles POST /api/v1/push/token (protected, any role)
func (h *PushHandler) RegisterToken(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req push.RegisterTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(errors.CodeInvalidRequest, "invalid request body"))
		return
	}

	resp, err := h.pushUseCase.RegisterToken(c.Request.Context(), userID, req)
	if err != nil {
		handlePushError(c, err)
		return
	}

	c.JSON(200, response.Success(resp))
}

// DeleteToken handles DELETE /api/v1/push/token (protected, any role)
func (h *PushHandler) DeleteToken(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req push.DeleteTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(errors.CodeInvalidRequest, "invalid request body"))
		return
	}

	if err := h.pushUseCase.DeleteToken(c.Request.Context(), userID, req); err != nil {
		handlePushError(c, err)
		return
	}

	c.JSON(200, response.Success(nil))
}

// SendTest handles POST /api/v1/push/test (protected, any role)
func (h *PushHandler) SendTest(c *gin.Context) {
	userID := middleware.GetUserID(c)

	resp, err := h.pushUseCase.SendTest(c.Request.Context(), userID)
	if err != nil {
		handlePushError(c, err)
		return
	}

	c.JSON(200, response.Success(resp))
}

// handlePushError converts a use case error into the appropriate HTTP response.
func handlePushError(c *gin.Context, err error) {
	if bizErr, ok := err.(*errors.Error); ok {
		c.JSON(bizErr.HTTPCode, response.Error(bizErr.Code, bizErr.Message))
		return
	}
	c.JSON(500, response.Error(errors.CodeInternalError, "internal server error"))
}
