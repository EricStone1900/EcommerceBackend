package product

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

func TestUpdate_Success(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	existing := newTestProduct()
	repo.On("GetByID", ctx, uint(1)).Return(existing, nil)
	repo.On("Update", ctx, mock.AnythingOfType("*entity.Product")).Return(nil)

	resp, err := uc.Update(ctx, 1, 1, UpdateProductRequest{
		Name:        "Updated Product",
		Description: "Updated description",
		Price:       29.99,
		Stock:       200,
		Status:      "off_sale",
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Updated Product", resp.Name)
	assert.Equal(t, 29.99, resp.Price)
	assert.Equal(t, 200, resp.Stock)
	assert.Equal(t, "off_sale", resp.Status)

	repo.AssertExpectations(t)
}

func TestUpdate_NotFound(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	repo.On("GetByID", ctx, uint(999)).Return(nil, nil)

	resp, err := uc.Update(ctx, 999, 1, UpdateProductRequest{
		Name:  "Product",
		Price: 10.0,
		Stock: 10,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, bizerr.ErrProductNotFound)

	repo.AssertExpectations(t)
}

func TestUpdate_EmptyName(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	repo.On("GetByID", ctx, uint(1)).Return(newTestProduct(), nil)

	resp, err := uc.Update(ctx, 1, 1, UpdateProductRequest{
		Name:  "",
		Price: 10.0,
		Stock: 10,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeValidationError, err.(*bizerr.Error).Code)

	repo.AssertExpectations(t)
}
