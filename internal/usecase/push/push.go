package push

import (
	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/internal/domain/port"
)

// --- DTOs ---

// RegisterTokenRequest is the input for registering a device push token.
type RegisterTokenRequest struct {
	DeviceToken string `json:"device_token"`
	Platform    string `json:"platform"`
}

// RegisterTokenResponse is returned after successful token registration.
type RegisterTokenResponse struct {
	ID        uint   `json:"id"`
	Platform  string `json:"platform"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// DeleteTokenRequest is the input for deleting a device push token.
type DeleteTokenRequest struct {
	DeviceToken string `json:"device_token"`
}

// SendTestResponse is returned after sending a test push notification.
type SendTestResponse struct {
	Sent int `json:"sent"`
}

// pushTokenToResponse converts a domain entity to the response DTO.
func pushTokenToResponse(t *entity.PushToken) *RegisterTokenResponse {
	if t == nil {
		return nil
	}
	return &RegisterTokenResponse{
		ID:        t.ID,
		Platform:  string(t.Platform),
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: t.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// PushUseCase implements push notification business logic.
type PushUseCase struct {
	pushTokenRepo port.PushTokenRepository
	notifier      port.Notifier
	logger        *zap.Logger
}

// NewPushUseCase creates a new PushUseCase.
func NewPushUseCase(
	pushTokenRepo port.PushTokenRepository,
	notifier port.Notifier,
	logger *zap.Logger,
) *PushUseCase {
	return &PushUseCase{
		pushTokenRepo: pushTokenRepo,
		notifier:      notifier,
		logger:        logger,
	}
}
