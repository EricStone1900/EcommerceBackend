package push

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

func TestDeleteToken_Success(t *testing.T) {
	repo, _, uc := newTestPushUseCase()

	repo.On("DeleteByUserAndDevice", mock.Anything, uint(1), "device-token-abc").Return(nil)

	err := uc.DeleteToken(context.Background(), 1, DeleteTokenRequest{
		DeviceToken: "device-token-abc",
	})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteToken_EmptyDeviceToken(t *testing.T) {
	_, _, uc := newTestPushUseCase()

	err := uc.DeleteToken(context.Background(), 1, DeleteTokenRequest{
		DeviceToken: "",
	})

	assert.Error(t, err)
	assert.Equal(t, bizerr.CodeValidationError, err.(*bizerr.Error).Code)
}

func TestDeleteToken_Idempotent(t *testing.T) {
	repo, _, uc := newTestPushUseCase()

	// DeleteByUserAndDevice returns nil even when token doesn't exist
	repo.On("DeleteByUserAndDevice", mock.Anything, uint(1), "nonexistent-token").Return(nil)

	err := uc.DeleteToken(context.Background(), 1, DeleteTokenRequest{
		DeviceToken: "nonexistent-token",
	})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDeleteToken_RepoError(t *testing.T) {
	repo, _, uc := newTestPushUseCase()

	repo.On("DeleteByUserAndDevice", mock.Anything, uint(1), "device-token-abc").Return(assert.AnError)

	err := uc.DeleteToken(context.Background(), 1, DeleteTokenRequest{
		DeviceToken: "device-token-abc",
	})

	assert.Error(t, err)
	assert.Equal(t, bizerr.CodeInternalError, err.(*bizerr.Error).Code)
	repo.AssertExpectations(t)
}
