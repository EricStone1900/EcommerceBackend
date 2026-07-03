package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HealthDependencies holds the dependencies needed by the health check handler.
type HealthDependencies struct {
	DB      *gorm.DB
	Redis   *redis.Client
	Logger  *zap.Logger
	StartAt time.Time
}

// HealthResponse is the JSON response for the health check endpoint.
type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Redis    string `json:"redis"`
	Uptime   string `json:"uptime"`
}

// @Summary      健康检查
// @Description  检查服务运行状态，包括数据库和 Redis 连接
// @Tags         系统管理
// @Produce      json
// @Success      200 {object} handler.HealthResponse
// @Failure      503 {object} handler.HealthResponse
// @Router       /health [get]
// NewHealthHandler creates a health check handler that pings DB and Redis.
func NewHealthHandler(deps *HealthDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		resp := HealthResponse{
			Status:   "ok",
			Database: "connected",
			Redis:    "connected",
			Uptime:   time.Since(deps.StartAt).Round(time.Second).String(),
		}

		overallOK := true

		// Ping database
		if deps.DB != nil {
			sqlDB, err := deps.DB.DB()
			if err != nil {
				deps.Logger.Warn("health check: failed to get sql.DB", zap.Error(err))
				resp.Database = "disconnected"
				overallOK = false
			} else {
				ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
				defer cancel()
				if err := sqlDB.PingContext(ctx); err != nil {
					deps.Logger.Warn("health check: database ping failed", zap.Error(err))
					resp.Database = "disconnected"
					overallOK = false
				}
			}
		} else {
			resp.Database = "not configured"
		}

		// Ping Redis
		if deps.Redis != nil {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()
			if err := deps.Redis.Ping(ctx).Err(); err != nil {
				deps.Logger.Warn("health check: redis ping failed", zap.Error(err))
				resp.Redis = "disconnected"
				overallOK = false
			}
		} else {
			resp.Redis = "not configured"
		}

		if !overallOK {
			c.JSON(http.StatusServiceUnavailable, resp)
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
