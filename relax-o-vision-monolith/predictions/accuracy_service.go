package predictions

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// AccuracyService handles prediction accuracy tracking and calculation
type AccuracyService struct {
	db *sql.DB
}

// NewAccuracyService creates a new accuracy service
func NewAccuracyService(db *sql.DB) *AccuracyService {
	return &AccuracyService{
		db: db,
	}
}

// RecordOutcome records the outcome of a prediction after a match completes
func (s *AccuracyService) RecordOutcome(ctx context.Context, predictionID uuid.UUID, matchID int) error {
	// Get the prediction
	var homeWinProb, drawProb, awayWinProb, confidence float64
	var competitionID int
	
	predQuery := `
		SELECT p.home_win_prob, p.draw_prob, p.away_win_prob, p.confidence, m.competition_id
		FROM predictions p
		JOIN matches m ON p.match_id = m.id
		WHERE p.id = $1
	`
	
	err := s.db.QueryRowContext(ctx, predQuery, predictionID).Scan(
		&homeWinProb, &drawProb, &awayWinProb, &confidence, &competitionID,
	)
	if err != nil {
		return fmt.Errorf("failed to get prediction: %w", err)
	}

	// Get match result
	var homeScore, awayScore sql.NullInt64
	var competitionName string
	
	matchQuery := `
		SELECT 
			(score->'fullTime'->'home')::int,
			(score->'fullTime'->'away')::int,
			competition->>'name'
		FROM matches
		WHERE id = $1 AND status = 'FINISHED'
	`
	
	err = s.db.QueryRowContext(ctx, matchQuery, matchID).Scan(&homeScore, &awayScore, &competitionName)
	if err != nil {
		return fmt.Errorf("failed to get match result: %w", err)
	}

	if !homeScore.Valid || !awayScore.Valid {
		return fmt.Errorf("match does not have final score")
	}

	// Determine actual winner
	actualWinner := "draw"
	if homeScore.Int64 > awayScore.Int64 {
		actualWinner = "home"
	} else if awayScore.Int64 > homeScore.Int64 {
		actualWinner = "away"
	}

	// Determine predicted winner (highest probability)
	predictedWinner := "draw"
	maxProb := drawProb
	if homeWinProb > maxProb {
		predictedWinner = "home"
		maxProb = homeWinProb
	}
	if awayWinProb > maxProb {
		predictedWinner = "away"
	}

	wasCorrect := predictedWinner == actualWinner

	// Save outcome
	outcome := &PredictionOutcome{
		ID:              uuid.New(),
		PredictionID:    predictionID,
		MatchID:         matchID,
		PredictedWinner: predictedWinner,
		ActualWinner:    actualWinner,
		WasCorrect:      wasCorrect,
		ConfidenceScore: confidence,
		HomeWinProb:     homeWinProb,
		DrawProb:        drawProb,
		AwayWinProb:     awayWinProb,
		ActualHomeScore: int(homeScore.Int64),
		ActualAwayScore: int(awayScore.Int64),
		CompetitionID:   competitionID,
		CompetitionName: competitionName,
		CreatedAt:       time.Now(),
	}

	insertQuery := `
		INSERT INTO prediction_outcomes (
			id, prediction_id, match_id, predicted_winner, actual_winner, 
			was_correct, confidence_score, home_win_prob, draw_prob, away_win_prob,
			actual_home_score, actual_away_score, competition_id, competition_name, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err = s.db.ExecContext(ctx, insertQuery,
		outcome.ID, outcome.PredictionID, outcome.MatchID,
		outcome.PredictedWinner, outcome.ActualWinner, outcome.WasCorrect,
		outcome.ConfidenceScore, outcome.HomeWinProb, outcome.DrawProb, outcome.AwayWinProb,
		outcome.ActualHomeScore, outcome.ActualAwayScore,
		outcome.CompetitionID, outcome.CompetitionName, outcome.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert outcome: %w", err)
	}

	slog.Info("Recorded prediction outcome", 
		"predictionId", predictionID,
		"matchId", matchID, 
		"wasCorrect", wasCorrect,
	)

	return nil
}

// GetOverallStats calculates overall accuracy statistics
func (s *AccuracyService) GetOverallStats(ctx context.Context) (*AccuracyStats, error) {
	stats := &AccuracyStats{
		ByCompetition:     make(map[string]*CompetitionAcc),
		ByConfidenceRange: make(map[string]*RangeAcc),
		ByProvider:        make(map[string]*ProviderAcc),
		ByAgent:           make(map[string]*AgentAcc),
		LastUpdated:       time.Now(),
	}

	// Overall stats
	query := `
		SELECT COUNT(*), SUM(CASE WHEN was_correct THEN 1 ELSE 0 END)
		FROM prediction_outcomes
	`
	
	err := s.db.QueryRowContext(ctx, query).Scan(&stats.TotalPredictions, &stats.CorrectPredictions)
	if err != nil {
		return nil, fmt.Errorf("failed to get overall stats: %w", err)
	}

	if stats.TotalPredictions > 0 {
		stats.AccuracyRate = float64(stats.CorrectPredictions) / float64(stats.TotalPredictions)
	}

	// By competition
	if err := s.calculateCompetitionStats(ctx, stats); err != nil {
		slog.Error("Failed to calculate competition stats", "error", err)
	}

	// By confidence range
	if err := s.calculateConfidenceStats(ctx, stats); err != nil {
		slog.Error("Failed to calculate confidence stats", "error", err)
	}

	return stats, nil
}

// calculateCompetitionStats calculates accuracy by competition
func (s *AccuracyService) calculateCompetitionStats(ctx context.Context, stats *AccuracyStats) error {
	query := `
		SELECT 
			competition_id,
			competition_name,
			COUNT(*) as total,
			SUM(CASE WHEN was_correct THEN 1 ELSE 0 END) as correct
		FROM prediction_outcomes
		GROUP BY competition_id, competition_name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var compID int
		var compName string
		var total, correct int

		if err := rows.Scan(&compID, &compName, &total, &correct); err != nil {
			continue
		}

		acc := &CompetitionAcc{
			CompetitionID:      compID,
			CompetitionName:    compName,
			TotalPredictions:   total,
			CorrectPredictions: correct,
		}
		if total > 0 {
			acc.AccuracyRate = float64(correct) / float64(total)
		}

		stats.ByCompetition[compName] = acc
	}

	return nil
}

