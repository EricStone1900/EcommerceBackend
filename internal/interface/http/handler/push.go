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

// @Summary      注册推送 Token
// @Description  注册设备推送通知 Token
// @Tags         推送管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body push.RegisterTokenRequest true "Token 注册请求"
// @Success      200 {object} response.Response{data=push.RegisterTokenResponse}
// @Failure      400 {object} response.Response
// @Router       /api/v1/push/token [post]
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

// @Summary      删除推送 Token
// @Description  删除设备推送通知 Token
// @Tags         推送管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body push.DeleteTokenRequest true "Token 删除请求"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Router       /api/v1/push/token [delete]
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

// @Summary      发送测试推送
// @Description  向当前用户发送一条测试推送通知
// @Tags         推送管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} response.Response{data=push.SendTestResponse}
// @Router       /api/v1/push/test [post]
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
