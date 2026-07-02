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

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/auth"
	"github.com/EricStone1900/ecommerce-backend/pkg/errors"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
)

// --- Mock AuthUseCase ---

type mockAuthUseCase struct {
	mock.Mock
}

func (m *mockAuthUseCase) Register(ctx context.Context, req auth.RegisterRequest) (*auth.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AuthResponse), args.Error(1)
}

func (m *mockAuthUseCase) Login(ctx context.Context, req auth.LoginRequest) (*auth.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AuthResponse), args.Error(1)
}

func (m *mockAuthUseCase) Logout(ctx context.Context, userID uint, refreshToken string) error {
	args := m.Called(ctx, userID, refreshToken)
	return args.Error(0)
}

func (m *mockAuthUseCase) Refresh(ctx context.Context, req auth.RefreshRequest) (*auth.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.AuthResponse), args.Error(1)
}

func (m *mockAuthUseCase) GetCurrentUser(ctx context.Context, userID uint) (*auth.UserInfo, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserInfo), args.Error(1)
}

// --- Test helpers ---

func setupAuthHandler(uc *mockAuthUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(uc)
	r := gin.New()
	r.POST("/api/v1/auth/register", h.Register)
	r.POST("/api/v1/auth/login", h.Login)
	return r
}

func successAuthResponse() *auth.AuthResponse {
	return &auth.AuthResponse{
		AccessToken:  "access.token.here",
		RefreshToken: "refresh.token.here",
		User: auth.UserInfo{
			ID:    1,
			Email: "test@example.com",
			Role:  entity.RoleCustomer,
		},
	}
}

// --- Tests ---

func TestRegisterHandler_Success(t *testing.T) {
	mockUC := new(mockAuthUseCase)
	router := setupAuthHandler(mockUC)

	body := `{"email":"test@example.com","password":"Password1"}`
	mockUC.On("Register", mock.Anything, auth.RegisterRequest{
		Email:    "test@example.com",
		Password: "Password1",
	}).Return(successAuthResponse(), nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestRegisterHandler_ValidationError(t *testing.T) {
	mockUC := new(mockAuthUseCase)
	router := setupAuthHandler(mockUC)

	// Invalid JSON body
	body := `{"email":"invalid","password":"weak"}`
	mockUC.On("Register", mock.Anything, auth.RegisterRequest{
		Email:    "invalid",
		Password: "weak",
	}).Return(nil, errors.NewValidationError("email", "invalid email format"))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	mockUC.AssertExpectations(t)
}

func TestRegisterHandler_BadRequestBody(t *testing.T) {
	mockUC := new(mockAuthUseCase)
	router := setupAuthHandler(mockUC)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register",
		bytes.NewBufferString(`invalid json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestLoginHandler_Success(t *testing.T) {
	mockUC := new(mockAuthUseCase)
	router := setupAuthHandler(mockUC)

	mockUC.On("Login", mock.Anything, auth.LoginRequest{
		Email:    "test@example.com",
		Password: "Password1",
	}).Return(successAuthResponse(), nil)

	body := `{"email":"test@example.com","password":"Password1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		bytes.NewBufferString(body))
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

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	mockUC := new(mockAuthUseCase)
	router := setupAuthHandler(mockUC)

	mockUC.On("Login", mock.Anything, auth.LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPass1",
	}).Return(nil, errors.ErrInvalidCredentials)

	body := `{"email":"test@example.com","password":"WrongPass1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeInvalidCredentials, resp.Code)

	mockUC.AssertExpectations(t)
}
