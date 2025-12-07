package predictions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service handles business logic for predictions
type Service struct {
	db                *sql.DB
	statisticalAgent  *StatisticalAgent
	formAgent         *FormAgent
	headToHeadAgent   *HeadToHeadAgent
	aggregatorAgent   *AggregatorAgent
}

// NewService creates a new prediction service
func NewService(db *sql.DB, openAIKey string) *Service {
	return &Service{
		db:               db,
		statisticalAgent: NewStatisticalAgent(openAIKey),
		formAgent:        NewFormAgent(openAIKey),
		headToHeadAgent:  NewHeadToHeadAgent(openAIKey),
		aggregatorAgent:  NewAggregatorAgent(openAIKey),
	}
}

// CreatePrediction creates a new prediction for a match
func (s *Service) CreatePrediction(ctx context.Context, matchID int) (*PredictionResult, error) {
	// Fetch match analysis data
	analysis, err := s.fetchMatchAnalysis(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch match analysis: %w", err)
	}

	// Run agents in parallel (simplified version - in production use Dapr workflow)
	statOutput, err := s.statisticalAgent.Analyze(ctx, analysis)
	if err != nil {
		return nil, fmt.Errorf("statistical analysis failed: %w", err)
	}

	formOutput, err := s.formAgent.Analyze(ctx, analysis)
	if err != nil {
		return nil, fmt.Errorf("form analysis failed: %w", err)
	}

	h2hOutput, err := s.headToHeadAgent.Analyze(ctx, analysis)
	if err != nil {
		return nil, fmt.Errorf("head-to-head analysis failed: %w", err)
	}

	// Aggregate results
	agentOutputs := []AgentOutput{*statOutput, *formOutput, *h2hOutput}
	finalOutput, err := s.aggregatorAgent.Aggregate(ctx, agentOutputs)
	if err != nil {
		return nil, fmt.Errorf("aggregation failed: %w", err)
	}

	// Save prediction to database
	prediction := &PredictionResult{
		ID:           uuid.New().String(),
		MatchID:      matchID,
		HomeWinProb:  finalOutput.HomeWinProb,
		DrawProb:     finalOutput.DrawProb,
		AwayWinProb:  finalOutput.AwayWinProb,
		Confidence:   finalOutput.Confidence,
		Status:       "completed",
		AgentOutputs: agentOutputs,
		Reasoning:    finalOutput.Reasoning,
		KeyFactors:   finalOutput.KeyFactors,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.savePrediction(ctx, prediction); err != nil {
		return nil, fmt.Errorf("failed to save prediction: %w", err)
	}

	return prediction, nil
}

// GetPrediction retrieves a prediction by ID
func (s *Service) GetPrediction(ctx context.Context, id string) (*PredictionResult, error) {
	query := `
		SELECT id, match_id, home_win_prob, draw_prob, away_win_prob, confidence, 
		       reasoning, agent_outputs, workflow_id, status, created_at, updated_at
		FROM predictions
		WHERE id = $1
	`

	var prediction PredictionResult
	var reasoningJSON, agentOutputsJSON []byte
	var workflowID sql.NullString

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&prediction.ID,
		&prediction.MatchID,
		&prediction.HomeWinProb,
		&prediction.DrawProb,
		&prediction.AwayWinProb,
		&prediction.Confidence,
		&reasoningJSON,
		&agentOutputsJSON,
		&workflowID,
		&prediction.Status,
		&prediction.CreatedAt,
		&prediction.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("prediction not found")
		}
		return nil, fmt.Errorf("failed to get prediction: %w", err)
	}

	prediction.WorkflowID = workflowID.String

	// Parse JSON fields
	var reasoning map[string]any
	if err := json.Unmarshal(reasoningJSON, &reasoning); err == nil {
		if r, ok := reasoning["text"].(string); ok {
			prediction.Reasoning = r
		}
	}

	if err := json.Unmarshal(agentOutputsJSON, &prediction.AgentOutputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent outputs: %w", err)
	}

	return &prediction, nil
}

