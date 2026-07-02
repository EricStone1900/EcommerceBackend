package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/middleware"
)

// Router wraps the Gin engine and provides Public/Protected route registration.
type Router struct {
	engine         *gin.Engine
	logger         *zap.Logger
	authMiddleware gin.HandlerFunc
}

// NewRouter creates a new Router with the given logger.
func NewRouter(logger *zap.Logger) *Router {
	engine := gin.New()
	engine.Use(gin.Recovery())

	// Request logging middleware
	engine.Use(func(c *gin.Context) {
		logger.Info("incoming request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.ClientIP()),
		)
		c.Next()
	})

	return &Router{
		engine: engine,
		logger: logger,
	}
}

// SetAuthMiddleware sets the JWT authentication middleware to be used
// by all Protected routes. Must be called before registering Protected routes.
func (r *Router) SetAuthMiddleware(mw gin.HandlerFunc) {
	r.authMiddleware = mw
}

// Public registers a route that is accessible without authentication.
func (r *Router) Public(method, path string, handler gin.HandlerFunc) {
	r.logger.Debug("register public route",
		zap.String("method", method),
		zap.String("path", path),
	)
	r.engine.Handle(method, path, handler)
}

// Protected registers a route that requires authentication and role-based access.
// The roles parameter specifies which roles are allowed. If empty, any authenticated
// user is permitted (AnyRole).
func (r *Router) Protected(method, path string, handler gin.HandlerFunc, roles ...entity.Role) {
	r.logger.Debug("register protected route",
		zap.String("method", method),
		zap.String("path", path),
		zap.Any("roles", roles),
	)

	// Build middleware chain: AuthMiddleware → RBACMiddleware → handler
	handlers := []gin.HandlerFunc{
		r.authMiddleware,
		middleware.RBACMiddleware(roles...),
		handler,
	}
	r.engine.Handle(method, path, handlers...)
}

// Engine returns the underlying Gin engine.
func (r *Router) Engine() *gin.Engine {
	return r.engine
}
