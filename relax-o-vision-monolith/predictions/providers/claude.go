package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ClaudeProvider implements LLMProvider for Anthropic Claude
type ClaudeProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider(apiKey, model string) *ClaudeProvider {
	if model == "" {
		model = "claude-3-5-sonnet-20241022" // Latest Claude 3.5 Sonnet
	}
	return &ClaudeProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

// Name returns the provider name
func (p *ClaudeProvider) Name() string {
	return "claude"
}

// Analyze performs analysis using Claude
func (p *ClaudeProvider) Analyze(ctx context.Context, prompt string, data interface{}) (*AnalysisResult, error) {
	dataJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}

	fullPrompt := fmt.Sprintf(`%s

Data:
%s

Provide your analysis in JSON format:
{
  "homeWinProb": <0-1>,
  "drawProb": <0-1>,
  "awayWinProb": <0-1>,
  "confidence": <0-1>,
  "reasoning": "<explanation>",
  "keyFactors": ["factor1", "factor2", ...]
}`, prompt, string(dataJSON))

	requestBody := map[string]interface{}{
		"model": p.model,
		"max_tokens": 1024,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": fullPrompt,
			},
		},
		"system": "You are an expert football analyst. Provide predictions based on the given data.",
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("claude api error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("claude api returned status %d: %s", resp.StatusCode, string(body))
	}

	var claudeResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	return parseAnalysisResponse(claudeResp.Content[0].Text)
}

// GenerateEmbedding generates an embedding using Claude's embeddings API
func (p *ClaudeProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Claude does not have a native embeddings API as of now
	// Configure OpenAI or Gemini providers for embedding generation instead
	return nil, fmt.Errorf("claude provider does not support embeddings - configure OpenAI or Gemini providers for embedding generation")
}
