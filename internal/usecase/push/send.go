package push

import (
	"context"

	"go.uber.org/zap"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// SendTest sends a test push notification to all devices registered by the user.
// Returns the number of notifications sent.
func (uc *PushUseCase) SendTest(ctx context.Context, userID uint) (*SendTestResponse, error) {
	tokens, err := uc.pushTokenRepo.GetByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error("failed to get push tokens", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}

	if len(tokens) == 0 {
		return nil, bizerr.ErrPushTokenNotFound
	}

	sent := 0
	for _, token := range tokens {
		if err := uc.notifier.SendPush(ctx, userID, token.DeviceToken, "Test Notification", "This is a test push from the ecommerce backend"); err != nil {
			uc.logger.Error("failed to send test push",
				zap.String("device_token", token.DeviceToken),
				zap.Error(err),
			)
			return nil, bizerr.ErrPushSendFailed
		}
		sent++
	}

	return &SendTestResponse{Sent: sent}, nil
}
