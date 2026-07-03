package product

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

func TestCreate_Success(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	repo.On("Create", ctx, mock.AnythingOfType("*entity.Product")).Return(nil)

	resp, err := uc.Create(ctx, 1, CreateProductRequest{
		Name:        "New Product",
		Description: "A new product",
		Price:       49.99,
		Stock:       100,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "New Product", resp.Name)
	assert.Equal(t, 49.99, resp.Price)
	assert.Equal(t, 100, resp.Stock)
	assert.Equal(t, "on_sale", resp.Status) // default status

	repo.AssertExpectations(t)
}

func TestCreate_EmptyName(t *testing.T) {
	_, uc := newTestProductUseCase()
	ctx := context.Background()

	resp, err := uc.Create(ctx, 1, CreateProductRequest{
		Name:  "",
		Price: 49.99,
		Stock: 100,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeValidationError, err.(*bizerr.Error).Code)
}

func TestCreate_ZeroPrice(t *testing.T) {
	_, uc := newTestProductUseCase()
	ctx := context.Background()

	resp, err := uc.Create(ctx, 1, CreateProductRequest{
		Name:  "Product",
		Price: 0,
		Stock: 100,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeValidationError, err.(*bizerr.Error).Code)
}

func TestCreate_NegativeStock(t *testing.T) {
	_, uc := newTestProductUseCase()
	ctx := context.Background()

	resp, err := uc.Create(ctx, 1, CreateProductRequest{
		Name:  "Product",
		Price: 10.0,
		Stock: -1,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeValidationError, err.(*bizerr.Error).Code)
}

func TestCreate_InvalidStatus(t *testing.T) {
	_, uc := newTestProductUseCase()
	ctx := context.Background()

	resp, err := uc.Create(ctx, 1, CreateProductRequest{
		Name:   "Product",
		Price:  10.0,
		Stock:  10,
		Status: "invalid_status",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeValidationError, err.(*bizerr.Error).Code)
}

func TestCreate_RepoError(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	repo.On("Create", ctx, mock.AnythingOfType("*entity.Product")).Return(assert.AnError)

	resp, err := uc.Create(ctx, 1, CreateProductRequest{
		Name:  "Product",
		Price: 10.0,
		Stock: 10,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, bizerr.ErrInternalError)

	repo.AssertExpectations(t)
}
