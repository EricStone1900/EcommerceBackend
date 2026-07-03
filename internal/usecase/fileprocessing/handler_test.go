package fileprocessing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/upload"
)

// mockFileRepo implements port.FileRepository with testify/mock.
type mockFileRepo struct {
	mock.Mock
}

func (m *mockFileRepo) Create(ctx context.Context, file *entity.File) error {
	return m.Called(ctx, file).Error(0)
}

func (m *mockFileRepo) GetByID(ctx context.Context, id uint) (*entity.File, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.File), args.Error(1)
}

func (m *mockFileRepo) UpdateStatus(ctx context.Context, id uint, status entity.FileStatus) error {
	return m.Called(ctx, id, status).Error(0)
}

func (m *mockFileRepo) ListByOwner(ctx context.Context, ownerID uint) ([]*entity.File, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.File), args.Error(1)
}

func newTestLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

func TestHandleFileUploaded_Image(t *testing.T) {
	fileRepo := new(mockFileRepo)
	handler := NewHandler(fileRepo, newTestLogger())

	event := upload.FileUploadedEvent{
		FileID:  1,
		Type:    "image",
		URL:     "/uploads/test.jpg",
		OwnerID: 1,
	}

	fileRepo.On("UpdateStatus", mock.Anything, uint(1), entity.FileStatusProcessed).Return(nil)

	err := handler.HandleFileUploaded(context.Background(), event)
	assert.NoError(t, err)
	fileRepo.AssertExpectations(t)
}

func TestHandleFileUploaded_NonImage(t *testing.T) {
	fileRepo := new(mockFileRepo)
	handler := NewHandler(fileRepo, newTestLogger())

	event := upload.FileUploadedEvent{
		FileID:  2,
		Type:    "document",
		URL:     "/uploads/doc.pdf",
		OwnerID: 1,
	}

	// Should NOT call UpdateStatus for non-image files
	err := handler.HandleFileUploaded(context.Background(), event)
	assert.NoError(t, err)
	fileRepo.AssertNotCalled(t, "UpdateStatus")
}

func TestHandleFileUploaded_RepoError(t *testing.T) {
	fileRepo := new(mockFileRepo)
	handler := NewHandler(fileRepo, newTestLogger())

	event := upload.FileUploadedEvent{
		FileID:  3,
		Type:    "image",
		URL:     "/uploads/error.jpg",
		OwnerID: 1,
	}

	fileRepo.On("UpdateStatus", mock.Anything, uint(3), entity.FileStatusProcessed).
		Return(assert.AnError)

	err := handler.HandleFileUploaded(context.Background(), event)
	assert.Error(t, err)
	fileRepo.AssertExpectations(t)
}

func TestHandleFileUploaded_WrongPayloadType(t *testing.T) {
	fileRepo := new(mockFileRepo)
	handler := NewHandler(fileRepo, newTestLogger())

	// Pass a string instead of FileUploadedEvent
	err := handler.HandleFileUploaded(context.Background(), "not an event")
	assert.NoError(t, err)
	fileRepo.AssertNotCalled(t, "UpdateStatus")
}
