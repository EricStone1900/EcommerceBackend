package auth

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/internal/domain/port"
	"github.com/EricStone1900/ecommerce-backend/pkg/jwt"
	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// --- Request DTOs ---

// RegisterRequest is the input for the register use case.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is the input for the login use case.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest is the input for the refresh token use case.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// --- Response DTOs ---

// AuthResponse is the output for login and register use cases.
type AuthResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         UserInfo `json:"user"`
}

// UserInfo is a public representation of a user.
type UserInfo struct {
	ID    uint        `json:"id"`
	Email string      `json:"email"`
	Role  entity.Role `json:"role"`
}

// AuthUseCase orchestrates user authentication operations.
type AuthUseCase struct {
	userRepo      port.UserRepository
	tokenStore    port.TokenStore
	jwtService    *jwt.JWTService
	logger        *zap.Logger
	accessExpire  time.Duration
	refreshExpire time.Duration
}

// NewAuthUseCase creates a new AuthUseCase with the given dependencies.
func NewAuthUseCase(
	userRepo port.UserRepository,
	tokenStore port.TokenStore,
	jwtService *jwt.JWTService,
	logger *zap.Logger,
	accessExpire time.Duration,
	refreshExpire time.Duration,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:      userRepo,
		tokenStore:    tokenStore,
		jwtService:    jwtService,
		logger:        logger,
		accessExpire:  accessExpire,
		refreshExpire: refreshExpire,
	}
}

// userToInfo converts a domain User to the public UserInfo DTO.
func userToInfo(u *entity.User) UserInfo {
	return UserInfo{
		ID:    u.ID,
		Email: u.Email,
		Role:  u.Role,
	}
}

// GetCurrentUser retrieves the current user's public info by ID.
func (uc *AuthUseCase) GetCurrentUser(ctx context.Context, userID uint) (*UserInfo, error) {
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		uc.logger.Error("failed to get current user", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}
	if user == nil {
		return nil, bizerr.ErrUserNotFound
	}
	info := userToInfo(user)
	return &info, nil
}
