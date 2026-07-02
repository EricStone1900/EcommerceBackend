package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/EricStone1900/ecommerce-backend/pkg/jwt"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

func TestRefresh_Success(t *testing.T) {
	userRepo, tokenStore, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	user := newTestUser()

	// Log in first
	userRepo.On("GetUserByEmail", ctx, "test@example.com").Return(user, nil)
	tokenStore.On("SaveRefreshToken", ctx, uint(1), mock.Anything, mock.Anything).Return(nil)

	loginResp, err := uc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: testPassword,
	})
	require.NoError(t, err)

	// Refresh: validate token exists, delete old, issue new
	tokenStore.On("ValidateRefreshToken", ctx, uint(1), mock.Anything).Return(true, nil)
	tokenStore.On("DeleteRefreshToken", ctx, uint(1), mock.Anything).Return(nil)
	userRepo.On("GetUserByID", ctx, uint(1)).Return(user, nil)
	tokenStore.On("SaveRefreshToken", ctx, uint(1), mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Refresh(ctx, RefreshRequest{
		RefreshToken: loginResp.RefreshToken,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.NotEqual(t, loginResp.RefreshToken, resp.RefreshToken, "refresh token should be rotated")

	userRepo.AssertExpectations(t)
	tokenStore.AssertExpectations(t)
}

func TestRefresh_ExpiredToken(t *testing.T) {
	_, _, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	// Create an expired refresh token manually
	expiredService := jwt.NewJWTService("test_secret_key_for_ut", 15*time.Minute, -1*time.Hour)
	expiredToken, _, err := expiredService.GenerateRefreshToken(1, "customer")
	require.NoError(t, err)

	resp, err := uc.Refresh(ctx, RefreshRequest{
		RefreshToken: expiredToken,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, bizerr.ErrTokenExpired)
}

func TestRefresh_RevokedToken(t *testing.T) {
	userRepo, tokenStore, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	user := newTestUser()

	// Log in first
	userRepo.On("GetUserByEmail", ctx, "test@example.com").Return(user, nil)
	tokenStore.On("SaveRefreshToken", ctx, uint(1), mock.Anything, mock.Anything).Return(nil)

	loginResp, err := uc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: testPassword,
	})
	require.NoError(t, err)

	// Token not found in Redis (revoked)
	tokenStore.On("ValidateRefreshToken", ctx, uint(1), mock.Anything).Return(false, nil)

	resp, err := uc.Refresh(ctx, RefreshRequest{
		RefreshToken: loginResp.RefreshToken,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, bizerr.ErrUnauthorized)
}

func TestRefresh_AccessTokenRejected(t *testing.T) {
	_, _, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	accessToken, err := newTestJWTService().GenerateAccessToken(1, "customer")
	require.NoError(t, err)

	resp, err := uc.Refresh(ctx, RefreshRequest{
		RefreshToken: accessToken,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, bizerr.ErrInvalidRequest)
}
