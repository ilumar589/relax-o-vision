package footballdata

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/edd/relaxovisionmonolith/cache"
)

// CachedClient wraps the football data client with caching
type CachedClient struct {
	client *Client
	cache  cache.Cache
}

// CacheTTL defines cache TTL by data type
var CacheTTL = map[string]time.Duration{
	"competitions": 24 * time.Hour,
	"teams":        12 * time.Hour,
	"matches":      5 * time.Minute, // Short TTL for live data
	"standings":    15 * time.Minute,
	"head2head":    1 * time.Hour,
}

// NewCachedClient creates a new cached client
func NewCachedClient(apiKey string, cacheImpl cache.Cache) *CachedClient {
	return &CachedClient{
		client: NewClient(apiKey),
		cache:  cacheImpl,
	}
}

// GetCompetition gets a competition with caching
func (c *CachedClient) GetCompetition(ctx context.Context, code string) (*Competition, error) {
	cacheKey := fmt.Sprintf("competition:%s", code)

	// Try cache first
	cached, err := c.cache.Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var comp Competition
		if err := json.Unmarshal(cached, &comp); err == nil {
			slog.Debug("Cache hit", "key", cacheKey)
			return &comp, nil
		}
	}

	// Cache miss - fetch from API
	slog.Debug("Cache miss", "key", cacheKey)
	comp, err := c.client.GetCompetition(ctx, code)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if data, err := json.Marshal(comp); err == nil {
		c.cache.Set(ctx, cacheKey, data, CacheTTL["competitions"])
	}

	return comp, nil
}

// GetTeam gets a team with caching
func (c *CachedClient) GetTeam(ctx context.Context, id int) (*Team, error) {
	cacheKey := fmt.Sprintf("team:%d", id)

	// Try cache first
	cached, err := c.cache.Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var team Team
		if err := json.Unmarshal(cached, &team); err == nil {
			slog.Debug("Cache hit", "key", cacheKey)
			return &team, nil
		}
	}

	// Cache miss - fetch from API
	slog.Debug("Cache miss", "key", cacheKey)
	team, err := c.client.GetTeam(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if data, err := json.Marshal(team); err == nil {
		c.cache.Set(ctx, cacheKey, data, CacheTTL["teams"])
	}

	return team, nil
}

// GetMatches gets matches with caching
func (c *CachedClient) GetMatches(ctx context.Context, competitionCode string) ([]Match, error) {
	cacheKey := fmt.Sprintf("matches:%s", competitionCode)

	// Try cache first
	cached, err := c.cache.Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var matches []Match
		if err := json.Unmarshal(cached, &matches); err == nil {
			slog.Debug("Cache hit", "key", cacheKey)
			return matches, nil
		}
	}

	// Cache miss - fetch from API
	slog.Debug("Cache miss", "key", cacheKey)
	matches, err := c.client.GetMatches(ctx, competitionCode)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if data, err := json.Marshal(matches); err == nil {
		c.cache.Set(ctx, cacheKey, data, CacheTTL["matches"])
	}

	return matches, nil
}

// GetStandings gets standings with caching
func (c *CachedClient) GetStandings(ctx context.Context, competitionCode string) (*Standing, error) {
	cacheKey := fmt.Sprintf("standings:%s", competitionCode)

	// Try cache first
	cached, err := c.cache.Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var standing Standing
		if err := json.Unmarshal(cached, &standing); err == nil {
			slog.Debug("Cache hit", "key", cacheKey)
			return &standing, nil
		}
	}

	// Cache miss - fetch from API
	slog.Debug("Cache miss", "key", cacheKey)
	standing, err := c.client.GetStandings(ctx, competitionCode)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if data, err := json.Marshal(standing); err == nil {
		c.cache.Set(ctx, cacheKey, data, CacheTTL["standings"])
	}

	return standing, nil
}

// InvalidateMatch invalidates cache for a specific match (e.g., when it finishes)
func (c *CachedClient) InvalidateMatch(ctx context.Context, competitionCode string) error {
	cacheKey := fmt.Sprintf("matches:%s", competitionCode)
	return c.cache.Delete(ctx, cacheKey)
}

// InvalidateCompetition invalidates cache for a competition
func (c *CachedClient) InvalidateCompetition(ctx context.Context, code string) error {
	cacheKey := fmt.Sprintf("competition:%s", code)
	return c.cache.Delete(ctx, cacheKey)
}

// ClearCache clears all cache (admin operation)
func (c *CachedClient) ClearCache(ctx context.Context) error {
	return c.cache.Clear(ctx)
}
