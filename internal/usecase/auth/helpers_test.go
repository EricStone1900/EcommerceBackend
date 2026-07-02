package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/pkg/jwt"
)

// --- Mocks ---

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *entity.User) error {
	args := m.Called(ctx, user)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	// Simulate ID and timestamp population by DB
	user.ID = 1
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

type mockTokenStore struct {
	mock.Mock
}

func (m *mockTokenStore) SaveRefreshToken(ctx context.Context, userID uint, tokenID string, expiration time.Duration) error {
	args := m.Called(ctx, userID, tokenID, expiration)
	return args.Error(0)
}

func (m *mockTokenStore) ValidateRefreshToken(ctx context.Context, userID uint, tokenID string) (bool, error) {
	args := m.Called(ctx, userID, tokenID)
	return args.Bool(0), args.Error(1)
}

func (m *mockTokenStore) DeleteRefreshToken(ctx context.Context, userID uint, tokenID string) error {
	args := m.Called(ctx, userID, tokenID)
	return args.Error(0)
}

// --- Test helpers ---

const testPassword = "Password1"

func newTestAuthUseCase(t *testing.T) (*mockUserRepo, *mockTokenStore, *AuthUseCase) {
	t.Helper()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	userRepo := new(mockUserRepo)
	tokenStore := new(mockTokenStore)

	jwtService := newTestJWTService()
	uc := NewAuthUseCase(
		userRepo,
		tokenStore,
		jwtService,
		logger,
		15*time.Minute,
		168*time.Hour,
	)

	return userRepo, tokenStore, uc
}

func newTestJWTService() *jwt.JWTService {
	return jwt.NewJWTService("test_secret_key_for_ut", 15*time.Minute, 168*time.Hour)
}

func newTestUser() *entity.User {
	hash, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return &entity.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: string(hash),
		Role:         entity.RoleCustomer,
	}
}
