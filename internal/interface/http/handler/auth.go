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

// @Summary      用户注册
// @Description  使用邮箱和密码注册新用户
// @Tags         认证管理
// @Accept       json
// @Produce      json
// @Param        request body auth.RegisterRequest true "注册请求"
// @Success      201 {object} response.Response{data=auth.AuthResponse}
// @Failure      400 {object} response.Response
// @Failure      409 {object} response.Response
// @Router       /api/v1/auth/register [post]
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

// @Summary      用户登录
// @Description  使用邮箱和密码登录，返回 JWT Token
// @Tags         认证管理
// @Accept       json
// @Produce      json
// @Param        request body auth.LoginRequest true "登录请求"
// @Success      200 {object} response.Response{data=auth.AuthResponse}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Router       /api/v1/auth/login [post]
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

// @Summary      用户登出
// @Description  使 refresh token 失效，完成登出
// @Tags         认证管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body auth.RefreshRequest true "登出请求"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Router       /api/v1/auth/logout [post]
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

// @Summary      刷新令牌
// @Description  使用 refresh token 获取新的访问令牌
// @Tags         认证管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body auth.RefreshRequest true "刷新请求"
// @Success      200 {object} response.Response{data=auth.AuthResponse}
// @Failure      400 {object} response.Response
// @Failure      401 {object} response.Response
// @Router       /api/v1/auth/refresh [post]
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

// @Summary      获取当前用户
// @Description  返回当前登录用户的信息
// @Tags         认证管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} response.Response{data=auth.UserInfo}
// @Failure      401 {object} response.Response
// @Router       /api/v1/auth/me [get]
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
