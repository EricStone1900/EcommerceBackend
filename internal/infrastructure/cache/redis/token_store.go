package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const refreshTokenKeyPrefix = "refresh_token"

// TokenStore implements port.TokenStore using Redis.
type TokenStore struct {
	client *redis.Client
}

// NewTokenStore creates a new Redis-backed TokenStore.
func NewTokenStore(client *redis.Client) *TokenStore {
	return &TokenStore{client: client}
}

// refreshTokenKey builds the Redis key for a refresh token.
func refreshTokenKey(userID uint, tokenID string) string {
	return fmt.Sprintf("%s:%d:%s", refreshTokenKeyPrefix, userID, tokenID)
}

// SaveRefreshToken stores a refresh token ID with the given TTL.
func (s *TokenStore) SaveRefreshToken(ctx context.Context, userID uint, tokenID string, expiration time.Duration) error {
	key := refreshTokenKey(userID, tokenID)
	if err := s.client.Set(ctx, key, "1", expiration).Err(); err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}
	return nil
}

// ValidateRefreshToken checks whether a refresh token ID exists and is still valid.
func (s *TokenStore) ValidateRefreshToken(ctx context.Context, userID uint, tokenID string) (bool, error) {
	key := refreshTokenKey(userID, tokenID)
	n, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to validate refresh token: %w", err)
	}
	return n > 0, nil
}

// DeleteRefreshToken removes a refresh token ID from the store (revocation).
func (s *TokenStore) DeleteRefreshToken(ctx context.Context, userID uint, tokenID string) error {
	key := refreshTokenKey(userID, tokenID)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}
