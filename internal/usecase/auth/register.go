package auth

import (
	"context"
	"regexp"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Register creates a new user account and returns auth tokens.
func (uc *AuthUseCase) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Validate input
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Check email uniqueness
	existing, err := uc.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		uc.logger.Error("failed to check email uniqueness", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}
	if existing != nil {
		return nil, bizerr.ErrEmailAlreadyExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		uc.logger.Error("failed to hash password", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}

	// Create user
	user := &entity.User{
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         entity.RoleCustomer, // Default role for new registrations
	}
	if err := uc.userRepo.CreateUser(ctx, user); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, bizerr.ErrEmailAlreadyExists
		}
		uc.logger.Error("failed to create user", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}

	// Generate tokens
	return uc.generateAuthResponse(ctx, user)
}

func validateEmail(email string) error {
	if len(email) == 0 || len(email) > 255 {
		return bizerr.NewValidationError("email", "must be between 1 and 255 characters")
	}
	if !emailRegex.MatchString(email) {
		return bizerr.NewValidationError("email", "invalid email format")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return bizerr.NewValidationError("password", "must be at least 8 characters")
	}
	if len(password) > 128 {
		return bizerr.NewValidationError("password", "must be at most 128 characters")
	}
	hasLetter := false
	hasDigit := false
	for _, c := range password {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	if !hasLetter {
		return bizerr.NewValidationError("password", "must contain at least one letter")
	}
	if !hasDigit {
		return bizerr.NewValidationError("password", "must contain at least one digit")
	}
	return nil
}

// generateAuthResponse creates tokens for a user and builds the auth response.
func (uc *AuthUseCase) generateAuthResponse(ctx context.Context, user *entity.User) (*AuthResponse, error) {
	accessToken, err := uc.jwtService.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		uc.logger.Error("failed to generate access token", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}

	refreshToken, tokenID, err := uc.jwtService.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		uc.logger.Error("failed to generate refresh token", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}

	// Store refresh token in Redis
	if err := uc.tokenStore.SaveRefreshToken(ctx, user.ID, tokenID, uc.refreshExpire); err != nil {
		uc.logger.Error("failed to save refresh token", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userToInfo(user),
	}, nil
}
