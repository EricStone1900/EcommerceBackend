package redis

import (
	"testing"

	"github.com/EricStone1900/ecommerce-backend/internal/infrastructure/config"
)

func TestNewRedisInvalidURL(t *testing.T) {
	cfg := &config.RedisConfig{
		URL:      "not-a-valid-redis-url",
		Password: "",
		DB:       0,
	}

	_, err := NewRedis(cfg)
	if err == nil {
		t.Error("NewRedis with invalid URL should return an error")
	}
}

func TestNewRedisEmptyURL(t *testing.T) {
	cfg := &config.RedisConfig{
		URL: "",
	}

	_, err := NewRedis(cfg)
	if err == nil {
		t.Error("NewRedis with empty URL should return an error")
	}
}

func TestRedisConfigDefaults(t *testing.T) {
	cfg := &config.RedisConfig{}

	if cfg.URL != "" {
		t.Errorf("default URL should be empty, got %q", cfg.URL)
	}
	if cfg.DB != 0 {
		t.Errorf("default DB should be 0, got %d", cfg.DB)
	}
}
