package pushstub

import (
	"context"

	"go.uber.org/zap"
)

// Notifier implements port.Notifier by logging push notifications.
//
// FUTURE: Replace this with an APNs implementation.
// See internal/infrastructure/push/apns/doc.go for setup instructions.
type Notifier struct {
	logger *zap.Logger
}

// NewNotifier creates a new stub Notifier.
func NewNotifier(logger *zap.Logger) *Notifier {
	return &Notifier{logger: logger}
}

// SendPush logs a push notification instead of sending it to a real device.
// This validates the notification pipeline end-to-end before APNs integration.
func (n *Notifier) SendPush(ctx context.Context, userID uint, deviceToken, title, body string) error {
	n.logger.Info("stub push notification",
		zap.Uint("user_id", userID),
		zap.String("device_token", deviceToken),
		zap.String("title", title),
		zap.String("body", body),
	)
	return nil
}
