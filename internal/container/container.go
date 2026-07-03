package container

import (
	"fmt"

	"github.com/gin-gonic/gin"
	goRedis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/port"
	"github.com/EricStone1900/ecommerce-backend/internal/infrastructure/cache/redis"
	"github.com/EricStone1900/ecommerce-backend/internal/infrastructure/config"
	"github.com/EricStone1900/ecommerce-backend/internal/infrastructure/event"
	"github.com/EricStone1900/ecommerce-backend/internal/infrastructure/log"
	gormdb "github.com/EricStone1900/ecommerce-backend/internal/infrastructure/persistence/gorm"
	"github.com/EricStone1900/ecommerce-backend/internal/infrastructure/storage/local"
	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/handler"
	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/middleware"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/assistant"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/auth"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/fileprocessing"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/product"
	"github.com/EricStone1900/ecommerce-backend/internal/usecase/upload"
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

	// Phase 3 — Product dependencies
	ProductRepo    *gormdb.ProductRepository
	ProductUseCase *product.ProductUseCase
	ProductHandler *handler.ProductHandler

	// Phase 4 — File upload dependencies
	EventBus          *event.EventBus
	FileRepo          *gormdb.FileRepository
	Storage           port.Storage
	UploadUseCase     *upload.UploadUseCase
	UploadHandler     *handler.UploadHandler

	// Phase 4 — Assistant stub
	Assistant port.AssistantPort
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
	productRepo := gormdb.NewProductRepository(db)

	// 6. Initialize use cases
	authUseCase := auth.NewAuthUseCase(
		userRepo,
		tokenStore,
		jwtService,
		logger,
		cfg.JWT.AccessExpire,
		cfg.JWT.RefreshExpire,
	)
	productUseCase := product.NewProductUseCase(productRepo, logger)

	// 7. Phase 4 — Initialize event bus, file storage, and upload use case
	eventBus := event.NewEventBus(logger)
	fileRepo := gormdb.NewFileRepository(db)

	var storage port.Storage
	switch cfg.Storage.Driver {
	case "s3":
		logger.Warn("S3 storage not yet implemented, falling back to local")
		fallthrough
	default:
		s, err := local.NewStorage(cfg.Storage.Local.BasePath, logger)
		if err != nil {
			return nil, fmt.Errorf("container: failed to initialize local storage: %w", err)
		}
		storage = s
	}

	uploadUseCase := upload.NewUploadUseCase(fileRepo, storage, eventBus, logger)

	// Register file processing event subscriber
	fileProcHandler := fileprocessing.NewHandler(fileRepo, logger)
	eventBus.Subscribe("file.uploaded", fileProcHandler.HandleFileUploaded)

	// Phase 4 — Initialize assistant stub
	ast := assistant.NewMockAssistant(logger)

	// 8. Initialize HTTP layer
	authHandler := handler.NewAuthHandler(authUseCase)
	authMiddleware := middleware.AuthMiddleware(jwtService)
	productHandler := handler.NewProductHandler(productUseCase)
	uploadHandler := handler.NewUploadHandler(uploadUseCase)

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

		ProductRepo:    productRepo,
		ProductUseCase: productUseCase,
		ProductHandler: productHandler,

		EventBus:      eventBus,
		FileRepo:      fileRepo,
		Storage:       storage,
		UploadUseCase: uploadUseCase,
		UploadHandler: uploadHandler,

		Assistant: ast,
	}, nil
}

// Close gracefully shuts down all resources.
func (c *Container) Close() {
	if c.EventBus != nil {
		c.EventBus.Close()
	}

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
