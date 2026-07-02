package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

func TestLogout_Success(t *testing.T) {
	userRepo, tokenStore, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	user := newTestUser()

	// Log in first to get a real refresh token
	userRepo.On("GetUserByEmail", ctx, "test@example.com").Return(user, nil)
	tokenStore.On("SaveRefreshToken", ctx, uint(1), mock.Anything, mock.Anything).Return(nil)

	loginResp, err := uc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: testPassword,
	})
	require.NoError(t, err)

	// Now logout
	tokenStore.On("DeleteRefreshToken", ctx, uint(1), mock.Anything).Return(nil)

	err = uc.Logout(ctx, 1, loginResp.RefreshToken)
	assert.NoError(t, err)

	userRepo.AssertExpectations(t)
	tokenStore.AssertExpectations(t)
}

func TestLogout_InvalidToken(t *testing.T) {
	_, _, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	err := uc.Logout(ctx, 1, "invalid-token-string")
	assert.Error(t, err)
	assert.ErrorIs(t, err, bizerr.ErrUnauthorized)
}

func TestLogout_WrongUser(t *testing.T) {
	_, tokenStore, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	tokenString, _, err := newTestJWTService().GenerateRefreshToken(1, "customer")
	require.NoError(t, err)

	// Try to logout as user 2 with user 1's token
	err = uc.Logout(ctx, 2, tokenString)
	assert.Error(t, err)
	assert.ErrorIs(t, err, bizerr.ErrForbidden)
	_ = tokenStore // unused in this test
}

func TestLogout_AccessTokenRejected(t *testing.T) {
	_, _, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	// Generate an access token (not a refresh token)
	accessToken, err := newTestJWTService().GenerateAccessToken(1, "customer")
	require.NoError(t, err)

	err = uc.Logout(ctx, 1, accessToken)
	assert.Error(t, err)
	assert.ErrorIs(t, err, bizerr.ErrInvalidRequest)
}
