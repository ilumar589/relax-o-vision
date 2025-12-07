package footballdata

import (
	"context"
	"database/sql"
	"fmt"
)

// TeamForm represents recent form analysis for a team
type TeamForm struct {
	TeamID            int      `json:"teamId"`
	TeamName          string   `json:"teamName"`
	Last5Results      []string `json:"last5Results"` // ["W", "D", "L", "W", "W"]
	Last5GoalsFor     int      `json:"last5GoalsFor"`
	Last5GoalsAgainst int      `json:"last5GoalsAgainst"`
	HomeForm          float64  `json:"homeForm"`   // Points per game at home
	AwayForm          float64  `json:"awayForm"`   // Points per game away
	GoalScoringTrend  float64  `json:"goalScoringTrend"`
	DefensiveTrend    float64  `json:"defensiveTrend"`
	FormScore         float64  `json:"formScore"` // Weighted composite
}

// FormAnalyzer calculates team form
type FormAnalyzer struct {
	db *sql.DB
}

// NewFormAnalyzer creates a new form analyzer
func NewFormAnalyzer(db *sql.DB) *FormAnalyzer {
	return &FormAnalyzer{db: db}
}

// AnalyzeTeamForm analyzes recent form for a team
func (f *FormAnalyzer) AnalyzeTeamForm(ctx context.Context, teamID int) (*TeamForm, error) {
	form := &TeamForm{
		TeamID: teamID,
	}

	// Get team name
	err := f.db.QueryRowContext(ctx, `
		SELECT (data->>'name')::text
		FROM teams
		WHERE id = $1
	`, teamID).Scan(&form.TeamName)
	if err != nil {
		return nil, fmt.Errorf("failed to get team name: %w", err)
	}

	// Get last 5 matches (simplified - in real implementation would parse JSON)
	// For now, return placeholder data
	form.Last5Results = []string{"W", "D", "W", "L", "W"}
	form.Last5GoalsFor = 8
	form.Last5GoalsAgainst = 5
	form.HomeForm = 2.1
	form.AwayForm = 1.5
	form.GoalScoringTrend = 0.3  // Increasing
	form.DefensiveTrend = -0.1   // Slightly improving
	form.FormScore = 0.72

	return form, nil
}

// CalculateFormScore calculates a weighted form score
func (f *FormAnalyzer) CalculateFormScore(results []string) float64 {
	if len(results) == 0 {
		return 0
	}

	score := 0.0
	weight := 1.0

	// More recent matches have higher weight
	for i := len(results) - 1; i >= 0; i-- {
		switch results[i] {
		case "W":
			score += 3.0 * weight
		case "D":
			score += 1.0 * weight
		case "L":
			score += 0.0
		}
		weight *= 0.8 // Decay weight for older matches
	}

	// Normalize to 0-1 range
	maxScore := 3.0 * (1.0 + 0.8 + 0.64 + 0.512 + 0.4096)
	return score / maxScore
}
