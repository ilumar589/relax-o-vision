package predictions

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

const (
	// Agent types
	AgentTypeStatistical  = "statistical"
	AgentTypeForm         = "form"
	AgentTypeHeadToHead   = "head-to-head"
	AgentTypeAggregator   = "aggregator"
)

// Agent represents an AI agent for match prediction
type Agent struct {
	agentType string
	client    *openai.Client
}

// NewAgent creates a new AI agent
func NewAgent(agentType string, apiKey string) *Agent {
	return &Agent{
		agentType: agentType,
		client:    openai.NewClient(apiKey),
	}
}

// StatisticalAgent analyzes historical statistics
type StatisticalAgent struct {
	*Agent
}

// NewStatisticalAgent creates a new statistical analysis agent
func NewStatisticalAgent(apiKey string) *StatisticalAgent {
	return &StatisticalAgent{
		Agent: NewAgent(AgentTypeStatistical, apiKey),
	}
}

// Analyze performs statistical analysis on match data
func (a *StatisticalAgent) Analyze(ctx context.Context, analysis *MatchAnalysis) (*AgentOutput, error) {
	prompt := buildStatisticalPrompt(analysis)
	
	resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert football analyst specializing in statistical analysis. Provide predictions based on team statistics.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get statistical analysis: %w", err)
	}

	return parseAgentResponse(a.agentType, resp.Choices[0].Message.Content)
}

// FormAgent evaluates recent team form
type FormAgent struct {
	*Agent
}

// NewFormAgent creates a new form analysis agent
func NewFormAgent(apiKey string) *FormAgent {
	return &FormAgent{
		Agent: NewAgent(AgentTypeForm, apiKey),
	}
}

// Analyze performs form analysis on match data
func (a *FormAgent) Analyze(ctx context.Context, analysis *MatchAnalysis) (*AgentOutput, error) {
	prompt := buildFormPrompt(analysis)
	
	resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert football analyst specializing in recent team form. Focus on momentum and current performance trends.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get form analysis: %w", err)
	}

	return parseAgentResponse(a.agentType, resp.Choices[0].Message.Content)
}

// HeadToHeadAgent analyzes head-to-head records
type HeadToHeadAgent struct {
	*Agent
}

// NewHeadToHeadAgent creates a new head-to-head analysis agent
func NewHeadToHeadAgent(apiKey string) *HeadToHeadAgent {
	return &HeadToHeadAgent{
		Agent: NewAgent(AgentTypeHeadToHead, apiKey),
	}
}

// Analyze performs head-to-head analysis on match data
func (a *HeadToHeadAgent) Analyze(ctx context.Context, analysis *MatchAnalysis) (*AgentOutput, error) {
	prompt := buildHeadToHeadPrompt(analysis)
	
	resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert football analyst specializing in head-to-head matchups. Analyze historical encounters between teams.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get head-to-head analysis: %w", err)
	}

	return parseAgentResponse(a.agentType, resp.Choices[0].Message.Content)
}

// AggregatorAgent combines insights from multiple agents
type AggregatorAgent struct {
	*Agent
}

// NewAggregatorAgent creates a new aggregator agent
func NewAggregatorAgent(apiKey string) *AggregatorAgent {
	return &AggregatorAgent{
		Agent: NewAgent(AgentTypeAggregator, apiKey),
	}
}

// Aggregate combines outputs from multiple agents
func (a *AggregatorAgent) Aggregate(ctx context.Context, outputs []AgentOutput) (*AgentOutput, error) {
	prompt := buildAggregatorPrompt(outputs)
	
	resp, err := a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an expert football analyst who synthesizes multiple perspectives into a final prediction. Weight the different analyses and provide a consensus prediction.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate predictions: %w", err)
	}

	return parseAgentResponse(a.agentType, resp.Choices[0].Message.Content)
}

// Helper functions for building prompts

func buildStatisticalPrompt(analysis *MatchAnalysis) string {
	data, _ := json.MarshalIndent(analysis, "", "  ")
	return fmt.Sprintf(`Analyze the following match data and provide a prediction based on team statistics:

%s

Provide your analysis in JSON format:
{
  "homeWinProb": <0-1>,
  "drawProb": <0-1>,
  "awayWinProb": <0-1>,
  "confidence": <0-1>,
  "reasoning": "<explanation>",
  "keyFactors": ["factor1", "factor2", ...]
}`, string(data))
}

func buildFormPrompt(analysis *MatchAnalysis) string {
	data, _ := json.MarshalIndent(analysis, "", "  ")
	return fmt.Sprintf(`Analyze the following match data focusing on recent team form:

%s

Provide your analysis in JSON format:
{
  "homeWinProb": <0-1>,
  "drawProb": <0-1>,
  "awayWinProb": <0-1>,
  "confidence": <0-1>,
  "reasoning": "<explanation>",
  "keyFactors": ["factor1", "factor2", ...]
}`, string(data))
}

func buildHeadToHeadPrompt(analysis *MatchAnalysis) string {
	data, _ := json.MarshalIndent(analysis, "", "  ")
	return fmt.Sprintf(`Analyze the head-to-head history between these teams:

%s

Provide your analysis in JSON format:
{
  "homeWinProb": <0-1>,
  "drawProb": <0-1>,
  "awayWinProb": <0-1>,
  "confidence": <0-1>,
  "reasoning": "<explanation>",
  "keyFactors": ["factor1", "factor2", ...]
}`, string(data))
}

func buildAggregatorPrompt(outputs []AgentOutput) string {
	data, _ := json.MarshalIndent(outputs, "", "  ")
	return fmt.Sprintf(`Synthesize the following agent predictions into a final consensus prediction:

%s

Provide your aggregated analysis in JSON format:
{
  "homeWinProb": <0-1>,
  "drawProb": <0-1>,
  "awayWinProb": <0-1>,
  "confidence": <0-1>,
  "reasoning": "<explanation>",
  "keyFactors": ["factor1", "factor2", ...]
}`, string(data))
}

func parseAgentResponse(agentType, response string) (*AgentOutput, error) {
	var output struct {
		HomeWinProb float64  `json:"homeWinProb"`
		DrawProb    float64  `json:"drawProb"`
		AwayWinProb float64  `json:"awayWinProb"`
		Confidence  float64  `json:"confidence"`
		Reasoning   string   `json:"reasoning"`
		KeyFactors  []string `json:"keyFactors"`
	}

	if err := json.Unmarshal([]byte(response), &output); err != nil {
		return nil, fmt.Errorf("failed to parse agent response: %w", err)
	}

	return &AgentOutput{
		AgentType:   agentType,
		HomeWinProb: output.HomeWinProb,
		DrawProb:    output.DrawProb,
		AwayWinProb: output.AwayWinProb,
		Confidence:  output.Confidence,
		Reasoning:   output.Reasoning,
		KeyFactors:  output.KeyFactors,
	}, nil
}
