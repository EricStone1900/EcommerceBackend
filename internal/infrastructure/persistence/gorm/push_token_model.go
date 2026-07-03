package gorm

import (
	"time"

	"gorm.io/gorm"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
)

// pushTokenModel is the GORM-specific persistence model for push notification tokens.
// It carries gorm.DeletedAt for soft delete — the domain entity does not.
type pushTokenModel struct {
	ID          uint           `gorm:"primaryKey"`
	UserID      uint           `gorm:"not null;uniqueIndex:idx_push_tokens_user_device"`
	DeviceToken string         `gorm:"not null;uniqueIndex:idx_push_tokens_user_device"`
	Platform    string         `gorm:"not null;size:20"`
	CreatedAt   time.Time      `gorm:"not null"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the default table name.
func (pushTokenModel) TableName() string {
	return "push_tokens"
}

// toEntity converts the GORM model to a domain entity.
func (m *pushTokenModel) toEntity() *entity.PushToken {
	return &entity.PushToken{
		ID:          m.ID,
		UserID:      m.UserID,
		DeviceToken: m.DeviceToken,
		Platform:    entity.Platform(m.Platform),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// toPushTokenModel converts a domain entity to the GORM model.
func toPushTokenModel(t *entity.PushToken) *pushTokenModel {
	return &pushTokenModel{
		ID:          t.ID,
		UserID:      t.UserID,
		DeviceToken: t.DeviceToken,
		Platform:    string(t.Platform),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
