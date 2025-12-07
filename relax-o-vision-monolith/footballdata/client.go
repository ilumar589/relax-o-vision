package footballdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL            = "https://api.football-data.org/v4"
	requestsPerMinute  = 10
	rateLimitDuration  = time.Minute
)

// Client represents the football-data.org API client
type Client struct {
	apiKey      string
	httpClient  *http.Client
	lastRequest time.Time
	mu          sync.Mutex // Protects lastRequest
}

// NewClient creates a new football-data.org API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		lastRequest: time.Time{},
	}
}

// doRequest performs an HTTP request with rate limiting and authentication
func (c *Client) doRequest(ctx context.Context, endpoint string) ([]byte, error) {
	// Rate limiting - ensure we don't exceed 10 requests per minute
	c.mu.Lock()
	if !c.lastRequest.IsZero() {
		elapsed := time.Since(c.lastRequest)
		if elapsed < rateLimitDuration/requestsPerMinute {
			sleepDuration := (rateLimitDuration / requestsPerMinute) - elapsed
			c.mu.Unlock()
			time.Sleep(sleepDuration)
			c.mu.Lock()
		}
	}
	c.lastRequest = time.Now()
	c.mu.Unlock()

	url := fmt.Sprintf("%s%s", baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Auth-Token", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetCompetitions fetches all available competitions
func (c *Client) GetCompetitions(ctx context.Context) ([]Competition, error) {
	body, err := c.doRequest(ctx, "/competitions")
	if err != nil {
		return nil, err
	}

	var response CompetitionsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Competitions, nil
}

// GetCompetition fetches a specific competition by code
func (c *Client) GetCompetition(ctx context.Context, code string) (*Competition, error) {
	endpoint := fmt.Sprintf("/competitions/%s", code)
	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var competition Competition
	if err := json.Unmarshal(body, &competition); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &competition, nil
}

// GetTeam fetches a specific team by ID
func (c *Client) GetTeam(ctx context.Context, teamID int) (*Team, error) {
	endpoint := fmt.Sprintf("/teams/%d", teamID)
	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var team Team
	if err := json.Unmarshal(body, &team); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &team, nil
}

// GetMatches fetches matches for a competition
func (c *Client) GetMatches(ctx context.Context, competitionCode string) ([]Match, error) {
	endpoint := fmt.Sprintf("/competitions/%s/matches", competitionCode)
	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var response MatchesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.Matches, nil
}

// GetStandings fetches standings for a competition
func (c *Client) GetStandings(ctx context.Context, competitionCode string) (*Standing, error) {
	endpoint := fmt.Sprintf("/competitions/%s/standings", competitionCode)
	body, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, err
	}

	var standing Standing
	if err := json.Unmarshal(body, &standing); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &standing, nil
}
