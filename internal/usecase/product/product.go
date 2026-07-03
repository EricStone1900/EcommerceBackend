package product

import (
	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/internal/domain/port"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
)

// --- Request DTOs ---

// CreateProductRequest is the input for creating a new product.
type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Status      string  `json:"status"` // optional, defaults to "on_sale"
}

// UpdateProductRequest is the input for updating an existing product.
type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Status      string  `json:"status"`
}

// ListProductsRequest holds pagination, search, and sort parameters.
type ListProductsRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Name     string `form:"name"`
	Status   string `form:"status"`
	SortBy   string `form:"sort_by"`  // "created_at" or "price"
	SortDesc bool   `form:"sort_desc"` // defaults to true
}

// --- Response DTOs ---

// ProductResponse is the public-facing product representation.
type ProductResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Status      string  `json:"status"`
	CreatedBy   uint    `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// productToResponse converts a domain entity to the response DTO.
func productToResponse(p *entity.Product) *ProductResponse {
	if p == nil {
		return nil
	}
	return &ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		Status:      string(p.Status),
		CreatedBy:   p.CreatedBy,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ProductUseCase implements product-related business logic.
type ProductUseCase struct {
	productRepo port.ProductRepository
	logger      *zap.Logger
}

// NewProductUseCase creates a new ProductUseCase.
func NewProductUseCase(productRepo port.ProductRepository, logger *zap.Logger) *ProductUseCase {
	return &ProductUseCase{
		productRepo: productRepo,
		logger:      logger,
	}
}

// PaginatedResponse wraps list results into a uniform paginated format.
func PaginatedResponse(total int64, page, pageSize int, list interface{}) *response.PaginatedData {
	return &response.PaginatedData{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     list,
	}
}

// ParseProductStatus converts a string to ProductStatus. Returns nil if empty.
func ParseProductStatus(s string) *entity.ProductStatus {
	if s == "" {
		return nil
	}
	status := entity.ProductStatus(s)
	return &status
}
