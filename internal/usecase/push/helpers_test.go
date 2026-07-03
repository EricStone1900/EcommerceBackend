package push

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
)

// mockPushTokenRepo implements port.PushTokenRepository with testify/mock.
type mockPushTokenRepo struct {
	mock.Mock
}

func (m *mockPushTokenRepo) Create(ctx context.Context, token *entity.PushToken) error {
	args := m.Called(ctx, token)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	// Simulate DB-assigned fields
	token.ID = 1
	token.CreatedAt = time.Now()
	token.UpdatedAt = time.Now()
	return nil
}

func (m *mockPushTokenRepo) GetByUserID(ctx context.Context, userID uint) ([]*entity.PushToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.PushToken), args.Error(1)
}

func (m *mockPushTokenRepo) DeleteByUserAndDevice(ctx context.Context, userID uint, deviceToken string) error {
	args := m.Called(ctx, userID, deviceToken)
	return args.Error(0)
}

// mockNotifier implements port.Notifier with testify/mock.
type mockNotifier struct {
	mock.Mock
}

func (m *mockNotifier) SendPush(ctx context.Context, userID uint, deviceToken, title, body string) error {
	args := m.Called(ctx, userID, deviceToken, title, body)
	return args.Error(0)
}

// newTestPushUseCase creates a test PushUseCase with mocked dependencies.
func newTestPushUseCase() (*mockPushTokenRepo, *mockNotifier, *PushUseCase) {
	repo := new(mockPushTokenRepo)
	notifier := new(mockNotifier)
	logger, _ := zap.NewDevelopment()
	uc := NewPushUseCase(repo, notifier, logger)
	return repo, notifier, uc
}

// newTestPushToken creates a sample push token entity for testing.
func newTestPushToken() *entity.PushToken {
	return &entity.PushToken{
		ID:          1,
		UserID:      1,
		DeviceToken: "device-token-abc-123",
		Platform:    entity.PlatformIOS,
	}
}

// newTestPushTokens creates a list of sample push tokens.
func newTestPushTokens(count int) []*entity.PushToken {
	tokens := make([]*entity.PushToken, count)
	for i := 0; i < count; i++ {
		t := newTestPushToken()
		t.ID = uint(i + 1)
		t.DeviceToken = "device-token-" + string(rune('a'+i))
		tokens[i] = t
	}
	return tokens
}
