package auth

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// Login authenticates a user and returns auth tokens.
func (uc *AuthUseCase) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// Look up user by email
	user, err := uc.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		uc.logger.Error("failed to get user by email", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}
	if user == nil {
		// Return same error for both wrong email and wrong password to prevent email enumeration
		return nil, bizerr.ErrInvalidCredentials
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, bizerr.ErrInvalidCredentials
	}

	// Generate tokens
	return uc.generateAuthResponse(ctx, user)
}
