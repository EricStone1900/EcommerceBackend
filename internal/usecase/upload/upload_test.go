package upload

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

func TestHandleUpload_Success(t *testing.T) {
	fileRepo, storage, eventBus, uc := newTestUploadUseCase()

	fileBytes := []byte("fake-image-data")
	storage.On("Upload", mock.Anything, mock.Anything, mock.AnythingOfType("string")).
		Return("/uploads/testfile.jpg", nil)
	fileRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	eventBus.On("Publish", mock.Anything, "file.uploaded", mock.Anything).Return(nil)

	resp, err := uc.HandleUpload(context.Background(), 1, fileBytes, "photo.jpg", "image")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, "image", resp.Type)
	assert.Equal(t, "photo.jpg", resp.OriginalName)
	assert.Equal(t, int64(15), resp.Size)
	assert.Equal(t, "pending", resp.Status)
	assert.NotEmpty(t, resp.URL)
	assert.NotEmpty(t, resp.CreatedAt)
	storage.AssertExpectations(t)
	fileRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestHandleUpload_InvalidFileType(t *testing.T) {
	_, _, _, uc := newTestUploadUseCase()

	resp, err := uc.HandleUpload(context.Background(), 1, []byte("data"), "file.xyz", "unknown")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeInvalidFileType, err.(*bizerr.Error).Code)
}

func TestHandleUpload_InvalidExtension(t *testing.T) {
	_, _, _, uc := newTestUploadUseCase()

	resp, err := uc.HandleUpload(context.Background(), 1, []byte("data"), "file.exe", "image")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeInvalidFileType, err.(*bizerr.Error).Code)
}

func TestHandleUpload_NoExtension(t *testing.T) {
	_, _, _, uc := newTestUploadUseCase()

	resp, err := uc.HandleUpload(context.Background(), 1, []byte("data"), "noext", "image")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeInvalidFileType, err.(*bizerr.Error).Code)
}

func TestHandleUpload_FileTooLarge(t *testing.T) {
	_, _, _, uc := newTestUploadUseCase()

	// 11MB > 10MB limit for images
	largeData := make([]byte, 11*1024*1024)
	resp, err := uc.HandleUpload(context.Background(), 1, largeData, "large.jpg", "image")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeFileTooLarge, err.(*bizerr.Error).Code)
}

func TestHandleUpload_StorageUploadFails(t *testing.T) {
	fileRepo, storage, _, uc := newTestUploadUseCase()

	fileBytes := []byte("data")
	storage.On("Upload", mock.Anything, mock.Anything, mock.AnythingOfType("string")).
		Return("", assert.AnError)

	resp, err := uc.HandleUpload(context.Background(), 1, fileBytes, "test.jpg", "image")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeFileUploadFailed, err.(*bizerr.Error).Code)
	storage.AssertExpectations(t)
	fileRepo.AssertNotCalled(t, "Create")
}

func TestHandleUpload_FileRepoCreateFails(t *testing.T) {
	fileRepo, storage, _, uc := newTestUploadUseCase()

	fileBytes := []byte("data")
	storage.On("Upload", mock.Anything, mock.Anything, mock.AnythingOfType("string")).
		Return("/uploads/test.jpg", nil)
	fileRepo.On("Create", mock.Anything, mock.Anything).Return(assert.AnError)

	resp, err := uc.HandleUpload(context.Background(), 1, fileBytes, "test.jpg", "image")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeInternalError, err.(*bizerr.Error).Code)
	storage.AssertExpectations(t)
	fileRepo.AssertExpectations(t)
}

func TestHandleUpload_EventPublishFails_DoesNotBlock(t *testing.T) {
	fileRepo, storage, eventBus, uc := newTestUploadUseCase()

	fileBytes := []byte("data")
	storage.On("Upload", mock.Anything, mock.Anything, mock.AnythingOfType("string")).
		Return("/uploads/test.jpg", nil)
	fileRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	eventBus.On("Publish", mock.Anything, "file.uploaded", mock.Anything).Return(assert.AnError)

	resp, err := uc.HandleUpload(context.Background(), 1, fileBytes, "test.jpg", "image")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	storage.AssertExpectations(t)
	fileRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestHandleUpload_DocumentType(t *testing.T) {
	fileRepo, storage, eventBus, uc := newTestUploadUseCase()

	fileBytes := []byte("fake-pdf-content")
	storage.On("Upload", mock.Anything, mock.Anything, mock.AnythingOfType("string")).
		Return("/uploads/doc.pdf", nil)
	fileRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	eventBus.On("Publish", mock.Anything, "file.uploaded", mock.Anything).Return(nil)

	resp, err := uc.HandleUpload(context.Background(), 1, fileBytes, "report.pdf", "document")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "document", resp.Type)
	storage.AssertExpectations(t)
	fileRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestHandleUpload_CaseInsensitiveExtension(t *testing.T) {
	fileRepo, storage, eventBus, uc := newTestUploadUseCase()

	fileBytes := []byte("data")
	storage.On("Upload", mock.Anything, mock.Anything, mock.AnythingOfType("string")).
		Return("/uploads/test.JPG", nil)
	fileRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	eventBus.On("Publish", mock.Anything, "file.uploaded", mock.Anything).Return(nil)

	resp, err := uc.HandleUpload(context.Background(), 1, fileBytes, "photo.JPG", "image")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	storage.AssertExpectations(t)
	fileRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}
