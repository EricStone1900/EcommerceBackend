package product

import (
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/internal/domain/port"
	"go.uber.org/zap"
)

// mockProductRepo implements port.ProductRepository with testify/mock.
type mockProductRepo struct {
	mock.Mock
}

func (m *mockProductRepo) Create(ctx context.Context, product *entity.Product) error {
	args := m.Called(ctx, product)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	// Simulate DB-assigned fields
	product.ID = 1
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	return nil
}

func (m *mockProductRepo) Update(ctx context.Context, product *entity.Product) error {
	args := m.Called(ctx, product)
	if args.Get(0) != nil {
		return args.Error(0)
	}
	product.UpdatedAt = time.Now()
	return nil
}

func (m *mockProductRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockProductRepo) GetByID(ctx context.Context, id uint) (*entity.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockProductRepo) List(ctx context.Context, filter port.ProductFilter) (*port.ProductListResult, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.ProductListResult), args.Error(1)
}

// newTestProductUseCase creates a test ProductUseCase with a mock repo.
func newTestProductUseCase() (*mockProductRepo, *ProductUseCase) {
	repo := new(mockProductRepo)
	logger, _ := zap.NewDevelopment()
	uc := NewProductUseCase(repo, logger)
	return repo, uc
}

// newTestProduct creates a sample product entity for testing.
func newTestProduct() *entity.Product {
	return &entity.Product{
		ID:          1,
		Name:        "Test Product",
		Description: "A test product description",
		Price:       99.99,
		Stock:       50,
		Status:      entity.ProductStatusOnSale,
		CreatedBy:   1,
		CreatedAt:   time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC),
	}
}

// newTestProductListResult creates a sample list result for testing.
func newTestProductListResult(count int) *port.ProductListResult {
	products := make([]*entity.Product, count)
	for i := 0; i < count; i++ {
		p := newTestProduct()
		p.ID = uint(i + 1)
		p.Name = fmt.Sprintf("Product %d", i+1)
		products[i] = p
	}
	return &port.ProductListResult{
		Products: products,
		Total:    int64(count),
	}
}
