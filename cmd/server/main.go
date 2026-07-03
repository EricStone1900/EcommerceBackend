// @title           E-Commerce Backend API
// @version         1.0
// @description     电商后台管理系统 API
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer token authentication

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/EricStone1900/ecommerce-backend/docs"

	"github.com/EricStone1900/ecommerce-backend/internal/container"
	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/handler"
	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/router"
)

const configPath = "configs/config.local.yaml"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: server <command>")
		fmt.Println("Commands:")
		fmt.Println("  serve    Start the HTTP server")
		fmt.Println("  migrate  Run database migrations")
		fmt.Println("  migrate  --down  Rollback database migrations")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "serve":
		runServe()
	case "migrate":
		runMigrate()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func runServe() {
	cfgPath := configPath
	if envPath := os.Getenv("APP_CONFIG_PATH"); envPath != "" {
		cfgPath = envPath
	}

	c, err := container.NewContainer(cfgPath)
	if err != nil {
		fmt.Printf("Failed to initialize container: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	c.Logger.Info("starting server",
		zap.Int("port", c.Config.Server.Port),
		zap.String("env", c.Config.Server.Env),
	)

	// Setup router
	r := router.NewRouter(c.Logger)

	// Set auth middleware for protected routes
	r.SetAuthMiddleware(c.AuthMiddleware)

	// Register health check
	healthHandler := handler.NewHealthHandler(&handler.HealthDependencies{
		DB:      c.DB,
		Redis:   c.Redis,
		Logger:  c.Logger,
		StartAt: time.Now(),
	})
	r.Public("GET", "/health", healthHandler)

	// Register auth routes
	r.Public("POST", "/api/v1/auth/register", c.AuthHandler.Register)
	r.Public("POST", "/api/v1/auth/login", c.AuthHandler.Login)
	r.Protected("POST", "/api/v1/auth/logout", c.AuthHandler.Logout)
	r.Protected("POST", "/api/v1/auth/refresh", c.AuthHandler.Refresh)
	r.Protected("GET", "/api/v1/auth/me", c.AuthHandler.Me)

	// Register product routes
	r.Public("GET", "/api/v1/products/:id", c.ProductHandler.GetDetail)
	r.Protected("GET", "/api/v1/products", c.ProductHandler.List)
	r.Protected("POST", "/api/v1/products", c.ProductHandler.Create, entity.RoleAdmin)
	r.Protected("PUT", "/api/v1/products/:id", c.ProductHandler.Update, entity.RoleAdmin)
	r.Protected("DELETE", "/api/v1/products/:id", c.ProductHandler.Delete, entity.RoleAdmin)

	// Register upload route (Phase 4)
	r.Protected("POST", "/api/v1/upload", c.UploadHandler.Upload)

	// Serve uploaded files statically
	r.Engine().Static("/uploads", c.Config.Storage.Local.BasePath)

	// Register push notification routes (Phase 5)
	r.Protected("POST", "/api/v1/push/token", c.PushHandler.RegisterToken)
	r.Protected("DELETE", "/api/v1/push/token", c.PushHandler.DeleteToken)
	r.Protected("POST", "/api/v1/push/test", c.PushHandler.SendTest)

	// Swagger UI — only enabled in development mode
	if c.Config.Server.Env == "development" {
		r.Engine().GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		c.Logger.Info("swagger UI enabled at /swagger/index.html")
	}

	// Start server
	addr := fmt.Sprintf(":%d", c.Config.Server.Port)
	c.Logger.Info("listening", zap.String("address", addr))
	if err := r.Engine().Run(addr); err != nil {
		c.Logger.Fatal("server failed", zap.Error(err))
	}
}

func runMigrate() {
	cfgPath := configPath
	if envPath := os.Getenv("APP_CONFIG_PATH"); envPath != "" {
		cfgPath = envPath
	}

	c, err := container.NewContainer(cfgPath)
	if err != nil {
		fmt.Printf("Failed to initialize container: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	isDown := len(os.Args) > 2 && os.Args[2] == "--down"

	migrationsDir := "migrations"
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		c.Logger.Fatal("failed to read migrations directory", zap.Error(err))
	}

	// Filter and sort migration files
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if isDown && strings.HasSuffix(name, ".down.sql") {
			files = append(files, name)
		} else if !isDown && strings.HasSuffix(name, ".up.sql") {
			files = append(files, name)
		}
	}
	sort.Strings(files)

	if isDown {
		// Reverse order for down migrations
		sort.Sort(sort.Reverse(sort.StringSlice(files)))
	}

	if len(files) == 0 {
		direction := "up"
		if isDown {
			direction = "down"
		}
		c.Logger.Info("no migration files found", zap.String("direction", direction))
		return
	}

	direction := "up"
	if isDown {
		direction = "down"
	}
	c.Logger.Info("running migrations",
		zap.String("direction", direction),
		zap.Int("count", len(files)),
	)

	for _, file := range files {
		filePath := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			c.Logger.Fatal("failed to read migration file",
				zap.String("file", file),
				zap.Error(err),
			)
		}

		sql := strings.TrimSpace(string(content))
		if sql == "" {
			c.Logger.Info("skipping empty migration file", zap.String("file", file))
			continue
		}

		c.Logger.Info("executing migration", zap.String("file", file))

		// Split by semicolons to handle multiple statements
		statements := strings.Split(sql, ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if err := c.DB.Exec(stmt).Error; err != nil {
				c.Logger.Fatal("migration statement failed",
					zap.String("file", file),
					zap.String("statement", stmt[:min(len(stmt), 80)]),
					zap.Error(err),
				)
			}
		}

		c.Logger.Info("migration applied successfully", zap.String("file", file))
	}

	c.Logger.Info("all migrations completed successfully",
		zap.String("direction", direction),
		zap.Int("count", len(files)),
	)
}
