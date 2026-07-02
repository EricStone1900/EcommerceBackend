package handler

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/middleware"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/auth"
	"github.com/EricStone1900/ecommerce-backend/pkg/errors"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
)

// AuthUseCase defines the interface for authentication operations.
// This abstraction enables testability via mock implementations.
type AuthUseCase interface {
	Register(ctx context.Context, req auth.RegisterRequest) (*auth.AuthResponse, error)
	Login(ctx context.Context, req auth.LoginRequest) (*auth.AuthResponse, error)
	Logout(ctx context.Context, userID uint, refreshToken string) error
	Refresh(ctx context.Context, req auth.RefreshRequest) (*auth.AuthResponse, error)
	GetCurrentUser(ctx context.Context, userID uint) (*auth.UserInfo, error)
}

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	authUsecase AuthUseCase
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authUsecase AuthUseCase) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase}
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(errors.CodeInvalidRequest, "invalid request body"))
		return
	}

	resp, err := h.authUsecase.Register(c.Request.Context(), req)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(201, response.Success(resp))
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(errors.CodeInvalidRequest, "invalid request body"))
		return
	}

	resp, err := h.authUsecase.Login(c.Request.Context(), req)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(200, response.Success(resp))
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req auth.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(errors.CodeInvalidRequest, "invalid request body"))
		return
	}

	if err := h.authUsecase.Logout(c.Request.Context(), userID, req.RefreshToken); err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(200, response.Success(nil))
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req auth.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(errors.CodeInvalidRequest, "invalid request body"))
		return
	}

	resp, err := h.authUsecase.Refresh(c.Request.Context(), req)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(200, response.Success(resp))
}

// Me handles GET /api/v1/auth/me — returns the current authenticated user's info.
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)

	resp, err := h.authUsecase.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		handleAuthError(c, err)
		return
	}

	c.JSON(200, response.Success(resp))
}

// handleAuthError converts a use case error into the appropriate HTTP response.
func handleAuthError(c *gin.Context, err error) {
	if bizErr, ok := err.(*errors.Error); ok {
		c.JSON(bizErr.HTTPCode, response.Error(bizErr.Code, bizErr.Message))
		return
	}
	c.JSON(500, response.Error(errors.CodeInternalError, "internal server error"))
}
