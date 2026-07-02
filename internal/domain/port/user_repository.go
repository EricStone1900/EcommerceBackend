package port

import (
	"context"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
)

// UserRepository defines the persistence contract for User entities.
// Implementation lives in the infrastructure layer (GORM).
type UserRepository interface {
	// CreateUser persists a new user. Returns the user with ID and timestamps populated.
	CreateUser(ctx context.Context, user *entity.User) error

	// GetUserByEmail looks up a user by email. Returns nil and no error if not found.
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)

	// GetUserByID looks up a user by primary key.
	GetUserByID(ctx context.Context, id uint) (*entity.User, error)
}
