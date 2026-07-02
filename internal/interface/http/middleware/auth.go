package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/pkg/jwt"
	"github.com/EricStone1900/ecommerce-backend/pkg/response"
	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// Context keys for storing auth info in Gin context.
type contextKey string

const (
	contextKeyUserID contextKey = "user_id"
	contextKeyRole   contextKey = "role"
)

// AuthMiddleware validates the JWT access token from the Authorization header.
// It extracts user_id and role and stores them in the Gin context.
// On failure, it returns 401 and aborts the request.
func AuthMiddleware(jwtService *jwt.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			abortWithError(c, bizerr.ErrUnauthorized)
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			abortWithError(c, bizerr.ErrUnauthorized)
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			abortWithError(c, err)
			return
		}

		// Ensure this is an access token, not a refresh token
		if !jwtService.IsAccessToken(claims) {
			abortWithError(c, bizerr.ErrUnauthorized)
			return
		}

		// Store auth info in context
		c.Set(string(contextKeyUserID), claims.UserID)
		c.Set(string(contextKeyRole), claims.Role)

		c.Next()
	}
}

// GetUserID retrieves the authenticated user's ID from the Gin context.
func GetUserID(c *gin.Context) uint {
	uid, exists := c.Get(string(contextKeyUserID))
	if !exists {
		return 0
	}
	return uid.(uint)
}

// GetRole retrieves the authenticated user's role from the Gin context.
func GetRole(c *gin.Context) entity.Role {
	role, exists := c.Get(string(contextKeyRole))
	if !exists {
		return ""
	}
	return role.(entity.Role)
}

// abortWithError sends a JSON error response and aborts the request.
func abortWithError(c *gin.Context, err error) {
	// Try to extract our business error type
	if bizErr, ok := err.(*bizerr.Error); ok {
		c.AbortWithStatusJSON(bizErr.HTTPCode, response.Error(bizErr.Code, bizErr.Message))
		return
	}
	// Fallback: generic internal error
	c.AbortWithStatusJSON(500, response.Error(bizerr.CodeInternalError, "internal server error"))
}
