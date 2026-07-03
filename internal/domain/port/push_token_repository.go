package port

import (
	"context"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
)

// PushTokenRepository defines the interface for device token persistence.
type PushTokenRepository interface {
	// Create saves a new push token or updates existing (idempotent by user_id + device_token).
	Create(ctx context.Context, token *entity.PushToken) error

	// GetByUserID retrieves all push tokens registered by a user.
	GetByUserID(ctx context.Context, userID uint) ([]*entity.PushToken, error)

	// DeleteByUserAndDevice soft-deletes a specific device token for a user.
	// Returns nil if the token doesn't exist (idempotent).
	DeleteByUserAndDevice(ctx context.Context, userID uint, deviceToken string) error
}
