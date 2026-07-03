package product

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/port"
	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

func TestList_Success(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	result := newTestProductListResult(3)
	repo.On("List", ctx, mock.AnythingOfType("port.ProductFilter")).Return(result, nil)

	resp, err := uc.List(ctx, ListProductsRequest{
		Page:     1,
		PageSize: 20,
	})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(3), resp.Total)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 20, resp.PageSize)
	assert.Len(t, resp.List.([]*ProductResponse), 3)

	repo.AssertExpectations(t)
}

func TestList_Empty(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	result := newTestProductListResult(0)
	repo.On("List", ctx, mock.AnythingOfType("port.ProductFilter")).Return(result, nil)

	resp, err := uc.List(ctx, ListProductsRequest{})

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(0), resp.Total)
	assert.Equal(t, 1, resp.Page) // default
	assert.Equal(t, 20, resp.PageSize) // default
	assert.Len(t, resp.List.([]*ProductResponse), 0)

	repo.AssertExpectations(t)
}

func TestList_WithNameFilter(t *testing.T) {
	repo, uc := newTestProductUseCase()
	ctx := context.Background()

	result := newTestProductListResult(1)
	repo.On("List", ctx, port.ProductFilter{
		Page:     1,
		PageSize: 20,
		Name:     "test",
		Status:   nil,
		SortBy:   "created_at",
		SortDesc: false,
	}).Return(result, nil)

	resp, err := uc.List(ctx, ListProductsRequest{
		Name:     "test",
		SortDesc: false,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(1), resp.Total)

	repo.AssertExpectations(t)
}

func TestList_InvalidSortBy(t *testing.T) {
	_, uc := newTestProductUseCase()
	ctx := context.Background()

	resp, err := uc.List(ctx, ListProductsRequest{
		SortBy: "invalid_field",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeValidationError, err.(*bizerr.Error).Code)
}

func TestList_InvalidStatus(t *testing.T) {
	_, uc := newTestProductUseCase()
	ctx := context.Background()

	resp, err := uc.List(ctx, ListProductsRequest{
		Status: "bogus",
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, bizerr.CodeValidationError, err.(*bizerr.Error).Code)
}
