package port

import (
	"context"
	"time"
)

// TokenStore defines the contract for refresh token persistence.
// Implementation lives in the infrastructure layer (Redis).
type TokenStore interface {
	// SaveRefreshToken stores a refresh token ID with the given TTL.
	SaveRefreshToken(ctx context.Context, userID uint, tokenID string, expiration time.Duration) error

	// ValidateRefreshToken checks whether a refresh token ID exists and is still valid.
	ValidateRefreshToken(ctx context.Context, userID uint, tokenID string) (bool, error)

	// DeleteRefreshToken removes a refresh token ID from the store (revocation).
	DeleteRefreshToken(ctx context.Context, userID uint, tokenID string) error
}
