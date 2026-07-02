package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

func TestRegister_Success(t *testing.T) {
	userRepo, tokenStore, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	// Email uniqueness — user not found
	userRepo.On("GetUserByEmail", ctx, "test@example.com").Return(nil, nil)
	// Create user — simulates success
	userRepo.On("CreateUser", ctx, mock.AnythingOfType("*entity.User")).Return(nil)
	// Save refresh token
	tokenStore.On("SaveRefreshToken", ctx, uint(1), mock.Anything, mock.Anything).Return(nil)

	resp, err := uc.Register(ctx, RegisterRequest{
		Email:    "Test@Example.com",
		Password: "Password1",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)
	assert.Equal(t, uint(1), resp.User.ID)
	assert.Equal(t, "test@example.com", resp.User.Email)
	assert.Equal(t, entity.RoleCustomer, resp.User.Role)
	userRepo.AssertExpectations(t)
	tokenStore.AssertExpectations(t)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	userRepo, _, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	existingUser := newTestUser()
	userRepo.On("GetUserByEmail", ctx, "test@example.com").Return(existingUser, nil)

	resp, err := uc.Register(ctx, RegisterRequest{
		Email:    "test@example.com",
		Password: "Password1",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, bizerr.ErrEmailAlreadyExists)
	userRepo.AssertExpectations(t)
}

func TestRegister_InvalidEmail(t *testing.T) {
	_, _, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	invalidEmails := []string{"", "notanemail", "@domain.com", "user@", "user@.com"}
	for _, email := range invalidEmails {
		resp, err := uc.Register(ctx, RegisterRequest{
			Email:    email,
			Password: "Password1",
		})
		assert.Error(t, err, "email: %s should be invalid", email)
		assert.Nil(t, resp)
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	_, _, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	weakPasswords := []string{"short", "onlyletters", "1234567890", ""}
	for _, pwd := range weakPasswords {
		resp, err := uc.Register(ctx, RegisterRequest{
			Email:    "test@example.com",
			Password: pwd,
		})
		assert.Error(t, err, "password: %q should be rejected", pwd)
		assert.Nil(t, resp)
	}
}

func TestRegister_RepoError(t *testing.T) {
	userRepo, _, uc := newTestAuthUseCase(t)
	ctx := context.Background()

	userRepo.On("GetUserByEmail", ctx, "test@example.com").Return(nil, nil)
	userRepo.On("CreateUser", ctx, mock.Anything).Return(errors.New("db connection failed"))

	resp, err := uc.Register(ctx, RegisterRequest{
		Email:    "test@example.com",
		Password: "Password1",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, bizerr.ErrInternalError)
	userRepo.AssertExpectations(t)
}
