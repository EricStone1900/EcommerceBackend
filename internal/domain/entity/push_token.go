package entity

import "time"

// Platform represents the mobile platform of a device.
type Platform string

const (
	PlatformIOS     Platform = "ios"
	PlatformAndroid Platform = "android" // reserved for future use
)

// PushToken represents a device push notification token registered by a user.
// Note: DeletedAt is intentionally omitted — it's infra-only (gorm.DeletedAt).
type PushToken struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	DeviceToken string    `json:"device_token"`
	Platform    Platform  `json:"platform"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
