package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
)

func setupRBACTest(allowedRoles ...entity.Role) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Simulate AuthMiddleware setting the role
	r.Use(func(c *gin.Context) {
		c.Set(string(contextKeyRole), entity.RoleMember)
		c.Next()
	})
	r.Use(RBACMiddleware(allowedRoles...))
	r.GET("/admin", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return r
}

func TestRBACMiddleware_AllowedRole(t *testing.T) {
	router := setupRBACTest(entity.RoleAdmin, entity.RoleMember)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRBACMiddleware_ForbiddenRole(t *testing.T) {
	router := setupRBACTest(entity.RoleAdmin) // only admin allowed, user is member

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp response.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 2003, resp.Code) // CodeForbidden
}

func TestRBACMiddleware_AnyRole(t *testing.T) {
	router := setupRBACTest() // empty = any authenticated user

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRBACMiddleware_MissingRoleInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Don't set role in context
	r.Use(RBACMiddleware(entity.RoleAdmin))
	r.GET("/admin", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
