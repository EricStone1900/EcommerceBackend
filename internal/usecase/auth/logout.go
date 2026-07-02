package auth

import (
	"context"

	"go.uber.org/zap"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// Logout revokes a refresh token, making it unusable for future refresh requests.
// This operation is idempotent — calling it again with the same (or already revoked) token succeeds.
func (uc *AuthUseCase) Logout(ctx context.Context, userID uint, refreshToken string) error {
	// Validate the refresh token
	claims, err := uc.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return err
	}

	// Ensure this is a refresh token, not an access token
	if !uc.jwtService.IsRefreshToken(claims) {
		return bizerr.ErrInvalidRequest
	}

	// Verify the token belongs to the requesting user
	if claims.UserID != userID {
		return bizerr.ErrForbidden
	}

	// Delete the token from Redis (idempotent)
	if err := uc.tokenStore.DeleteRefreshToken(ctx, userID, claims.ID); err != nil {
		uc.logger.Error("failed to delete refresh token", zap.Error(err))
		return bizerr.ErrInternalError
	}

	return nil
}
