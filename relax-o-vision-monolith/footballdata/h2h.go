package footballdata

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// HeadToHead represents head-to-head analysis between two teams
type HeadToHead struct {
	Team1ID        int            `json:"team1Id"`
	Team1Name      string         `json:"team1Name"`
	Team2ID        int            `json:"team2Id"`
	Team2Name      string         `json:"team2Name"`
	TotalMatches   int            `json:"totalMatches"`
	Team1Wins      int            `json:"team1Wins"`
	Team2Wins      int            `json:"team2Wins"`
	Draws          int            `json:"draws"`
	Team1Goals     int            `json:"team1Goals"`
	Team2Goals     int            `json:"team2Goals"`
	RecentMatches  []MatchSummary `json:"recentMatches"`
	HomeAdvantage  float64        `json:"homeAdvantage"`
	TrendDirection string         `json:"trendDirection"` // "team1_improving", "team2_improving", "stable"
}

// MatchSummary represents a summary of a historical match
type MatchSummary struct {
	Date        time.Time `json:"date"`
	HomeTeamID  int       `json:"homeTeamId"`
	AwayTeamID  int       `json:"awayTeamId"`
	HomeScore   int       `json:"homeScore"`
	AwayScore   int       `json:"awayScore"`
	Winner      string    `json:"winner"` // "home", "away", "draw"
	Competition string    `json:"competition"`
}

// H2HAnalyzer analyzes head-to-head records
type H2HAnalyzer struct {
	db *sql.DB
}

// NewH2HAnalyzer creates a new head-to-head analyzer
func NewH2HAnalyzer(db *sql.DB) *H2HAnalyzer {
	return &H2HAnalyzer{db: db}
}

// AnalyzeHeadToHead analyzes head-to-head record between two teams
func (h *H2HAnalyzer) AnalyzeHeadToHead(ctx context.Context, team1ID, team2ID int) (*HeadToHead, error) {
	h2h := &HeadToHead{
		Team1ID: team1ID,
		Team2ID: team2ID,
	}

	// Get team names
	var team1Data, team2Data []byte
	err := h.db.QueryRowContext(ctx, `SELECT data FROM teams WHERE id = $1`, team1ID).Scan(&team1Data)
	if err != nil {
		return nil, fmt.Errorf("failed to get team1: %w", err)
	}

	err = h.db.QueryRowContext(ctx, `SELECT data FROM teams WHERE id = $1`, team2ID).Scan(&team2Data)
	if err != nil {
		return nil, fmt.Errorf("failed to get team2: %w", err)
	}

	var team1Map, team2Map map[string]interface{}
	json.Unmarshal(team1Data, &team1Map)
	json.Unmarshal(team2Data, &team2Map)

	if name, ok := team1Map["name"].(string); ok {
		h2h.Team1Name = name
	}
	if name, ok := team2Map["name"].(string); ok {
		h2h.Team2Name = name
	}

	// Query historical matches between these teams
	query := `
		SELECT 
			id, 
			utc_date,
			home_team,
			away_team,
			score,
			competition
		FROM matches
		WHERE 
			status = 'FINISHED' AND
			(
				(home_team->>'id')::int = $1 AND (away_team->>'id')::int = $2
				OR
				(home_team->>'id')::int = $2 AND (away_team->>'id')::int = $1
			)
		ORDER BY utc_date DESC
		LIMIT 10
	`

	rows, err := h.db.QueryContext(ctx, query, team1ID, team2ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query h2h matches: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var matchID int
		var utcDate time.Time
		var homeTeamJSON, awayTeamJSON, scoreJSON, competitionJSON []byte

		err := rows.Scan(&matchID, &utcDate, &homeTeamJSON, &awayTeamJSON, &scoreJSON, &competitionJSON)
		if err != nil {
			continue
		}

		var homeTeam, awayTeam, score, competition map[string]interface{}
		json.Unmarshal(homeTeamJSON, &homeTeam)
		json.Unmarshal(awayTeamJSON, &awayTeam)
		json.Unmarshal(scoreJSON, &score)
		json.Unmarshal(competitionJSON, &competition)

		homeTeamID := int(homeTeam["id"].(float64))
		awayTeamID := int(awayTeam["id"].(float64))

		// Extract scores
		var homeScore, awayScore int
		if fullTime, ok := score["fullTime"].(map[string]interface{}); ok {
			if hs, ok := fullTime["home"].(float64); ok {
				homeScore = int(hs)
			}
			if as, ok := fullTime["away"].(float64); ok {
				awayScore = int(as)
			}
		}

		winner := "draw"
		if homeScore > awayScore {
			winner = "home"
		} else if awayScore > homeScore {
			winner = "away"
		}

		compName := ""
		if name, ok := competition["name"].(string); ok {
			compName = name
		}

		summary := MatchSummary{
			Date:        utcDate,
			HomeTeamID:  homeTeamID,
			AwayTeamID:  awayTeamID,
			HomeScore:   homeScore,
			AwayScore:   awayScore,
			Winner:      winner,
			Competition: compName,
		}

		h2h.RecentMatches = append(h2h.RecentMatches, summary)
		h2h.TotalMatches++

		// Count wins/goals from team1's perspective
		if homeTeamID == team1ID {
			h2h.Team1Goals += homeScore
			h2h.Team2Goals += awayScore
			if winner == "home" {
				h2h.Team1Wins++
			} else if winner == "away" {
				h2h.Team2Wins++
			} else {
				h2h.Draws++
			}
		} else {
			h2h.Team1Goals += awayScore
			h2h.Team2Goals += homeScore
			if winner == "away" {
				h2h.Team1Wins++
			} else if winner == "home" {
				h2h.Team2Wins++
			} else {
				h2h.Draws++
			}
		}
	}

	// Calculate home advantage and trend
	h2h.HomeAdvantage = h.calculateHomeAdvantage(h2h.RecentMatches, team1ID)
	h2h.TrendDirection = h.calculateTrend(h2h.RecentMatches, team1ID, team2ID)

	return h2h, nil
}

