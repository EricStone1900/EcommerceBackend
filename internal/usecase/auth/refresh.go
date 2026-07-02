package auth

import (
	"context"

	"go.uber.org/zap"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// Refresh validates a refresh token and issues a new token pair (token rotation).
func (uc *AuthUseCase) Refresh(ctx context.Context, req RefreshRequest) (*AuthResponse, error) {
	// Validate the refresh token
	claims, err := uc.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Ensure this is a refresh token, not an access token
	if !uc.jwtService.IsRefreshToken(claims) {
		return nil, bizerr.ErrInvalidRequest
	}

	// Check that the token exists in Redis (hasn't been revoked)
	valid, err := uc.tokenStore.ValidateRefreshToken(ctx, claims.UserID, claims.ID)
	if err != nil {
		uc.logger.Error("failed to validate refresh token in store", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}
	if !valid {
		return nil, bizerr.ErrUnauthorized
	}

	// Delete the old refresh token (rotation)
	if err := uc.tokenStore.DeleteRefreshToken(ctx, claims.UserID, claims.ID); err != nil {
		uc.logger.Error("failed to delete old refresh token", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}

	// Fetch the user to confirm they still exist
	user, err := uc.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		uc.logger.Error("failed to get user by id", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}
	if user == nil {
		return nil, bizerr.ErrUserNotFound
	}

	// Generate new token pair
	return uc.generateAuthResponse(ctx, user)
}
