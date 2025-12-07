package footballdata

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
)

// MockCache implements cache.Cache for testing
type MockCache struct {
	mu   sync.RWMutex
	data map[string][]byte
	ttls map[string]time.Duration
}

// NewMockCache creates a new mock cache
func NewMockCache() *MockCache {
	return &MockCache{
		data: make(map[string][]byte),
		ttls: make(map[string]time.Duration),
	}
}

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (m *MockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	m.ttls[key] = ttl
	return nil
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	delete(m.ttls, key)
	return nil
}

func (m *MockCache) Exists(ctx context.Context, key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok
}

func (m *MockCache) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string][]byte)
	m.ttls = make(map[string]time.Duration)
	return nil
}

// MockRepository implements Repository interface for testing
type MockRepository struct {
	mu           sync.RWMutex
	competitions map[int]*Competition
	teams        map[int]*Team
	matches      map[int]*Match
	saveError    error
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		competitions: make(map[int]*Competition),
		teams:        make(map[int]*Team),
		matches:      make(map[int]*Match),
	}
}

func (m *MockRepository) SaveCompetition(ctx context.Context, comp *Competition) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.competitions[comp.ID] = comp
	return nil
}

func (m *MockRepository) GetCompetition(ctx context.Context, id int) (*Competition, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	comp, ok := m.competitions[id]
	if !ok {
		return nil, errors.New("competition not found")
	}
	return comp, nil
}

func (m *MockRepository) SaveTeam(ctx context.Context, team *Team) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.teams[team.ID] = team
	return nil
}

func (m *MockRepository) GetTeam(ctx context.Context, id int) (*Team, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	team, ok := m.teams[id]
	if !ok {
		return nil, errors.New("team not found")
	}
	return team, nil
}

func (m *MockRepository) SaveMatch(ctx context.Context, match *Match) error {
	if m.saveError != nil {
		return m.saveError
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.matches[match.ID] = match
	return nil
}

func (m *MockRepository) GetMatch(ctx context.Context, id int) (*Match, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	match, ok := m.matches[id]
	if !ok {
		return nil, errors.New("match not found")
	}
	return match, nil
}

func (m *MockRepository) UpdateCompetitionEmbedding(ctx context.Context, id int, embedding []float32) error {
	return nil
}

func (m *MockRepository) UpdateTeamEmbedding(ctx context.Context, id int, embedding []float32) error {
	return nil
}

func (m *MockRepository) UpdateMatchEmbedding(ctx context.Context, id int, embedding []float32) error {
	return nil
}

// SetSaveError allows tests to simulate save errors
func (m *MockRepository) SetSaveError(err error) {
	m.saveError = err
}

// MockHTTPClient for testing API client
type MockHTTPClient struct {
	mu        sync.RWMutex
	responses map[string]*http.Response
	errors    map[string]error
}

// NewMockHTTPClient creates a new mock HTTP client
func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		responses: make(map[string]*http.Response),
		errors:    make(map[string]error),
	}
}

// AddResponse adds a mock response for a URL
func (m *MockHTTPClient) AddResponse(url string, response *http.Response) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[url] = response
}

// AddError adds a mock error for a URL
func (m *MockHTTPClient) AddError(url string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[url] = err
}

// GetResponse retrieves a mock response
func (m *MockHTTPClient) GetResponse(url string) (*http.Response, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if err, ok := m.errors[url]; ok {
		return nil, err
	}
	if resp, ok := m.responses[url]; ok {
		return resp, nil
	}
	return nil, errors.New("no mock response configured")
}
