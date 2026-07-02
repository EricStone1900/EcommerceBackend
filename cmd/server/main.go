package main

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/container"
	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/handler"
	"github.com/EricStone1900/ecommerce-backend/internal/interface/http/router"
)

const configPath = "configs/config.local.yaml"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: server <command>")
		fmt.Println("Commands:")
		fmt.Println("  serve    Start the HTTP server")
		fmt.Println("  migrate  Run database migrations (placeholder)")
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

	// Register health check
	healthHandler := handler.NewHealthHandler(&handler.HealthDependencies{
		DB:      c.DB,
		Redis:   c.Redis,
		Logger:  c.Logger,
		StartAt: time.Now(),
	})
	r.Public("GET", "/health", healthHandler)

	// Start server
	addr := fmt.Sprintf(":%d", c.Config.Server.Port)
	c.Logger.Info("listening", zap.String("address", addr))
	if err := r.Engine().Run(addr); err != nil {
		c.Logger.Fatal("server failed", zap.Error(err))
	}
}

func runMigrate() {
	fmt.Println("Database migration placeholder")
	fmt.Println("Migrations will be implemented in a future phase")
	os.Exit(0)
}
