package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
)

// userModel is the GORM-specific persistence model for users.
// It carries gorm.DeletedAt for soft delete — the domain entity does not.
type userModel struct {
	ID           uint           `gorm:"primaryKey"`
	Email        string         `gorm:"uniqueIndex;not null;size:255"`
	PasswordHash string         `gorm:"not null;size:255"`
	Role         entity.Role    `gorm:"not null;default:customer"`
	CreatedAt    time.Time      `gorm:"not null"`
	UpdatedAt    time.Time      `gorm:"not null"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the default table name.
func (userModel) TableName() string {
	return "users"
}

// toEntity converts the GORM model to a domain entity.
func (m *userModel) toEntity() *entity.User {
	return &entity.User{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Role:         m.Role,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// toModel converts a domain entity to the GORM model.
func toModel(u *entity.User) *userModel {
	return &userModel{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Role:         u.Role,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// UserRepository implements port.UserRepository using GORM.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new GORM-backed UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser persists a new user and populates the ID and timestamps.
func (r *UserRepository) CreateUser(ctx context.Context, user *entity.User) error {
	model := toModel(user)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("email %s already exists: %w", user.Email, err)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	// Copy generated fields back
	*user = *model.toEntity()
	return nil
}

// GetUserByEmail looks up a user by email. Returns nil if not found.
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model userModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return model.toEntity(), nil
}

// GetUserByID looks up a user by primary key.
func (r *UserRepository) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	var model userModel
	err := r.db.WithContext(ctx).First(&model, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return model.toEntity(), nil
}

// isDuplicateKeyError checks if the error is a PostgreSQL unique constraint violation.
func isDuplicateKeyError(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey)
}
