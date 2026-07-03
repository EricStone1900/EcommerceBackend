package port

import (
	"context"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
)

// ProductRepository defines the persistence interface for products.
// Implementations: gorm.ProductRepository
type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*entity.Product, error)
	List(ctx context.Context, filter ProductFilter) (*ProductListResult, error)
}

// ProductFilter holds pagination, search, and sorting parameters for listing products.
type ProductFilter struct {
	Page     int
	PageSize int
	Name     string
	Status   *entity.ProductStatus // nil means no status filter
	SortBy   string                // "created_at" or "price"; defaults to "created_at"
	SortDesc bool                  // defaults to true (newest first)
}

// ProductListResult contains the paginated product list and total count.
type ProductListResult struct {
	Products []*entity.Product
	Total    int64
}
