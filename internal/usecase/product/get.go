package product

import (
	"context"

	"go.uber.org/zap"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// GetByID retrieves a single product by ID.
func (uc *ProductUseCase) GetByID(ctx context.Context, id uint) (*ProductResponse, error) {
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("failed to get product by id", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}
	if product == nil {
		return nil, bizerr.ErrProductNotFound
	}

	return productToResponse(product), nil
}
