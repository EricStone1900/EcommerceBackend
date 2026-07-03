package push

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

func TestSendTest_Success(t *testing.T) {
	repo, notifier, uc := newTestPushUseCase()

	tokens := newTestPushTokens(1)
	repo.On("GetByUserID", mock.Anything, uint(1)).Return(tokens, nil)
	notifier.On("SendPush", mock.Anything, uint(1), tokens[0].DeviceToken, "Test Notification", "This is a test push from the ecommerce backend").Return(nil)

	resp, err := uc.SendTest(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.Sent)
	repo.AssertExpectations(t)
	notifier.AssertExpectations(t)
}

func TestSendTest_MultipleTokens(t *testing.T) {
	repo, notifier, uc := newTestPushUseCase()

	tokens := newTestPushTokens(3)
	repo.On("GetByUserID", mock.Anything, uint(1)).Return(tokens, nil)
	for _, tok := range tokens {
		notifier.On("SendPush", mock.Anything, uint(1), tok.DeviceToken, "Test Notification", "This is a test push from the ecommerce backend").Return(nil)
	}

	resp, err := uc.SendTest(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 3, resp.Sent)
	repo.AssertExpectations(t)
	notifier.AssertExpectations(t)
}

func TestSendTest_NoTokens(t *testing.T) {
	repo, _, uc := newTestPushUseCase()

	repo.On("GetByUserID", mock.Anything, uint(1)).Return(nil, nil)

	resp, err := uc.SendTest(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodePushTokenNotFound, err.(*bizerr.Error).Code)
	repo.AssertExpectations(t)
}

func TestSendTest_PushFailure(t *testing.T) {
	repo, notifier, uc := newTestPushUseCase()

	tokens := newTestPushTokens(1)
	repo.On("GetByUserID", mock.Anything, uint(1)).Return(tokens, nil)
	notifier.On("SendPush", mock.Anything, uint(1), tokens[0].DeviceToken, "Test Notification", "This is a test push from the ecommerce backend").Return(assert.AnError)

	resp, err := uc.SendTest(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodePushSendFailed, err.(*bizerr.Error).Code)
	repo.AssertExpectations(t)
	notifier.AssertExpectations(t)
}

func TestSendTest_RepoError(t *testing.T) {
	repo, _, uc := newTestPushUseCase()

	repo.On("GetByUserID", mock.Anything, uint(1)).Return(nil, assert.AnError)

	resp, err := uc.SendTest(context.Background(), 1)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeInternalError, err.(*bizerr.Error).Code)
	repo.AssertExpectations(t)
}
