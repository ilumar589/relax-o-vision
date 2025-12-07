package predictions

import (
	"time"

	"github.com/google/uuid"
)

// PredictionOutcome represents the outcome of a prediction after a match completes
type PredictionOutcome struct {
	ID               uuid.UUID `json:"id"`
	PredictionID     uuid.UUID `json:"predictionId"`
	MatchID          int       `json:"matchId"`
	PredictedWinner  string    `json:"predictedWinner"` // "home", "away", "draw"
	ActualWinner     string    `json:"actualWinner"`
	WasCorrect       bool      `json:"wasCorrect"`
	ConfidenceScore  float64   `json:"confidenceScore"`
	HomeWinProb      float64   `json:"homeWinProb"`
	DrawProb         float64   `json:"drawProb"`
	AwayWinProb      float64   `json:"awayWinProb"`
	ActualHomeScore  int       `json:"actualHomeScore"`
	ActualAwayScore  int       `json:"actualAwayScore"`
	CompetitionID    int       `json:"competitionId"`
	CompetitionName  string    `json:"competitionName"`
	Provider         string    `json:"provider,omitempty"`
	AgentType        string    `json:"agentType,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
}

// AccuracyStats represents overall accuracy statistics
type AccuracyStats struct {
	TotalPredictions    int                        `json:"totalPredictions"`
	CorrectPredictions  int                        `json:"correctPredictions"`
	AccuracyRate        float64                    `json:"accuracyRate"`
	ByCompetition       map[string]*CompetitionAcc `json:"byCompetition"`
	ByConfidenceRange   map[string]*RangeAcc       `json:"byConfidenceRange"`
	ByProvider          map[string]*ProviderAcc    `json:"byProvider"`
	ByAgent             map[string]*AgentAcc       `json:"byAgent"`
	LastUpdated         time.Time                  `json:"lastUpdated"`
}

// CompetitionAcc represents accuracy for a competition
type CompetitionAcc struct {
	CompetitionID      int     `json:"competitionId"`
	CompetitionName    string  `json:"competitionName"`
	TotalPredictions   int     `json:"totalPredictions"`
	CorrectPredictions int     `json:"correctPredictions"`
	AccuracyRate       float64 `json:"accuracyRate"`
}

// RangeAcc represents accuracy for a confidence range
type RangeAcc struct {
	Range              string  `json:"range"` // e.g., "0.7-0.8"
	TotalPredictions   int     `json:"totalPredictions"`
	CorrectPredictions int     `json:"correctPredictions"`
	AccuracyRate       float64 `json:"accuracyRate"`
}

// ProviderAcc represents accuracy for an LLM provider
type ProviderAcc struct {
	ProviderName       string  `json:"providerName"`
	TotalPredictions   int     `json:"totalPredictions"`
	CorrectPredictions int     `json:"correctPredictions"`
	AccuracyRate       float64 `json:"accuracyRate"`
}

// AgentAcc represents accuracy for an agent type
type AgentAcc struct {
	AgentType          string  `json:"agentType"`
	TotalPredictions   int     `json:"totalPredictions"`
	CorrectPredictions int     `json:"correctPredictions"`
	AccuracyRate       float64 `json:"accuracyRate"`
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	Name               string  `json:"name"`
	Type               string  `json:"type"` // "provider" or "agent"
	TotalPredictions   int     `json:"totalPredictions"`
	CorrectPredictions int     `json:"correctPredictions"`
	AccuracyRate       float64 `json:"accuracyRate"`
	Rank               int     `json:"rank"`
}
