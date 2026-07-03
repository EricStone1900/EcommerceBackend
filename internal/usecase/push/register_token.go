package push

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// RegisterToken registers a device push token for the authenticated user.
// Idempotent: registering the same token again updates the timestamp.
func (uc *PushUseCase) RegisterToken(ctx context.Context, userID uint, req RegisterTokenRequest) (*RegisterTokenResponse, error) {
	// Validate device token
	deviceToken := strings.TrimSpace(req.DeviceToken)
	if deviceToken == "" {
		return nil, bizerr.NewValidationError("device_token", "cannot be empty")
	}

	// Validate platform
	if req.Platform != string(entity.PlatformIOS) {
		return nil, bizerr.ErrInvalidPlatform
	}

	token := &entity.PushToken{
		UserID:      userID,
		DeviceToken: deviceToken,
		Platform:    entity.Platform(req.Platform),
	}

	if err := uc.pushTokenRepo.Create(ctx, token); err != nil {
		uc.logger.Error("failed to create push token", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}

	return pushTokenToResponse(token), nil
}
