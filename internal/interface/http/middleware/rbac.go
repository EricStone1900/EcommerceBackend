package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// RBACMiddleware checks that the authenticated user's role is in the allowed list.
// If allowedRoles is empty, any authenticated user is permitted (AnyRole).
func RBACMiddleware(allowedRoles ...entity.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If no roles specified, any authenticated user is allowed
		if len(allowedRoles) == 0 {
			c.Next()
			return
		}

		// Get user role from context (set by AuthMiddleware)
		userRole := GetRole(c)
		if userRole == "" {
			c.AbortWithStatusJSON(403, response.Error(bizerr.CodeForbidden, bizerr.ErrForbidden.Message))
			return
		}

		// Check if role is in the allowed list
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		// Role not allowed
		c.AbortWithStatusJSON(403, response.Error(bizerr.CodeForbidden, bizerr.ErrForbidden.Message))
	}
}
