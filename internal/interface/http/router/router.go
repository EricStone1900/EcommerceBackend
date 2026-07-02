package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Router wraps the Gin engine and provides Public/Protected route registration.
type Router struct {
	engine *gin.Engine
	logger *zap.Logger
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

// Public registers a route that is accessible without authentication.
func (r *Router) Public(method, path string, handler gin.HandlerFunc) {
	r.logger.Debug("register public route",
		zap.String("method", method),
		zap.String("path", path),
	)
	r.engine.Handle(method, path, handler)
}

// Protected registers a route that requires authentication.
// The roles parameter specifies which roles are allowed to access this route.
// RBAC enforcement will be implemented in a future phase.
func (r *Router) Protected(method, path string, handler gin.HandlerFunc, roles ...string) {
	r.logger.Debug("register protected route",
		zap.String("method", method),
		zap.String("path", path),
		zap.Strings("roles", roles),
	)
	// TODO(Phase 2): Add AuthMiddleware + RBACMiddleware here
	// r.engine.Handle(method, path, AuthMiddleware(), RBACMiddleware(roles...), handler)
	r.engine.Handle(method, path, handler)
}

// Engine returns the underlying Gin engine.
func (r *Router) Engine() *gin.Engine {
	return r.engine
}
