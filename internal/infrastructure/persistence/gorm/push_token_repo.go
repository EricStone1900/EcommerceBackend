package gorm

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
)

// PushTokenRepository implements port.PushTokenRepository using GORM.
type PushTokenRepository struct {
	db *gorm.DB
}

// NewPushTokenRepository creates a new GORM-backed PushTokenRepository.
func NewPushTokenRepository(db *gorm.DB) *PushTokenRepository {
	return &PushTokenRepository{db: db}
}

// Create upserts a push token using ON CONFLICT DO UPDATE for idempotency.
// If the same user_id + device_token already exists, it updates the timestamp.
func (r *PushTokenRepository) Create(ctx context.Context, token *entity.PushToken) error {
	model := toPushTokenModel(token)

	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "device_token"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"updated_at": time.Now()}),
		}).
		Create(model).Error
	if err != nil {
		return fmt.Errorf("failed to upsert push token: %w", err)
	}

	// Re-fetch to get ID and timestamps (upsert doesn't auto-populate on conflict)
	var fetched pushTokenModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND device_token = ?", model.UserID, model.DeviceToken).
		First(&fetched).Error; err != nil {
		return fmt.Errorf("failed to fetch push token after upsert: %w", err)
	}
	*token = *fetched.toEntity()
	return nil
}

// GetByUserID retrieves all active push tokens for a user.
func (r *PushTokenRepository) GetByUserID(ctx context.Context, userID uint) ([]*entity.PushToken, error) {
	var models []pushTokenModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get push tokens by user id: %w", err)
	}

	tokens := make([]*entity.PushToken, len(models))
	for i := range models {
		tokens[i] = models[i].toEntity()
	}
	return tokens, nil
}

// DeleteByUserAndDevice soft-deletes a specific device token for a user.
// Returns nil if the token doesn't exist (idempotent).
func (r *PushTokenRepository) DeleteByUserAndDevice(ctx context.Context, userID uint, deviceToken string) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND device_token = ?", userID, deviceToken).
		Delete(&pushTokenModel{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete push token: %w", result.Error)
	}
	return nil
}