// calculateConfidenceStats calculates accuracy by confidence ranges
func (s *AccuracyService) calculateConfidenceStats(ctx context.Context, stats *AccuracyStats) error {
	ranges := []struct {
		name string
		min  float64
		max  float64
	}{
		{"0.0-0.5", 0.0, 0.5},
		{"0.5-0.6", 0.5, 0.6},
		{"0.6-0.7", 0.6, 0.7},
		{"0.7-0.8", 0.7, 0.8},
		{"0.8-0.9", 0.8, 0.9},
		{"0.9-1.0", 0.9, 1.0},
	}

	for _, r := range ranges {
		query := `
			SELECT 
				COUNT(*) as total,
				SUM(CASE WHEN was_correct THEN 1 ELSE 0 END) as correct
			FROM prediction_outcomes
			WHERE confidence_score >= $1 AND confidence_score < $2
		`

		var total, correct int
		err := s.db.QueryRowContext(ctx, query, r.min, r.max).Scan(&total, &correct)
		if err != nil {
			continue
		}

		if total > 0 {
			acc := &RangeAcc{
				Range:              r.name,
				TotalPredictions:   total,
				CorrectPredictions: correct,
			}
			acc.AccuracyRate = float64(correct) / float64(total)
			stats.ByConfidenceRange[r.name] = acc
		}
	}

	return nil
}

// GetCompetitionStats gets accuracy stats for a specific competition
func (s *AccuracyService) GetCompetitionStats(ctx context.Context, competitionID int) (*CompetitionAcc, error) {
	query := `
		SELECT 
			competition_id,
			competition_name,
			COUNT(*) as total,
			SUM(CASE WHEN was_correct THEN 1 ELSE 0 END) as correct
		FROM prediction_outcomes
		WHERE competition_id = $1
		GROUP BY competition_id, competition_name
	`

	var acc CompetitionAcc
	err := s.db.QueryRowContext(ctx, query, competitionID).Scan(
		&acc.CompetitionID,
		&acc.CompetitionName,
		&acc.TotalPredictions,
		&acc.CorrectPredictions,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no predictions found for competition")
		}
		return nil, err
	}

	if acc.TotalPredictions > 0 {
		acc.AccuracyRate = float64(acc.CorrectPredictions) / float64(acc.TotalPredictions)
	}

	return &acc, nil
}

// GetLeaderboard gets a leaderboard of providers and agents
func (s *AccuracyService) GetLeaderboard(ctx context.Context) ([]LeaderboardEntry, error) {
	// For now, return empty as we need to extend the schema to track provider/agent per outcome
	// This would require storing agent outputs with provider information
	return []LeaderboardEntry{}, nil
}

// CheckCompletedMatches checks for completed matches and records outcomes
func (s *AccuracyService) CheckCompletedMatches(ctx context.Context) error {
	// Find predictions for completed matches that don't have outcomes yet
	query := `
		SELECT p.id, p.match_id
		FROM predictions p
		JOIN matches m ON p.match_id = m.id
		LEFT JOIN prediction_outcomes po ON p.id = po.prediction_id
		WHERE m.status = 'FINISHED' AND po.id IS NULL
		LIMIT 100
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query completed matches: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var predictionID uuid.UUID
		var matchID int

		if err := rows.Scan(&predictionID, &matchID); err != nil {
			slog.Error("Failed to scan prediction", "error", err)
			continue
		}

		if err := s.RecordOutcome(ctx, predictionID, matchID); err != nil {
			slog.Error("Failed to record outcome", "predictionId", predictionID, "error", err)
			continue
		}

		count++
	}

	if count > 0 {
		slog.Info("Recorded prediction outcomes", "count", count)
	}

	return nil
}
