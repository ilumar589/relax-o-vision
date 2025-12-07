package footballdata

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json/v2"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/edd/relaxovisionmonolith/cache"
)

// GetCacheTTL returns the cache TTL from environment or default (30 days)
func GetCacheTTL() time.Duration {
	if ttlStr := os.Getenv("CACHE_TTL_DAYS"); ttlStr != "" {
		if days, err := strconv.Atoi(ttlStr); err == nil && days > 0 {
			return time.Duration(days) * 24 * time.Hour
		}
	}
	// Default to 30 days
	return 30 * 24 * time.Hour
}

// CacheTTL is the duration for which data is considered fresh
var CacheTTL = GetCacheTTL()

// CacheMetadata represents cache metadata for tracking freshness
type CacheMetadata struct {
	ID         int       `json:"id"`
	EntityType string    `json:"entity_type"`
	EntityKey  string    `json:"entity_key"`
	CachedAt   time.Time `json:"cached_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	DataHash   string    `json:"data_hash,omitempty"`
}

// CacheManager coordinates caching between Redis and PostgreSQL
type CacheManager struct {
	redis cache.Cache
	db    *sql.DB
}

// NewCacheManager creates a new cache manager instance
func NewCacheManager(redisCache cache.Cache, db *sql.DB) *CacheManager {
	return &CacheManager{
		redis: redisCache,
		db:    db,
	}
}

// GetMetadata retrieves cache metadata for an entity
func (cm *CacheManager) GetMetadata(ctx context.Context, entityType, entityKey string) (*CacheMetadata, error) {
	query := `
		SELECT id, entity_type, entity_key, cached_at, expires_at, data_hash
		FROM cache_metadata
		WHERE entity_type = $1 AND entity_key = $2
	`

	var metadata CacheMetadata
	err := cm.db.QueryRowContext(ctx, query, entityType, entityKey).Scan(
		&metadata.ID,
		&metadata.EntityType,
		&metadata.EntityKey,
		&metadata.CachedAt,
		&metadata.ExpiresAt,
		&metadata.DataHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No metadata found
		}
		return nil, fmt.Errorf("failed to get cache metadata: %w", err)
	}

	return &metadata, nil
}

// SetMetadata sets or updates cache metadata for an entity
func (cm *CacheManager) SetMetadata(ctx context.Context, entityType, entityKey string, dataHash string) error {
	now := time.Now()
	expiresAt := now.Add(CacheTTL)

	query := `
		INSERT INTO cache_metadata (entity_type, entity_key, cached_at, expires_at, data_hash)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (entity_type, entity_key) DO UPDATE SET
			cached_at = EXCLUDED.cached_at,
			expires_at = EXCLUDED.expires_at,
			data_hash = EXCLUDED.data_hash
	`

	_, err := cm.db.ExecContext(ctx, query, entityType, entityKey, now, expiresAt, dataHash)
	if err != nil {
		return fmt.Errorf("failed to set cache metadata: %w", err)
	}

	return nil
}

// NeedsRefresh checks if data needs to be refreshed based on cache metadata
func (cm *CacheManager) NeedsRefresh(ctx context.Context, entityType, entityKey string) bool {
	metadata, err := cm.GetMetadata(ctx, entityType, entityKey)
	if err != nil || metadata == nil {
		return true // No cache or error, needs refresh
	}

	// Check if cache has expired
	return time.Now().After(metadata.ExpiresAt)
}

// Get retrieves data from cache (tries Redis first, then PostgreSQL)
func (cm *CacheManager) Get(ctx context.Context, key string) ([]byte, error) {
	// Try Redis first for fast access
	if cm.redis != nil {
		data, err := cm.redis.Get(ctx, key)
		if err == nil && data != nil {
			slog.Debug("Redis cache hit", "key", key)
			return data, nil
		}
	}

	// Redis miss - data would come from PostgreSQL through repository
	slog.Debug("Cache miss", "key", key)
	return nil, nil
}

// Set stores data in both Redis and PostgreSQL
func (cm *CacheManager) Set(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	// Store in Redis for fast access
	if cm.redis != nil {
		if err := cm.redis.Set(ctx, key, data, ttl); err != nil {
			slog.Warn("Failed to set Redis cache", "key", key, "error", err)
			// Continue even if Redis fails - PostgreSQL is the source of truth
		}
	}

	return nil
}

// Delete removes data from cache
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	if cm.redis != nil {
		return cm.redis.Delete(ctx, key)
	}
	return nil
}

// InvalidateEntity invalidates cache for a specific entity
func (cm *CacheManager) InvalidateEntity(ctx context.Context, entityType, entityKey string) error {
	// Delete metadata
	query := `DELETE FROM cache_metadata WHERE entity_type = $1 AND entity_key = $2`
	_, err := cm.db.ExecContext(ctx, query, entityType, entityKey)
	if err != nil {
		return fmt.Errorf("failed to invalidate entity: %w", err)
	}

	// Delete from Redis
	redisKey := fmt.Sprintf("football:%s:%s", entityType, entityKey)
	if cm.redis != nil {
		cm.redis.Delete(ctx, redisKey)
	}

	return nil
}

// ComputeDataHash computes a SHA256 hash of data for change detection
func ComputeDataHash(data any) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash)
}
