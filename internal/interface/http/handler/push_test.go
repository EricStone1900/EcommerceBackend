package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/EricStone1900/ecommerce-backend/internal/usecase/push"
	"github.com/EricStone1900/ecommerce-backend/pkg/errors"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
)

// --- Mock PushUseCase ---

type mockPushUseCase struct {
	mock.Mock
}

func (m *mockPushUseCase) RegisterToken(ctx context.Context, userID uint, req push.RegisterTokenRequest) (*push.RegisterTokenResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*push.RegisterTokenResponse), args.Error(1)
}

func (m *mockPushUseCase) DeleteToken(ctx context.Context, userID uint, req push.DeleteTokenRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *mockPushUseCase) SendTest(ctx context.Context, userID uint) (*push.SendTestResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*push.SendTestResponse), args.Error(1)
}

// --- Test helpers ---

func setupPushHandler(uc *mockPushUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewPushHandler(uc)
	r := gin.New()

	// Simulate auth middleware — set user_id in context
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Next()
	})

	r.POST("/api/v1/push/token", h.RegisterToken)
	r.DELETE("/api/v1/push/token", h.DeleteToken)
	r.POST("/api/v1/push/test", h.SendTest)

	return r
}

func successRegisterTokenResponse() *push.RegisterTokenResponse {
	return &push.RegisterTokenResponse{
		ID:        1,
		Platform:  "ios",
		CreatedAt: "2026-07-03T10:00:00Z",
		UpdatedAt: "2026-07-03T10:00:00Z",
	}
}

// --- Tests ---

func TestRegisterToken_Success(t *testing.T) {
	mockUC := new(mockPushUseCase)
	router := setupPushHandler(mockUC)

	mockUC.On("RegisterToken", mock.Anything, uint(1), push.RegisterTokenRequest{
		DeviceToken: "device-token-abc",
		Platform:    "ios",
	}).Return(successRegisterTokenResponse(), nil)

	body := `{"device_token":"device-token-abc","platform":"ios"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/push/token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestRegisterToken_BadRequest(t *testing.T) {
	mockUC := new(mockPushUseCase)
	router := setupPushHandler(mockUC)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/push/token", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterToken_ValidationError(t *testing.T) {
	mockUC := new(mockPushUseCase)
	router := setupPushHandler(mockUC)

	mockUC.On("RegisterToken", mock.Anything, uint(1), push.RegisterTokenRequest{
		DeviceToken: "",
		Platform:    "ios",
	}).Return(nil, errors.NewValidationError("device_token", "cannot be empty"))

	body := `{"device_token":"","platform":"ios"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/push/token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeValidationError, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestDeleteToken_Success(t *testing.T) {
	mockUC := new(mockPushUseCase)
	router := setupPushHandler(mockUC)

	mockUC.On("DeleteToken", mock.Anything, uint(1), push.DeleteTokenRequest{
		DeviceToken: "device-token-abc",
	}).Return(nil)

	body := `{"device_token":"device-token-abc"}`
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/push/token", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestDeleteToken_BadRequest(t *testing.T) {
	mockUC := new(mockPushUseCase)
	router := setupPushHandler(mockUC)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/push/token", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendTest_Success(t *testing.T) {
	mockUC := new(mockPushUseCase)
	router := setupPushHandler(mockUC)

	mockUC.On("SendTest", mock.Anything, uint(1)).Return(&push.SendTestResponse{Sent: 2}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/push/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestSendTest_NoToken(t *testing.T) {
	mockUC := new(mockPushUseCase)
	router := setupPushHandler(mockUC)

	mockUC.On("SendTest", mock.Anything, uint(1)).Return(nil, errors.ErrPushTokenNotFound)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/push/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodePushTokenNotFound, resp.Code)

	mockUC.AssertExpectations(t)
}
