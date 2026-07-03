package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EricStone1900/ecommerce-backend/pkg/errors"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
)

func TestProduct_CreateAndList(t *testing.T) {
	// This test exercises the admin product flow:
	// Admin login -> Create product -> List products -> Get product detail.

	// We need both auth and product routers since we login first, then
	// perform product operations.
	authRouter := newAuthRouter()
	productRouter := newProductRouter()

	// 1. Login as the seeded admin user.
	adminTokens := loginUser(t, authRouter, testAdminEmail, testAdminPassword)
	assert.Equal(t, "admin", string(adminTokens.UserRole))

	// 2. Admin creates a product.
	createBody := `{"name":"Test Widget","description":"A high-quality widget","price":29.99,"stock":100}`
	w := httpDo(t, productRouter, http.MethodPost, "/api/v1/products", createBody, adminTokens.AccessToken)
	require.Equal(t, http.StatusCreated, w.Code)

	var createResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ID          uint    `json:"id"`
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Price       float64 `json:"price"`
			Stock       int     `json:"stock"`
			Status      string  `json:"status"`
			CreatedBy   uint    `json:"created_by"`
			CreatedAt   string  `json:"created_at"`
			UpdatedAt   string  `json:"updated_at"`
		} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)
	assert.Equal(t, 0, createResp.Code)
	assert.Equal(t, "Test Widget", createResp.Data.Name)
	assert.Equal(t, 29.99, createResp.Data.Price)
	assert.Equal(t, 100, createResp.Data.Stock)
	assert.Equal(t, "on_sale", createResp.Data.Status)
	assert.Equal(t, adminTokens.UserID, createResp.Data.CreatedBy)

	productID := createResp.Data.ID

	// 3. List products (protected, any authenticated user).
	w2 := httpDo(t, productRouter, http.MethodGet, "/api/v1/products", "", adminTokens.AccessToken)
	require.Equal(t, http.StatusOK, w2.Code)

	var listResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Total    int64         `json:"total"`
			Page     int           `json:"page"`
			PageSize int           `json:"page_size"`
			List     []interface{} `json:"list"`
		} `json:"data"`
	}
	err = json.Unmarshal(w2.Body.Bytes(), &listResp)
	require.NoError(t, err)
	assert.Equal(t, 0, listResp.Code)
	assert.GreaterOrEqual(t, listResp.Data.Total, int64(1))
	assert.NotEmpty(t, listResp.Data.List)

	// 4. Get product detail (public, no auth required).
	w3 := httpDo(t, productRouter, http.MethodGet, "/api/v1/products/"+fmt.Sprintf("%d", productID), "", "")
	require.Equal(t, http.StatusOK, w3.Code)

	var getResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ID          uint    `json:"id"`
			Name        string  `json:"name"`
			Description string  `json:"description"`
			Price       float64 `json:"price"`
			Stock       int     `json:"stock"`
			Status      string  `json:"status"`
		} `json:"data"`
	}
	err = json.Unmarshal(w3.Body.Bytes(), &getResp)
	require.NoError(t, err)
	assert.Equal(t, 0, getResp.Code)
	assert.Equal(t, productID, getResp.Data.ID)
	assert.Equal(t, "Test Widget", getResp.Data.Name)
	assert.Equal(t, "A high-quality widget", getResp.Data.Description)
}

func TestProduct_NonAdminCannotCreate(t *testing.T) {
	// A non-admin user (customer role) must get 403 when trying to create a
	// product.

	authRouter := newAuthRouter()
	productRouter := newProductRouter()

	// 1. Register a new customer user.
	customerTokens := registerUser(t, authRouter, "customer@test.com", "CustPass1")
	assert.Equal(t, "customer", string(customerTokens.UserRole))

	// 2. Try to create a product as a customer — should be forbidden.
	w := httpDo(t, productRouter, http.MethodPost, "/api/v1/products",
		`{"name":"Forbidden Item","description":"Should not be created","price":9.99,"stock":1}`,
		customerTokens.AccessToken)
	require.Equal(t, http.StatusForbidden, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, errors.CodeForbidden, resp.Code)
}
