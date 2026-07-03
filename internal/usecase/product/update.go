package product

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// Update updates an existing product and returns the response DTO.
func (uc *ProductUseCase) Update(ctx context.Context, id, userID uint, req UpdateProductRequest) (*ProductResponse, error) {
	// Fetch existing product
	existing, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("failed to get product for update", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}
	if existing == nil {
		return nil, bizerr.ErrProductNotFound
	}

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

	// Update fields
	existing.Name = name
	existing.Description = req.Description
	existing.Price = req.Price
	existing.Stock = req.Stock
	existing.Status = status

	if err := uc.productRepo.Update(ctx, existing); err != nil {
		uc.logger.Error("failed to update product", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}

	return productToResponse(existing), nil
}
