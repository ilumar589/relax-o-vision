package cache

import (
	"context"
	"container/list"
	"sync"
	"time"
)

// MemoryCache implements Cache interface using in-memory LRU cache
type MemoryCache struct {
	maxSize int
	items   map[string]*cacheItem
	lru     *list.List
	mu      sync.RWMutex
	done    chan struct{}
}

type cacheItem struct {
	key       string
	value     []byte
	expiresAt time.Time
	element   *list.Element
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(maxSize int) *MemoryCache {
	if maxSize <= 0 {
		maxSize = 1000
	}

	cache := &MemoryCache{
		maxSize: maxSize,
		items:   make(map[string]*cacheItem),
		lru:     list.New(),
		done:    make(chan struct{}),
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a value from cache
func (c *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		return nil, nil
	}

	// Check if expired
	if time.Now().After(item.expiresAt) {
		c.removeItem(item)
		return nil, nil
	}

	// Move to front (most recently used)
	c.lru.MoveToFront(item.element)

	return item.value, nil
}

// Set stores a value in cache with TTL
func (c *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key exists, update it
	if item, exists := c.items[key]; exists {
		item.value = value
		item.expiresAt = time.Now().Add(ttl)
		c.lru.MoveToFront(item.element)
		return nil
	}

	// Evict if at capacity
	if c.lru.Len() >= c.maxSize {
		c.evictOldest()
	}

	// Add new item
	item := &cacheItem{
		key:       key,
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	item.element = c.lru.PushFront(item)
	c.items[key] = item

	return nil
}

// Delete removes a key from cache
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.items[key]; exists {
		c.removeItem(item)
	}

	return nil
}

// Exists checks if a key exists in cache
func (c *MemoryCache) Exists(ctx context.Context, key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return false
	}

	// Check if expired
	if time.Now().After(item.expiresAt) {
		return false
	}

	return true
}

// Clear removes all keys from cache
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
	c.lru = list.New()

	return nil
}

// removeItem removes an item from cache (must be called with lock held)
func (c *MemoryCache) removeItem(item *cacheItem) {
	c.lru.Remove(item.element)
	delete(c.items, item.key)
}

// evictOldest removes the least recently used item (must be called with lock held)
func (c *MemoryCache) evictOldest() {
	elem := c.lru.Back()
	if elem != nil {
		item := elem.Value.(*cacheItem)
		c.removeItem(item)
	}
}

// cleanupExpired periodically removes expired items
func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for key, item := range c.items {
				if now.After(item.expiresAt) {
					c.lru.Remove(item.element)
					delete(c.items, key)
				}
			}
			c.mu.Unlock()
		}
	}
}

// Close stops the cleanup goroutine
func (c *MemoryCache) Close() error {
	close(c.done)
	return nil
}
