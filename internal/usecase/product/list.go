package product

import (
	"context"

	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/internal/domain/port"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// List returns a paginated list of products with optional filtering and sorting.
func (uc *ProductUseCase) List(ctx context.Context, req ListProductsRequest) (*response.PaginatedData, error) {
	// Apply defaults
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Parse status filter
	var statusFilter *entity.ProductStatus
	if req.Status != "" {
		s := entity.ProductStatus(req.Status)
		switch s {
		case entity.ProductStatusOnSale, entity.ProductStatusOffSale:
			statusFilter = &s
		default:
			return nil, bizerr.NewValidationError("status", "must be 'on_sale' or 'off_sale'")
		}
	}

	// Validate sort_by
	sortBy := "created_at"
	if req.SortBy == "price" {
		sortBy = "price"
	} else if req.SortBy != "" && req.SortBy != "created_at" {
		return nil, bizerr.NewValidationError("sort_by", "must be 'created_at' or 'price'")
	}

	filter := port.ProductFilter{
		Page:     page,
		PageSize: pageSize,
		Name:     req.Name,
		Status:   statusFilter,
		SortBy:   sortBy,
		SortDesc: req.SortDesc,
	}

	result, err := uc.productRepo.List(ctx, filter)
	if err != nil {
		uc.logger.Error("failed to list products", zap.Error(err))
		return nil, bizerr.ErrInternalError
	}

	// Convert to response DTOs
	responses := make([]*ProductResponse, len(result.Products))
	for i, p := range result.Products {
		responses[i] = productToResponse(p)
	}

	return &response.PaginatedData{
		Total:    result.Total,
		Page:     page,
		PageSize: pageSize,
		List:     responses,
	}, nil
}
