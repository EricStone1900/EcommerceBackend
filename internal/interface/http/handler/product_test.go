package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/EricStone1900/ecommerce-backend/internal/usecase/product"
	"github.com/EricStone1900/ecommerce-backend/pkg/errors"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
)

// --- Mock ProductUseCase ---

type mockProductUseCase struct {
	mock.Mock
}

func (m *mockProductUseCase) Create(ctx context.Context, userID uint, req product.CreateProductRequest) (*product.ProductResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.ProductResponse), args.Error(1)
}

func (m *mockProductUseCase) Update(ctx context.Context, id, userID uint, req product.UpdateProductRequest) (*product.ProductResponse, error) {
	args := m.Called(ctx, id, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.ProductResponse), args.Error(1)
}

func (m *mockProductUseCase) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockProductUseCase) GetByID(ctx context.Context, id uint) (*product.ProductResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.ProductResponse), args.Error(1)
}

func (m *mockProductUseCase) List(ctx context.Context, req product.ListProductsRequest) (*response.PaginatedData, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.PaginatedData), args.Error(1)
}

// --- Test helpers ---

func successProductResponse() *product.ProductResponse {
	return &product.ProductResponse{
		ID:          1,
		Name:        "Test Product",
		Description: "A test product",
		Price:       99.99,
		Stock:       50,
		Status:      "on_sale",
		CreatedBy:   1,
		CreatedAt:   "2026-07-01T10:00:00Z",
		UpdatedAt:   "2026-07-01T10:00:00Z",
	}
}

func setupProductHandler(uc *mockProductUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewProductHandler(uc)
	r := gin.New()

	// Register routes matching the real setup
	r.GET("/api/v1/products/:id", h.GetDetail)
	r.GET("/api/v1/products", h.List)
	r.POST("/api/v1/products", h.Create)
	r.PUT("/api/v1/products/:id", h.Update)
	r.DELETE("/api/v1/products/:id", h.Delete)

	return r
}

// --- Tests ---

func TestGetDetail_Success(t *testing.T) {
	mockUC := new(mockProductUseCase)
	router := setupProductHandler(mockUC)

	mockUC.On("GetByID", mock.Anything, uint(1)).Return(successProductResponse(), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestGetDetail_NotFound(t *testing.T) {
	mockUC := new(mockProductUseCase)
	router := setupProductHandler(mockUC)

	mockUC.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.ErrProductNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeProductNotFound, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestGetDetail_InvalidID(t *testing.T) {
	router := setupProductHandler(new(mockProductUseCase))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/abc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeInvalidRequest, resp.Code)
}

func TestList_Success(t *testing.T) {
	mockUC := new(mockProductUseCase)
	router := setupProductHandler(mockUC)

	paginated := &response.PaginatedData{
		Total:    1,
		Page:     1,
		PageSize: 20,
		List:     []*product.ProductResponse{successProductResponse()},
	}
	mockUC.On("List", mock.Anything, mock.AnythingOfType("product.ListProductsRequest")).Return(paginated, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestList_Empty(t *testing.T) {
	mockUC := new(mockProductUseCase)
	router := setupProductHandler(mockUC)

	paginated := &response.PaginatedData{
		Total:    0,
		Page:     1,
		PageSize: 20,
		List:     []*product.ProductResponse{},
	}
	mockUC.On("List", mock.Anything, mock.AnythingOfType("product.ListProductsRequest")).Return(paginated, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestCreate_Success(t *testing.T) {
	mockUC := new(mockProductUseCase)
	router := setupProductHandler(mockUC)

	mockUC.On("Create", mock.Anything, uint(0), product.CreateProductRequest{
		Name:        "New Product",
		Description: "A new product",
		Price:       49.99,
		Stock:       100,
	}).Return(successProductResponse(), nil)

	body := `{"name":"New Product","description":"A new product","price":49.99,"stock":100}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestCreate_ValidationError(t *testing.T) {
	mockUC := new(mockProductUseCase)
	router := setupProductHandler(mockUC)

	mockUC.On("Create", mock.Anything, uint(0), product.CreateProductRequest{
		Name:  "",
		Price: 0,
		Stock: 0,
	}).Return(nil, errors.NewValidationError("name", "cannot be empty"))

	body := `{"name":"","price":0,"stock":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeValidationError, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestUpdate_Success(t *testing.T) {
	mockUC := new(mockProductUseCase)
	router := setupProductHandler(mockUC)

	mockUC.On("Update", mock.Anything, uint(1), uint(0), product.UpdateProductRequest{
		Name:        "Updated",
		Description: "Updated desc",
		Price:       19.99,
		Stock:       200,
		Status:      "off_sale",
	}).Return(successProductResponse(), nil)

	body := `{"name":"Updated","description":"Updated desc","price":19.99,"stock":200,"status":"off_sale"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/products/1",
		bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestDelete_Success(t *testing.T) {
	mockUC := new(mockProductUseCase)
	router := setupProductHandler(mockUC)

	mockUC.On("Delete", mock.Anything, uint(1)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/products/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.Code)

	mockUC.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	mockUC := new(mockProductUseCase)
	router := setupProductHandler(mockUC)

	mockUC.On("Delete", mock.Anything, uint(999)).Return(errors.ErrProductNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/products/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeProductNotFound, resp.Code)

	mockUC.AssertExpectations(t)
}
