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

// @Summary      获取商品详情
// @Description  根据商品 ID 获取商品详细信息
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Param        id path int true "商品 ID"
// @Success      200 {object} response.Response{data=product.ProductResponse}
// @Failure      400 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/products/{id} [get]
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

// @Summary      获取商品列表
// @Description  分页获取商品列表，支持按名称、状态筛选和排序
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        page query int false "页码（默认 1）"
// @Param        page_size query int false "每页条数（默认 20）"
// @Param        name query string false "商品名称（模糊搜索）"
// @Param        status query string false "商品状态（on_sale|off_sale）"
// @Param        sort_by query string false "排序字段（created_at|price）"
// @Param        sort_desc query bool false "是否降序"
// @Success      200 {object} response.Response{data=response.PaginatedData{list=[]product.ProductResponse}}
// @Failure      400 {object} response.Response
// @Router       /api/v1/products [get]
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

// @Summary      创建商品
// @Description  创建新商品（仅管理员）
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body product.CreateProductRequest true "商品信息"
// @Success      201 {object} response.Response{data=product.ProductResponse}
// @Failure      400 {object} response.Response
// @Failure      403 {object} response.Response
// @Router       /api/v1/products [post]
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

// @Summary      更新商品
// @Description  更新指定商品的信息（仅管理员）
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "商品 ID"
// @Param        request body product.UpdateProductRequest true "更新信息"
// @Success      200 {object} response.Response{data=product.ProductResponse}
// @Failure      400 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/products/{id} [put]
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

// @Summary      删除商品
// @Description  软删除指定商品（仅管理员）
// @Tags         商品管理
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path int true "商品 ID"
// @Success      200 {object} response.Response
// @Failure      400 {object} response.Response
// @Failure      403 {object} response.Response
// @Failure      404 {object} response.Response
// @Router       /api/v1/products/{id} [delete]
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
