package predictions

import (
	"time"
)

// PredictionRequest represents a request for match prediction
type PredictionRequest struct {
	MatchID int `json:"matchId"`
}

// AgentOutput represents the output from a single AI agent
type AgentOutput struct {
	AgentType   string             `json:"agentType"`
	HomeWinProb float64            `json:"homeWinProb"`
	DrawProb    float64            `json:"drawProb"`
	AwayWinProb float64            `json:"awayWinProb"`
	Confidence  float64            `json:"confidence"`
	Reasoning   string             `json:"reasoning"`
	KeyFactors  []string           `json:"keyFactors"`
	Metadata    map[string]any     `json:"metadata,omitempty"`
}

// PredictionResult represents the final prediction output
type PredictionResult struct {
	ID           string        `json:"id"`
	MatchID      int           `json:"matchId"`
	HomeWinProb  float64       `json:"homeWinProb"`
	DrawProb     float64       `json:"drawProb"`
	AwayWinProb  float64       `json:"awayWinProb"`
	Confidence   float64       `json:"confidence"`
	Status       string        `json:"status"`
	WorkflowID   string        `json:"workflowId"`
	AgentOutputs []AgentOutput `json:"agentOutputs"`
	Reasoning    string        `json:"reasoning"`
	KeyFactors   []string      `json:"keyFactors"`
	CreatedAt    time.Time     `json:"createdAt"`
	UpdatedAt    time.Time     `json:"updatedAt"`
}

// WorkflowInput represents input data for the prediction workflow
type WorkflowInput struct {
	MatchID int `json:"matchId"`
}

// WorkflowOutput represents output from the prediction workflow
type WorkflowOutput struct {
	HomeWinProb  float64       `json:"homeWinProb"`
	DrawProb     float64       `json:"drawProb"`
	AwayWinProb  float64       `json:"awayWinProb"`
	Confidence   float64       `json:"confidence"`
	Reasoning    string        `json:"reasoning"`
	AgentOutputs []AgentOutput `json:"agentOutputs"`
}

// MatchAnalysis represents data about a match for analysis
type MatchAnalysis struct {
	MatchID       int                    `json:"matchId"`
	HomeTeam      TeamAnalysis           `json:"homeTeam"`
	AwayTeam      TeamAnalysis           `json:"awayTeam"`
	Competition   string                 `json:"competition"`
	MatchDate     time.Time              `json:"matchDate"`
	HeadToHead    []HistoricalMatch      `json:"headToHead"`
	Metadata      map[string]any         `json:"metadata,omitempty"`
}

// TeamAnalysis represents team data for prediction analysis
type TeamAnalysis struct {
	ID            int                    `json:"id"`
	Name          string                 `json:"name"`
	RecentForm    []string               `json:"recentForm"` // W, D, L for last 5 games
	Statistics    TeamStatistics         `json:"statistics"`
	CurrentForm   string                 `json:"currentForm"`
}

// TeamStatistics represents team performance statistics
type TeamStatistics struct {
	GoalsScored     int     `json:"goalsScored"`
	GoalsConceded   int     `json:"goalsConceded"`
	MatchesPlayed   int     `json:"matchesPlayed"`
	Wins            int     `json:"wins"`
	Draws           int     `json:"draws"`
	Losses          int     `json:"losses"`
	GoalDifference  int     `json:"goalDifference"`
	AvgGoalsScored  float64 `json:"avgGoalsScored"`
	AvgConceded     float64 `json:"avgConceded"`
}

// HistoricalMatch represents a past match between two teams
type HistoricalMatch struct {
	Date         time.Time `json:"date"`
	HomeTeamID   int       `json:"homeTeamId"`
	AwayTeamID   int       `json:"awayTeamId"`
	HomeScore    int       `json:"homeScore"`
	AwayScore    int       `json:"awayScore"`
	Competition  string    `json:"competition"`
}
