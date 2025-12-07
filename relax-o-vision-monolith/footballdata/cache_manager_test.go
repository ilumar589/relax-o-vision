package footballdata

import (
	"context"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestCacheManager_GetMetadata(t *testing.T) {
	// Skip if no database available
	t.Skip("Integration test - requires PostgreSQL database")
}

func TestCacheManager_SetMetadata(t *testing.T) {
	// Skip if no database available
	t.Skip("Integration test - requires PostgreSQL database")
}

func TestCacheManager_NeedsRefresh(t *testing.T) {
	tests := []struct {
		name        string
		metadata    *CacheMetadata
		metadataErr error
		expected    bool
	}{
		{
			name:        "no metadata - needs refresh",
			metadata:    nil,
			metadataErr: nil,
			expected:    true,
		},
		{
			name: "fresh data - no refresh",
			metadata: &CacheMetadata{
				ID:         1,
				EntityType: "competition",
				EntityKey:  "PL",
				CachedAt:   time.Now().Add(-1 * time.Hour),
				ExpiresAt:  time.Now().Add(29 * 24 * time.Hour),
			},
			metadataErr: nil,
			expected:    false,
		},
		{
			name: "expired data - needs refresh",
			metadata: &CacheMetadata{
				ID:         1,
				EntityType: "competition",
				EntityKey:  "PL",
				CachedAt:   time.Now().Add(-31 * 24 * time.Hour),
				ExpiresAt:  time.Now().Add(-1 * time.Hour),
			},
			metadataErr: nil,
			expected:    true,
		},
		{
			name: "borderline - exactly at expiration",
			metadata: &CacheMetadata{
				ID:         1,
				EntityType: "competition",
				EntityKey:  "PL",
				CachedAt:   time.Now().Add(-30 * 24 * time.Hour),
				ExpiresAt:  time.Now(),
			},
			metadataErr: nil,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This would need a database connection to test properly
			// For unit testing, we would mock the database
			t.Skip("Requires database mocking")
		})
	}
}

func TestCacheManager_Get_CacheHit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockCache := NewMockCache()
	
	cm := &CacheManager{
		redis: mockCache,
		db:    nil, // Not needed for this test
	}

	// Set up test data
	testKey := "test:key"
	testData := []byte("test data")
	err := mockCache.Set(ctx, testKey, testData, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Test cache hit
	data, err := cm.Get(ctx, testKey)
	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if string(data) != string(testData) {
		t.Errorf("Get() = %v, want %v", string(data), string(testData))
	}
}

func TestCacheManager_Get_CacheMiss(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockCache := NewMockCache()
	
	cm := &CacheManager{
		redis: mockCache,
		db:    nil,
	}

	// Test cache miss
	data, err := cm.Get(ctx, "non-existent-key")
	if err != nil {
		t.Errorf("Get() error = %v, want nil", err)
	}
	if data != nil {
		t.Errorf("Get() = %v, want nil", data)
	}
}

func TestCacheManager_Set(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockCache := NewMockCache()
	
	cm := &CacheManager{
		redis: mockCache,
		db:    nil,
	}

	testKey := "test:key"
	testData := []byte("test data")
	testTTL := 1 * time.Hour

	err := cm.Set(ctx, testKey, testData, testTTL)
	if err != nil {
		t.Errorf("Set() error = %v, want nil", err)
	}

	// Verify data was stored
	data, err := mockCache.Get(ctx, testKey)
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("Stored data = %v, want %v", string(data), string(testData))
	}
}

func TestCacheManager_Delete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockCache := NewMockCache()
	
	cm := &CacheManager{
		redis: mockCache,
		db:    nil,
	}

	testKey := "test:key"
	testData := []byte("test data")

	// Set data first
	mockCache.Set(ctx, testKey, testData, 1*time.Hour)

	// Delete it
	err := cm.Delete(ctx, testKey)
	if err != nil {
		t.Errorf("Delete() error = %v, want nil", err)
	}

	// Verify it's gone
	data, _ := mockCache.Get(ctx, testKey)
	if data != nil {
		t.Errorf("Data still exists after Delete()")
	}
}

func TestCacheManager_InvalidateEntity(t *testing.T) {
	// Skip - requires database
	t.Skip("Integration test - requires PostgreSQL database")
}

func TestComputeDataHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data any
	}{
		{
			name: "simple string",
			data: "test",
		},
		{
			name: "struct data",
			data: struct {
				ID   int
				Name string
			}{ID: 1, Name: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeDataHash(tt.data)
			if got == "" {
				t.Errorf("ComputeDataHash() returned empty string")
			}
			if len(got) != 64 { // SHA256 hash is 64 hex characters
				t.Errorf("ComputeDataHash() length = %d, want 64", len(got))
			}
			
			// Verify consistency - same input should give same hash
			got2 := ComputeDataHash(tt.data)
			if got != got2 {
				t.Errorf("ComputeDataHash() not consistent: %v != %v", got, got2)
			}
		})
	}
}

func TestGetCacheTTL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		envVar  string
		want    time.Duration
	}{
		{
			name:    "default TTL",
			envVar:  "",
			want:    30 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv("CACHE_TTL_DAYS", tt.envVar)
			}
			got := GetCacheTTL()
			if got != tt.want {
				t.Errorf("GetCacheTTL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacheManager_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockCache := NewMockCache()
	
	cm := &CacheManager{
		redis: mockCache,
		db:    nil,
	}

	// Run concurrent Get/Set operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := "concurrent:key"
			data := []byte("test data")
			
			// Set
			cm.Set(ctx, key, data, 1*time.Hour)
			
			// Get
			cm.Get(ctx, key)
			
			// Delete
			cm.Delete(ctx, key)
			
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestCacheManager_NilRedis(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	
	cm := &CacheManager{
		redis: nil,
		db:    nil,
	}

	// Test that operations don't panic with nil redis
	_, err := cm.Get(ctx, "test:key")
	if err != nil {
		t.Errorf("Get() with nil redis should not error: %v", err)
	}

	err = cm.Set(ctx, "test:key", []byte("data"), 1*time.Hour)
	if err != nil {
		t.Errorf("Set() with nil redis should not error: %v", err)
	}

	err = cm.Delete(ctx, "test:key")
	if err != nil {
		t.Errorf("Delete() with nil redis should not error: %v", err)
	}
}

// Benchmark tests
func BenchmarkCacheManager_Get(b *testing.B) {
	ctx := context.Background()
	mockCache := NewMockCache()
	cm := &CacheManager{redis: mockCache}
	
	testKey := "bench:key"
	testData := []byte("benchmark data")
	mockCache.Set(ctx, testKey, testData, 1*time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.Get(ctx, testKey)
	}
}

func BenchmarkCacheManager_Set(b *testing.B) {
	ctx := context.Background()
	mockCache := NewMockCache()
	cm := &CacheManager{redis: mockCache}
	
	testKey := "bench:key"
	testData := []byte("benchmark data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.Set(ctx, testKey, testData, 1*time.Hour)
	}
}
