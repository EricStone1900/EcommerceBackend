package product

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

func TestGetByID_Success(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	repo.On("GetByID", ctx, uint(1)).Return(newTestProduct(), nil)

	resp, err := uc.GetByID(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Test Product", resp.Name)
	assert.Equal(t, 99.99, resp.Price)

	repo.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	repo.On("GetByID", ctx, uint(999)).Return(nil, nil)

	resp, err := uc.GetByID(ctx, 999)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, bizerr.ErrProductNotFound)

	repo.AssertExpectations(t)
}

func TestGetByID_RepoError(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	repo.On("GetByID", ctx, uint(1)).Return(nil, assert.AnError)

	resp, err := uc.GetByID(ctx, 1)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, bizerr.ErrInternalError)

	repo.AssertExpectations(t)
}
