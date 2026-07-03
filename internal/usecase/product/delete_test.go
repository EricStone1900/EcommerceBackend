package product

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

func TestDelete_Success(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	repo.On("GetByID", ctx, uint(1)).Return(newTestProduct(), nil)
	repo.On("Delete", ctx, uint(1)).Return(nil)

	err := uc.Delete(ctx, 1)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	repo.On("GetByID", ctx, uint(999)).Return(nil, nil)

	err := uc.Delete(ctx, 999)

	assert.Error(t, err)
	assert.ErrorIs(t, err, bizerr.ErrProductNotFound)
	repo.AssertExpectations(t)
}

func TestDelete_RepoError(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	repo.On("GetByID", ctx, uint(1)).Return(newTestProduct(), nil)
	repo.On("Delete", ctx, uint(1)).Return(assert.AnError)

	err := uc.Delete(ctx, 1)

	assert.Error(t, err)
	assert.ErrorIs(t, err, bizerr.ErrInternalError)
	repo.AssertExpectations(t)
}
