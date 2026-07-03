package push

import (
	"context"
	"strings"

	"go.uber.org/zap"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// DeleteToken removes a device push token for the authenticated user.
// Idempotent: deleting a non-existent token succeeds silently.
func (uc *PushUseCase) DeleteToken(ctx context.Context, userID uint, req DeleteTokenRequest) error {
	deviceToken := strings.TrimSpace(req.DeviceToken)
	if deviceToken == "" {
		return bizerr.NewValidationError("device_token", "cannot be empty")
	}

	if err := uc.pushTokenRepo.DeleteByUserAndDevice(ctx, userID, deviceToken); err != nil {
		uc.logger.Error("failed to delete push token", zap.Error(err))
		return bizerr.ErrInternalError
	}

	return nil
}