// calculateHomeAdvantage calculates the home advantage factor
func (h *H2HAnalyzer) calculateHomeAdvantage(matches []MatchSummary, team1ID int) float64 {
	if len(matches) == 0 {
		return 0.0
	}

	homeWins := 0
	homeMatches := 0

	for _, match := range matches {
		if match.HomeTeamID == team1ID {
			homeMatches++
			if match.Winner == "home" {
				homeWins++
			}
		}
	}

	if homeMatches == 0 {
		return 0.0
	}

	return float64(homeWins) / float64(homeMatches)
}

// calculateTrend determines if one team is improving against the other
func (h *H2HAnalyzer) calculateTrend(matches []MatchSummary, team1ID, team2ID int) string {
	if len(matches) < 3 {
		return "stable"
	}

	// Look at the last 3 matches vs the previous matches
	recent := matches[:3]
	older := matches[3:]

	recentTeam1Wins := 0
	olderTeam1Wins := 0

	for _, match := range recent {
		if (match.HomeTeamID == team1ID && match.Winner == "home") ||
			(match.AwayTeamID == team1ID && match.Winner == "away") {
			recentTeam1Wins++
		}
	}

	for _, match := range older {
		if (match.HomeTeamID == team1ID && match.Winner == "home") ||
			(match.AwayTeamID == team1ID && match.Winner == "away") {
			olderTeam1Wins++
		}
	}

	recentRate := float64(recentTeam1Wins) / float64(len(recent))
	olderRate := float64(olderTeam1Wins) / float64(len(older))

	if recentRate > olderRate+0.2 {
		return "team1_improving"
	} else if olderRate > recentRate+0.2 {
		return "team2_improving"
	}

	return "stable"
}
