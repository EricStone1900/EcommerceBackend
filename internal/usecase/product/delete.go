package product

import (
	"context"

	"go.uber.org/zap"

	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// Delete performs a soft delete on a product.
func (uc *ProductUseCase) Delete(ctx context.Context, id uint) error {
	// Check existence first
	existing, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("failed to get product for delete", zap.Error(err))
		return bizerr.ErrInternalError
	}
	if existing == nil {
		return bizerr.ErrProductNotFound
	}

	if err := uc.productRepo.Delete(ctx, id); err != nil {
		uc.logger.Error("failed to delete product", zap.Error(err))
		return bizerr.ErrInternalError
	}

	return nil
}
