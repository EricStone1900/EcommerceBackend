package product

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// Create creates a new product and returns the response DTO.
func (uc *ProductUseCase) Create(ctx context.Context, userID uint, req CreateProductRequest) (*ProductResponse, error) {
	// Validate name
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, bizerr.NewValidationError("name", "cannot be empty")
	}

	// Validate price
	if req.Price <= 0 {
		return nil, bizerr.NewValidationError("price", "must be greater than 0")
	}

	// Validate stock
	if req.Stock < 0 {
		return nil, bizerr.NewValidationError("stock", "cannot be negative")
	}

	// Validate and parse status
	status := entity.ProductStatusOnSale
	if req.Status != "" {
		switch entity.ProductStatus(req.Status) {
		case entity.ProductStatusOnSale, entity.ProductStatusOffSale:
			status = entity.ProductStatus(req.Status)
		default:
			return nil, bizerr.NewValidationError("status", "must be 'on_sale' or 'off_sale'")
		}
	}

	product := &entity.Product{
		Name:        name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Status:      status,
		CreatedBy:   userID,
	}

	if err := uc.productRepo.Create(ctx, product); err != nil {
		uc.logger.Error("failed to create product", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}

	return productToResponse(product), nil
}
