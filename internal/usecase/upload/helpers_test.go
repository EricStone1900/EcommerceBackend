package upload

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
)

// mockFileRepo implements port.FileRepository with testify/mock.
type mockFileRepo struct {
	mock.Mock
}

func (m *mockFileRepo) Create(ctx context.Context, file *entity.File) error {
	args := m.Called(ctx, file)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	// Simulate DB-assigned fields
	file.ID = 1
	file.CreatedAt = time.Now()
	return nil
}

func (m *mockFileRepo) GetByID(ctx context.Context, id uint) (*entity.File, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.File), args.Error(1)
}

func (m *mockFileRepo) UpdateStatus(ctx context.Context, id uint, status entity.FileStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *mockFileRepo) ListByOwner(ctx context.Context, ownerID uint) ([]*entity.File, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.File), args.Error(1)
}

// mockStorage implements port.Storage with testify/mock.
type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) Upload(ctx context.Context, data []byte, filename string) (string, error) {
	args := m.Called(ctx, data, filename)
	return args.String(0), args.Error(1)
}

func (m *mockStorage) Delete(ctx context.Context, url string) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

// mockEventBus implements port.EventPublisher with testify/mock.
type mockEventBus struct {
	mock.Mock
}

func (m *mockEventBus) Publish(ctx context.Context, eventName string, payload any) error {
	args := m.Called(ctx, eventName, payload)
	return args.Error(0)
}

func (m *mockEventBus) Subscribe(eventName string, handler func(ctx context.Context, payload any) error) {
	m.Called(eventName, handler)
}

// newTestUploadUseCase creates a test UploadUseCase with all mocked dependencies.
func newTestUploadUseCase() (*mockFileRepo, *mockStorage, *mockEventBus, *UploadUseCase) {
	fileRepo := new(mockFileRepo)
	storage := new(mockStorage)
	eventBus := new(mockEventBus)
	logger, _ := zap.NewDevelopment()
	uc := NewUploadUseCase(fileRepo, storage, eventBus, logger)
	return fileRepo, storage, eventBus, uc
}
