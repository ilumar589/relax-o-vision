package cache

import (
	"context"
	"time"
)

// Cache interface defines caching operations
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) bool
	Clear(ctx context.Context) error
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Type string // "redis" or "memory"
	
	// Redis config
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	
	// Memory config
	MaxSize int // Maximum number of items in memory cache
}

// NewCache creates a new cache based on configuration
func NewCache(config CacheConfig) (Cache, error) {
	switch config.Type {
	case "redis":
		return NewRedisCache(config.RedisAddr, config.RedisPassword, config.RedisDB)
	case "memory":
		return NewMemoryCache(config.MaxSize), nil
	default:
		// Default to memory cache
		return NewMemoryCache(1000), nil
	}
}
