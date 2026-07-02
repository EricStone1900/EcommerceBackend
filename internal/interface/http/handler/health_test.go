package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHealthHandler_NoDependencies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	deps := &HealthDependencies{
		DB:      nil,
		Redis:   nil,
		Logger:  logger,
		StartAt: time.Now(),
	}

	router := gin.New()
	router.GET("/health", NewHealthHandler(deps))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, "not configured", resp.Database)
	assert.Equal(t, "not configured", resp.Redis)
	assert.NotEmpty(t, resp.Uptime)
}

func TestHealthHandler_WithDBFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	deps := &HealthDependencies{
		DB:      nil,   // nil DB simulates failure
		Redis:   nil,   // nil Redis
		Logger:  logger,
		StartAt: time.Now(),
	}

	router := gin.New()
	router.GET("/health", NewHealthHandler(deps))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "not configured", resp.Database)
	assert.Equal(t, "not configured", resp.Redis)
}

func TestHealthHandler_ResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	deps := &HealthDependencies{
		DB:      nil,
		Redis:   nil,
		Logger:  logger,
		StartAt: time.Now(),
	}

	router := gin.New()
	router.GET("/health", NewHealthHandler(deps))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp HealthResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify all expected fields are present
	assert.Contains(t, w.Body.String(), "status")
	assert.Contains(t, w.Body.String(), "database")
	assert.Contains(t, w.Body.String(), "redis")
	assert.Contains(t, w.Body.String(), "uptime")
}
