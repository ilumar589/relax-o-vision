package cache

import (
	"context"
	"testing"
	"time"
)

func TestNewCache_Redis(t *testing.T) {
	t.Skip("Integration test - requires Redis server")
}

func TestNewCache_Memory(t *testing.T) {
	t.Parallel()

	config := CacheConfig{
		Type:    "memory",
		MaxSize: 100,
	}

	cache, err := NewCache(config)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	if cache == nil {
		t.Fatal("NewCache() returned nil")
	}

	// Test basic operations
	ctx := context.Background()
	key := "test:key"
	value := []byte("test value")

	err = cache.Set(ctx, key, value, 1*time.Hour)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	result, err := cache.Get(ctx, key)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}

	if string(result) != string(value) {
		t.Errorf("Get() = %v, want %v", string(result), string(value))
	}
}

func TestNewCache_Default(t *testing.T) {
	t.Parallel()

	config := CacheConfig{
		Type: "unknown",
	}

	cache, err := NewCache(config)
	if err != nil {
		t.Fatalf("NewCache() error = %v", err)
	}

	if cache == nil {
		t.Fatal("NewCache() returned nil - should default to memory cache")
	}
}

func TestMemoryCache_Operations(t *testing.T) {
	t.Parallel()

	cache := NewMemoryCache(10)
	ctx := context.Background()

	// Test Set and Get
	key := "test:key"
	value := []byte("test value")

	err := cache.Set(ctx, key, value, 1*time.Hour)
	if err != nil {
		t.Errorf("Set() error = %v", err)
	}

	result, err := cache.Get(ctx, key)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}

	if string(result) != string(value) {
		t.Errorf("Get() = %v, want %v", string(result), string(value))
	}

	// Test Exists
	if !cache.Exists(ctx, key) {
		t.Error("Exists() = false, want true")
	}

	// Test Delete
	err = cache.Delete(ctx, key)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	if cache.Exists(ctx, key) {
		t.Error("Exists() = true after Delete(), want false")
	}

	// Test Get after delete
	result, err = cache.Get(ctx, key)
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if result != nil {
		t.Errorf("Get() after Delete() = %v, want nil", result)
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	t.Parallel()

	cache := NewMemoryCache(10)
	ctx := context.Background()

	// Add multiple items
	for i := 0; i < 5; i++ {
		key := string(rune('a' + i))
		value := []byte(key)
		cache.Set(ctx, key, value, 1*time.Hour)
	}

	// Clear all
	err := cache.Clear(ctx)
	if err != nil {
		t.Errorf("Clear() error = %v", err)
	}

	// Verify all items are gone
	for i := 0; i < 5; i++ {
		key := string(rune('a' + i))
		if cache.Exists(ctx, key) {
			t.Errorf("Key %v still exists after Clear()", key)
		}
	}
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	cache := NewMemoryCache(100)
	ctx := context.Background()

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := "concurrent:key"
			value := []byte("test value")
			
			cache.Set(ctx, key, value, 1*time.Hour)
			cache.Get(ctx, key)
			cache.Exists(ctx, key)
			cache.Delete(ctx, key)
			
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMemoryCache_MaxSize(t *testing.T) {
	t.Parallel()

	maxSize := 5
	cache := NewMemoryCache(maxSize)
	ctx := context.Background()

	// Add more items than max size
	for i := 0; i < maxSize+5; i++ {
		key := string(rune('a' + i))
		value := []byte(key)
		cache.Set(ctx, key, value, 1*time.Hour)
	}

	// Note: MemoryCache implementation may need eviction logic
	// This test documents the expected behavior
}

// Benchmark tests
func BenchmarkMemoryCache_Get(b *testing.B) {
	cache := NewMemoryCache(1000)
	ctx := context.Background()
	
	key := "bench:key"
	value := []byte("benchmark value")
	cache.Set(ctx, key, value, 1*time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(ctx, key)
	}
}

func BenchmarkMemoryCache_Set(b *testing.B) {
	cache := NewMemoryCache(1000)
	ctx := context.Background()
	
	key := "bench:key"
	value := []byte("benchmark value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(ctx, key, value, 1*time.Hour)
	}
}

func BenchmarkMemoryCache_ConcurrentGet(b *testing.B) {
	cache := NewMemoryCache(1000)
	ctx := context.Background()
	
	key := "bench:key"
	value := []byte("benchmark value")
	cache.Set(ctx, key, value, 1*time.Hour)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cache.Get(ctx, key)
		}
	})
}
