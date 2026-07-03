package port

import "context"

// Notifier defines the interface for push notification delivery.
// Implementations can be stub (local) or real APNs/FCM (production).
// Method signatures include context.Context and return error to reflect
// potential network failures.
//
// FUTURE: Replace the stub with APNs (iOS) or FCM (Android) implementations.
// The port.Notifier interface does NOT need to change.
type Notifier interface {
	// SendPush sends a push notification to the specified device.
	SendPush(ctx context.Context, userID uint, deviceToken, title, body string) error
}
