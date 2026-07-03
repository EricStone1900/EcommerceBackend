package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/EricStone1900/ecommerce-backend/internal/usecase/upload"
	"github.com/EricStone1900/ecommerce-backend/pkg/errors"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
)

// --- Mock UploadUseCase ---

type mockUploadUseCase struct {
	mock.Mock
}

func (m *mockUploadUseCase) HandleUpload(ctx context.Context, ownerID uint, fileBytes []byte, filename string, fileType string) (*upload.UploadResponse, error) {
	args := m.Called(ctx, ownerID, fileBytes, filename, fileType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*upload.UploadResponse), args.Error(1)
}

// --- Test helpers ---

func setupUploadHandler(uc *mockUploadUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewUploadHandler(uc)
	r := gin.New()

	// Add middleware to set user_id (simulates auth middleware)
	r.POST("/api/v1/upload", h.Upload)

	return r
}

func successUploadResponse() *upload.UploadResponse {
	return &upload.UploadResponse{
		ID:           1,
		Type:         "image",
		OriginalName: "photo.jpg",
		URL:          "/uploads/abc123.jpg",
		Size:         1024,
		Status:       "pending",
		CreatedAt:    "2026-07-03T10:00:00Z",
	}
}

func createMultipartUploadRequest(t *testing.T, contentType string, fileContent []byte, filename string, fileType string) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Write file part
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)
	_, err = part.Write(fileContent)
	require.NoError(t, err)

	// Write type field
	err = writer.WriteField("type", fileType)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// --- Tests ---

func TestUpload_Success(t *testing.T) {
	mockUC := new(mockUploadUseCase)
	router := setupUploadHandler(mockUC)

	mockUC.On("HandleUpload", mock.Anything, mock.AnythingOfType("uint"), mock.Anything, "photo.jpg", "image").
		Return(successUploadResponse(), nil)

	req := createMultipartUploadRequest(t, "", []byte("fake-image-data"), "photo.jpg", "image")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestUpload_MissingFile(t *testing.T) {
	mockUC := new(mockUploadUseCase)
	router := setupUploadHandler(mockUC)

	// Request with no file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("type", "image")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeInvalidRequest, resp.Code)
}

func TestUpload_MissingType(t *testing.T) {
	mockUC := new(mockUploadUseCase)
	router := setupUploadHandler(mockUC)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "photo.jpg")
	part.Write([]byte("data"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeInvalidFileType, resp.Code)
}

func TestUpload_InvalidFileType(t *testing.T) {
	mockUC := new(mockUploadUseCase)
	router := setupUploadHandler(mockUC)

	mockUC.On("HandleUpload", mock.Anything, mock.AnythingOfType("uint"), mock.Anything, "photo.jpg", "invalid").
		Return(nil, errors.ErrInvalidFileType)

	req := createMultipartUploadRequest(t, "", []byte("data"), "photo.jpg", "invalid")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeInvalidFileType, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestUpload_FileTooLarge(t *testing.T) {
	mockUC := new(mockUploadUseCase)
	router := setupUploadHandler(mockUC)

	mockUC.On("HandleUpload", mock.Anything, mock.AnythingOfType("uint"), mock.Anything, "large.jpg", "image").
		Return(nil, errors.ErrFileTooLarge)

	req := createMultipartUploadRequest(t, "", make([]byte, 100), "large.jpg", "image")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeFileTooLarge, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestUpload_InternalError(t *testing.T) {
	mockUC := new(mockUploadUseCase)
	router := setupUploadHandler(mockUC)

	mockUC.On("HandleUpload", mock.Anything, mock.AnythingOfType("uint"), mock.Anything, "photo.jpg", "image").
		Return(nil, errors.ErrFileUploadFailed)

	req := createMultipartUploadRequest(t, "", []byte("data"), "photo.jpg", "image")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeFileUploadFailed, resp.Code)

	mockUC.AssertExpectations(t)
}

// Test with empty body (no multipart at all)
func TestUpload_EmptyBody(t *testing.T) {
	mockUC := new(mockUploadUseCase)
	router := setupUploadHandler(mockUC)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader([]byte{}))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	// With empty body, PostForm("type") returns "" which triggers CodeInvalidFileType
	assert.Equal(t, errors.CodeInvalidFileType, resp.Code)
}
