package providers

import (
	"context"
	"fmt"
)

// LLMProvider interface for different LLM providers
type LLMProvider interface {
	Name() string
	Analyze(ctx context.Context, prompt string, data interface{}) (*AnalysisResult, error)
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
}

// ProviderConfig holds configuration for a provider
type ProviderConfig struct {
	Name    string
	APIKey  string
	Model   string
	Enabled bool
	Weight  float64 // For weighted aggregation
}

// AnalysisResult represents the result from LLM analysis
type AnalysisResult struct {
	HomeWinProb float64            `json:"homeWinProb"`
	DrawProb    float64            `json:"drawProb"`
	AwayWinProb float64            `json:"awayWinProb"`
	Confidence  float64            `json:"confidence"`
	Reasoning   string             `json:"reasoning"`
	KeyFactors  []string           `json:"keyFactors"`
	Metadata    map[string]any     `json:"metadata,omitempty"`
}

// ProviderFactory creates LLM providers based on configuration
type ProviderFactory struct {
	configs []ProviderConfig
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(configs []ProviderConfig) *ProviderFactory {
	return &ProviderFactory{
		configs: configs,
	}
}

// CreateProviders creates all enabled providers
func (f *ProviderFactory) CreateProviders() ([]LLMProvider, error) {
	var providers []LLMProvider
	
	for _, config := range f.configs {
		if !config.Enabled {
			continue
		}
		
		provider, err := f.createProvider(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create provider %s: %w", config.Name, err)
		}
		providers = append(providers, provider)
	}
	
	if len(providers) == 0 {
		return nil, fmt.Errorf("no enabled providers configured")
	}
	
	return providers, nil
}

// createProvider creates a single provider based on config
func (f *ProviderFactory) createProvider(config ProviderConfig) (LLMProvider, error) {
	switch config.Name {
	case "openai":
		return NewOpenAIProvider(config.APIKey, config.Model), nil
	case "claude":
		return NewClaudeProvider(config.APIKey, config.Model), nil
	case "gemini":
		return NewGeminiProvider(config.APIKey, config.Model), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", config.Name)
	}
}

// GetProvider returns a provider by name
func (f *ProviderFactory) GetProvider(name string) (LLMProvider, error) {
	for _, config := range f.configs {
		if config.Name == name && config.Enabled {
			return f.createProvider(config)
		}
	}
	return nil, fmt.Errorf("provider %s not found or not enabled", name)
}
