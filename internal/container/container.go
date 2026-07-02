package container

import (
	"fmt"

	"github.com/gin-gonic/gin"
	goRedis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/EricStone1900/ecommerce-backend/internal/infrastructure/cache/redis"
	gormdb "github.com/EricStone1900/ecommerce-backend/internal/infrastructure/persistence/gorm"
	"github.com/EricStone1900/ecommerce-backend/internal/infrastructure/config"
	"github.com/EricStone1900/ecommerce-backend/internal/infrastructure/log"
	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/handler"
	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/middleware"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/auth"
	"github.com/EricStone1900/ecommerce-backend/pkg/jwt"
)

// Container holds all initialized dependencies for the application.
type Container struct {
	Config *config.Config
	Logger *zap.Logger
	DB     *gorm.DB
	Redis  *goRedis.Client

	// Phase 2 — Auth dependencies
	UserRepo       *gormdb.UserRepository
	TokenStore     *redis.TokenStore
	JWTService     *jwt.JWTService
	AuthUseCase    *auth.AuthUseCase
	AuthHandler    *handler.AuthHandler
	AuthMiddleware gin.HandlerFunc
}

// NewContainer initializes all dependencies in order.
func NewContainer(cfgPath string) (*Container, error) {
	// 1. Load configuration
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("container: failed to load config: %w", err)
	}

	// 2. Initialize logger
	logger, err := log.InitLogger(cfg.Server.Env)
	if err != nil {
		return nil, fmt.Errorf("container: failed to init logger: %w", err)
	}

	// 3. Initialize database
	db, err := gormdb.NewDB(&cfg.Database, cfg.Server.Env)
	if err != nil {
		logger.Error("failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("container: failed to connect to database: %w", err)
	}

	// 4. Initialize Redis
	rdb, err := redis.NewRedis(&cfg.Redis)
	if err != nil {
		logger.Error("failed to connect to redis", zap.Error(err))
		return nil, fmt.Errorf("container: failed to connect to redis: %w", err)
	}

	// 5. Initialize repositories and services
	userRepo := gormdb.NewUserRepository(db)
	tokenStore := redis.NewTokenStore(rdb)
	jwtService := jwt.NewJWTService(cfg.JWT.Secret, cfg.JWT.AccessExpire, cfg.JWT.RefreshExpire)

	// 6. Initialize use cases
	authUseCase := auth.NewAuthUseCase(
		userRepo,
		tokenStore,
		jwtService,
		logger,
		cfg.JWT.AccessExpire,
		cfg.JWT.RefreshExpire,
	)

	// 7. Initialize HTTP layer
	authHandler := handler.NewAuthHandler(authUseCase)
	authMiddleware := middleware.AuthMiddleware(jwtService)

	logger.Info("container initialized",
		zap.String("env", cfg.Server.Env),
		zap.Int("port", cfg.Server.Port),
	)

	return &Container{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Redis:  rdb,

		UserRepo:       userRepo,
		TokenStore:     tokenStore,
		JWTService:     jwtService,
		AuthUseCase:    authUseCase,
		AuthHandler:    authHandler,
		AuthMiddleware: authMiddleware,
	}, nil
}

// Close gracefully shuts down all resources.
func (c *Container) Close() {
	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			c.Logger.Error("failed to close redis", zap.Error(err))
		}
	}

	if c.DB != nil {
		sqlDB, err := c.DB.DB()
		if err != nil {
			c.Logger.Error("failed to get sql.DB for closing", zap.Error(err))
		} else {
			if err := sqlDB.Close(); err != nil {
				c.Logger.Error("failed to close database", zap.Error(err))
			}
		}
	}

	if c.Logger != nil {
		_ = c.Logger.Sync()
	}
}
