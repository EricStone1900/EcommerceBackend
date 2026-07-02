package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

func TestLogin_Success(t *testing.T) {
	userRepo, tokenStore, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	user := newTestUser()
	userRepo.On("GetUserByEmail", ctx, "test@example.com").Return(user, nil)
	tokenStore.On("SaveRefreshToken", ctx, uint(1), mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: testPassword,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, uint(1), resp.User.ID)
	userRepo.AssertExpectations(t)
	tokenStore.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo, _, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	user := newTestUser()
	userRepo.On("GetUserByEmail", ctx, "test@example.com").Return(user, nil)

	resp, err := uc.Login(ctx, LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPass1",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, bizerr.ErrInvalidCredentials)
	userRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo, _, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	userRepo.On("GetUserByEmail", ctx, "unknown@example.com").Return(nil, nil)

	resp, err := uc.Login(ctx, LoginRequest{
		Email:    "unknown@example.com",
		Password: "Password1",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, bizerr.ErrInvalidCredentials)
	userRepo.AssertExpectations(t)
}
