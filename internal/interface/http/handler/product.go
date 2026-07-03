package handler

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/middleware"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/product"
	"github.com/EricStone1900/ecommerce-backend/pkg/errors"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
)

// ProductUseCase defines the interface for product operations.
// This abstraction enables testability via mock implementations.
type ProductUseCase interface {
	Create(ctx context.Context, userID uint, req product.CreateProductRequest) (*product.ProductResponse, error)
	Update(ctx context.Context, id, userID uint, req product.UpdateProductRequest) (*product.ProductResponse, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*product.ProductResponse, error)
	List(ctx context.Context, req product.ListProductsRequest) (*response.PaginatedData, error)
}

// ProductHandler handles HTTP requests for product endpoints.
type ProductHandler struct {
	productUsecase ProductUseCase
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(productUsecase ProductUseCase) *ProductHandler {
	return &ProductHandler{productUsecase: productUsecase}
}

// GetDetail handles GET /api/v1/products/:id (public)
func (h *ProductHandler) GetDetail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, response.Error(errors.CodeInvalidRequest, "invalid product id"))
		return
	}

	resp, err := h.productUsecase.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		handleProductError(c, err)
		return
	}

	c.JSON(200, response.Success(resp))
}

// List handles GET /api/v1/products (protected, any role)
func (h *ProductHandler) List(c *gin.Context) {
	var req product.ListProductsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, response.Error(errors.CodeInvalidRequest, "invalid query parameters"))
		return
	}

	resp, err := h.productUsecase.List(c.Request.Context(), req)
	if err != nil {
		handleProductError(c, err)
		return
	}

	c.JSON(200, response.Success(resp))
}

// Create handles POST /api/v1/products (protected, admin only)
func (h *ProductHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req product.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(errors.CodeInvalidRequest, "invalid request body"))
		return
	}

	resp, err := h.productUsecase.Create(c.Request.Context(), userID, req)
	if err != nil {
		handleProductError(c, err)
		return
	}

	c.JSON(201, response.Success(resp))
}

// Update handles PUT /api/v1/products/:id (protected, admin only)
func (h *ProductHandler) Update(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, response.Error(errors.CodeInvalidRequest, "invalid product id"))
		return
	}

	var req product.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, response.Error(errors.CodeInvalidRequest, "invalid request body"))
		return
	}

	resp, err := h.productUsecase.Update(c.Request.Context(), uint(id), userID, req)
	if err != nil {
		handleProductError(c, err)
		return
	}

	c.JSON(200, response.Success(resp))
}

// Delete handles DELETE /api/v1/products/:id (protected, admin only)
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, response.Error(errors.CodeInvalidRequest, "invalid product id"))
		return
	}

	if err := h.productUsecase.Delete(c.Request.Context(), uint(id)); err != nil {
		handleProductError(c, err)
		return
	}

	c.JSON(200, response.Success(nil))
}

// handleProductError converts a use case error into the appropriate HTTP response.
func handleProductError(c *gin.Context, err error) {
	if bizErr, ok := err.(*errors.Error); ok {
		c.JSON(bizErr.HTTPCode, response.Error(bizErr.Code, bizErr.Message))
		return
	}
	c.JSON(500, response.Error(errors.CodeInternalError, "internal server error"))
}
