package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EricStone1900/ecommerce-backend/pkg/jwt"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
)

func newTestJWTService() *jwt.JWTService {
	return jwt.NewJWTService("test_secret_key_for_mw", 15*time.Minute, 168*time.Hour)
}

func setupAuthTest(service *jwt.JWTService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthMiddleware(service))
	r.GET("/protected", func(c *gin.Context) {
		userID := GetUserID(c)
		role := GetRole(c)
		c.JSON(200, gin.H{"user_id": userID, "role": role})
	})
	return r
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	svc := newTestJWTService()
	router := setupAuthTest(svc)

	token, err := svc.GenerateAccessToken(42, "admin")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, float64(42), body["user_id"])
	assert.Equal(t, "admin", body["role"])
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	svc := newTestJWTService()
	router := setupAuthTest(svc)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2001, resp.Code)
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	svc := newTestJWTService()
	router := setupAuthTest(svc)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	// Create a service with instant expiry
	svc := jwt.NewJWTService("test_secret_key_for_mw", -1*time.Minute, 168*time.Hour)
	router := setupAuthTest(svc)

	token, err := svc.GenerateAccessToken(1, "customer")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2002, resp.Code) // CodeTokenExpired
}

func TestAuthMiddleware_RefreshTokenRejected(t *testing.T) {
	svc := newTestJWTService()
	router := setupAuthTest(svc)

	refreshToken, _, err := svc.GenerateRefreshToken(1, "customer")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
