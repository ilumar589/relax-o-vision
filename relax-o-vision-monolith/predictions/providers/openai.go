package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements LLMProvider for OpenAI
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = openai.GPT4
	}
	return &OpenAIProvider{
		client: openai.NewClient(apiKey),
		model:  model,
	}
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Analyze performs analysis using OpenAI
func (p *OpenAIProvider) Analyze(ctx context.Context, prompt string, data interface{}) (*AnalysisResult, error) {
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

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert football analyst. Provide predictions based on the given data.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fullPrompt,
			},
		},
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("openai api error: %w", err)
	}

	return parseAnalysisResponse(resp.Choices[0].Message.Content)
}

// GenerateEmbedding generates an embedding using OpenAI
func (p *OpenAIProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	resp, err := p.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return resp.Data[0].Embedding, nil
}

// parseAnalysisResponse parses the LLM response into AnalysisResult
func parseAnalysisResponse(response string) (*AnalysisResult, error) {
	var result AnalysisResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("failed to parse analysis response: %w", err)
	}
	return &result, nil
}