// GetPredictionsByMatch retrieves all predictions for a match
func (s *Service) GetPredictionsByMatch(ctx context.Context, matchID int) ([]PredictionResult, error) {
	query := `
		SELECT id, match_id, home_win_prob, draw_prob, away_win_prob, confidence,
		       reasoning, agent_outputs, workflow_id, status, created_at, updated_at
		FROM predictions
		WHERE match_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query predictions: %w", err)
	}
	defer rows.Close()

	var predictions []PredictionResult
	for rows.Next() {
		var prediction PredictionResult
		var reasoningJSON, agentOutputsJSON []byte
		var workflowID sql.NullString

		err := rows.Scan(
			&prediction.ID,
			&prediction.MatchID,
			&prediction.HomeWinProb,
			&prediction.DrawProb,
			&prediction.AwayWinProb,
			&prediction.Confidence,
			&reasoningJSON,
			&agentOutputsJSON,
			&workflowID,
			&prediction.Status,
			&prediction.CreatedAt,
			&prediction.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan prediction: %w", err)
		}

		prediction.WorkflowID = workflowID.String

		// Parse JSON fields
		var reasoning map[string]any
		if err := json.Unmarshal(reasoningJSON, &reasoning); err == nil {
			if r, ok := reasoning["text"].(string); ok {
				prediction.Reasoning = r
			}
		}

		if err := json.Unmarshal(agentOutputsJSON, &prediction.AgentOutputs); err != nil {
			continue
		}

		predictions = append(predictions, prediction)
	}

	return predictions, nil
}

// Helper functions

func (s *Service) fetchMatchAnalysis(ctx context.Context, matchID int) (*MatchAnalysis, error) {
	// Fetch match details from database
	query := `
		SELECT id, home_team, away_team, utc_date
		FROM matches
		WHERE id = $1
	`

	var homeTeamJSON, awayTeamJSON []byte
	var utcDate time.Time
	err := s.db.QueryRowContext(ctx, query, matchID).Scan(&matchID, &homeTeamJSON, &awayTeamJSON, &utcDate)
	if err != nil {
		return nil, fmt.Errorf("match not found: %w", err)
	}

	// Parse team data
	var homeTeamData, awayTeamData map[string]any
	if err := json.Unmarshal(homeTeamJSON, &homeTeamData); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(awayTeamJSON, &awayTeamData); err != nil {
		return nil, err
	}

	// Build analysis structure (simplified version)
	analysis := &MatchAnalysis{
		MatchID:   matchID,
		MatchDate: utcDate,
		HomeTeam: TeamAnalysis{
			ID:   int(homeTeamData["id"].(float64)),
			Name: homeTeamData["name"].(string),
			Statistics: TeamStatistics{
				MatchesPlayed: 10,
				Wins:          5,
				Draws:         3,
				Losses:        2,
			},
		},
		AwayTeam: TeamAnalysis{
			ID:   int(awayTeamData["id"].(float64)),
			Name: awayTeamData["name"].(string),
			Statistics: TeamStatistics{
				MatchesPlayed: 10,
				Wins:          4,
				Draws:         4,
				Losses:        2,
			},
		},
		HeadToHead: []HistoricalMatch{},
	}

	return analysis, nil
}

func (s *Service) savePrediction(ctx context.Context, prediction *PredictionResult) error {
	reasoningJSON, err := json.Marshal(map[string]any{"text": prediction.Reasoning})
	if err != nil {
		return fmt.Errorf("failed to marshal reasoning: %w", err)
	}

	agentOutputsJSON, err := json.Marshal(prediction.AgentOutputs)
	if err != nil {
		return fmt.Errorf("failed to marshal agent outputs: %w", err)
	}

	query := `
		INSERT INTO predictions (id, match_id, home_win_prob, draw_prob, away_win_prob, confidence, reasoning, agent_outputs, workflow_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = s.db.ExecContext(ctx, query,
		prediction.ID,
		prediction.MatchID,
		prediction.HomeWinProb,
		prediction.DrawProb,
		prediction.AwayWinProb,
		prediction.Confidence,
		reasoningJSON,
		agentOutputsJSON,
		sql.NullString{String: prediction.WorkflowID, Valid: prediction.WorkflowID != ""},
		prediction.Status,
		prediction.CreatedAt,
		prediction.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert prediction: %w", err)
	}

	return nil
}
