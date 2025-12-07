package predictions

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dapr/go-sdk/workflow"
)

// PredictionWorkflow defines the Dapr workflow for match predictions
func PredictionWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	var input WorkflowInput
	if err := ctx.GetInput(&input); err != nil {
		return nil, fmt.Errorf("failed to get workflow input: %w", err)
	}

	slog.Info("Starting prediction workflow", "matchId", input.MatchID)

	// Step 1: Fetch match data
	var matchAnalysis MatchAnalysis
	if err := ctx.CallActivity(FetchMatchDataActivity, workflow.ActivityInput(input.MatchID)).Await(&matchAnalysis); err != nil {
		return nil, fmt.Errorf("failed to fetch match data: %w", err)
	}

	// Step 2: Run statistical agent
	var statOutput AgentOutput
	if err := ctx.CallActivity(StatisticalAnalysisActivity, workflow.ActivityInput(matchAnalysis)).Await(&statOutput); err != nil {
		slog.Error("Statistical agent failed", "error", err)
		statOutput = AgentOutput{
			AgentType:   AgentTypeStatistical,
			Confidence:  0.0,
			Reasoning:   "Analysis failed",
		}
	}

	// Step 3: Run form agent
	var formOutput AgentOutput
	if err := ctx.CallActivity(FormAnalysisActivity, workflow.ActivityInput(matchAnalysis)).Await(&formOutput); err != nil {
		slog.Error("Form agent failed", "error", err)
		formOutput = AgentOutput{
			AgentType:   AgentTypeForm,
			Confidence:  0.0,
			Reasoning:   "Analysis failed",
		}
	}

	// Step 4: Run head-to-head agent
	var h2hOutput AgentOutput
	if err := ctx.CallActivity(HeadToHeadAnalysisActivity, workflow.ActivityInput(matchAnalysis)).Await(&h2hOutput); err != nil {
		slog.Error("Head-to-head agent failed", "error", err)
		h2hOutput = AgentOutput{
			AgentType:   AgentTypeHeadToHead,
			Confidence:  0.0,
			Reasoning:   "Analysis failed",
		}
	}

	// Step 5: Aggregate predictions
	agentOutputs := []AgentOutput{statOutput, formOutput, h2hOutput}
	var aggregateOutput AgentOutput
	if err := ctx.CallActivity(AggregateAnalysisActivity, workflow.ActivityInput(agentOutputs)).Await(&aggregateOutput); err != nil {
		return nil, fmt.Errorf("failed to aggregate predictions: %w", err)
	}

	// Build final output
	output := WorkflowOutput{
		HomeWinProb:  aggregateOutput.HomeWinProb,
		DrawProb:     aggregateOutput.DrawProb,
		AwayWinProb:  aggregateOutput.AwayWinProb,
		Confidence:   aggregateOutput.Confidence,
		Reasoning:    aggregateOutput.Reasoning,
		AgentOutputs: agentOutputs,
	}

	slog.Info("Prediction workflow completed", "matchId", input.MatchID, "confidence", output.Confidence)
	return output, nil
}

// Activity names
const (
	FetchMatchDataActivity          = "FetchMatchDataActivity"
	StatisticalAnalysisActivity     = "StatisticalAnalysisActivity"
	FormAnalysisActivity            = "FormAnalysisActivity"
	HeadToHeadAnalysisActivity      = "HeadToHeadAnalysisActivity"
	AggregateAnalysisActivity       = "AggregateAnalysisActivity"
)

// Activity functions (to be implemented by the service)

// FetchMatchDataActivityFunc fetches match data for analysis
type FetchMatchDataActivityFunc func(ctx context.Context, matchID int) (*MatchAnalysis, error)

// StatisticalAnalysisActivityFunc performs statistical analysis
type StatisticalAnalysisActivityFunc func(ctx context.Context, analysis *MatchAnalysis) (*AgentOutput, error)

// FormAnalysisActivityFunc performs form analysis
type FormAnalysisActivityFunc func(ctx context.Context, analysis *MatchAnalysis) (*AgentOutput, error)

// HeadToHeadAnalysisActivityFunc performs head-to-head analysis
type HeadToHeadAnalysisActivityFunc func(ctx context.Context, analysis *MatchAnalysis) (*AgentOutput, error)

// AggregateAnalysisActivityFunc aggregates multiple agent outputs
type AggregateAnalysisActivityFunc func(ctx context.Context, outputs []AgentOutput) (*AgentOutput, error)
