package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/EricStone1900/ecommerce-backend/internal/infrastructure/config"
)

// NewRedis creates a new Redis client using the provided configuration.
func NewRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}

	if cfg.Password != "" {
		opts.Password = cfg.Password
	}
	if cfg.DB != 0 {
		opts.DB = cfg.DB
	}

	client := redis.NewClient(opts)

	// Verify connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	zap.L().Info("redis connection established",
		zap.String("addr", opts.Addr),
		zap.Int("db", opts.DB),
	)

	return client, nil
}

// HealthCheck verifies the Redis connection by sending a PING command.
func HealthCheck(ctx context.Context, client *redis.Client) error {
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}
	return nil
}
