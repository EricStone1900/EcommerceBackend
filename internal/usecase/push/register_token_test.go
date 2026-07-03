package push

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

func TestRegisterToken_Success(t *testing.T) {
	repo, _, uc := newTestPushUseCase()

	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.PushToken")).Return(nil)

	resp, err := uc.RegisterToken(context.Background(), 1, RegisterTokenRequest{
		DeviceToken: "device-token-abc",
		Platform:    "ios",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, "ios", resp.Platform)
	repo.AssertExpectations(t)
}

func TestRegisterToken_EmptyDeviceToken(t *testing.T) {
	_, _, uc := newTestPushUseCase()

	resp, err := uc.RegisterToken(context.Background(), 1, RegisterTokenRequest{
		DeviceToken: "",
		Platform:    "ios",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeValidationError, err.(*bizerr.Error).Code)
}

func TestRegisterToken_InvalidPlatform(t *testing.T) {
	_, _, uc := newTestPushUseCase()

	resp, err := uc.RegisterToken(context.Background(), 1, RegisterTokenRequest{
		DeviceToken: "device-token-abc",
		Platform:    "android",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeInvalidPlatform, err.(*bizerr.Error).Code)
}

func TestRegisterToken_RepoError(t *testing.T) {
	repo, _, uc := newTestPushUseCase()

	repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.PushToken")).Return(assert.AnError)

	resp, err := uc.RegisterToken(context.Background(), 1, RegisterTokenRequest{
		DeviceToken: "device-token-abc",
		Platform:    "ios",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeInternalError, err.(*bizerr.Error).Code)
	repo.AssertExpectations(t)
}
