package embeddings

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/edd/relaxovisionmonolith/footballdata"
	"github.com/edd/relaxovisionmonolith/predictions/providers"
)

// Service handles embedding generation and semantic search
type Service struct {
	db        *sql.DB
	providers []providers.LLMProvider
}

// NewService creates a new embedding service
func NewService(db *sql.DB, llmProviders []providers.LLMProvider) *Service {
	return &Service{
		db:        db,
		providers: llmProviders,
	}
}

// GenerateTeamEmbedding generates an embedding for a team
func (s *Service) GenerateTeamEmbedding(ctx context.Context, team *footballdata.Team) ([]float32, error) {
	// Create text representation of the team for embedding
	text := fmt.Sprintf("Team: %s (%s)\nVenue: %s\nColors: %s\nFounded: %d\nCity: %s",
		team.Name,
		team.TLA,
		team.Venue,
		team.ClubColors,
		team.Founded,
		team.Address,
	)

	// Use the first available provider that supports embeddings
	for _, provider := range s.providers {
		embedding, err := provider.GenerateEmbedding(ctx, text)
		if err == nil {
			return embedding, nil
		}
		slog.Warn("Provider failed to generate embedding", "provider", provider.Name(), "error", err)
	}

	return nil, fmt.Errorf("no provider could generate embedding")
}

// GenerateMatchEmbedding generates an embedding for a match
func (s *Service) GenerateMatchEmbedding(ctx context.Context, match *footballdata.Match) ([]float32, error) {
	// Create text representation of the match for embedding
	homeTeamJSON, _ := json.Marshal(match.HomeTeam)
	awayTeamJSON, _ := json.Marshal(match.AwayTeam)
	
	var homeTeam, awayTeam map[string]interface{}
	json.Unmarshal(homeTeamJSON, &homeTeam)
	json.Unmarshal(awayTeamJSON, &awayTeam)
	
	homeName := "Unknown"
	awayName := "Unknown"
	if name, ok := homeTeam["name"].(string); ok {
		homeName = name
	}
	if name, ok := awayTeam["name"].(string); ok {
		awayName = name
	}

	text := fmt.Sprintf("Match: %s vs %s\nCompetition: %s\nDate: %s\nStage: %s\nStatus: %s",
		homeName,
		awayName,
		match.Competition.Name,
		match.UTCDate.Format("2006-01-02"),
		match.Stage,
		match.Status,
	)

	// Use the first available provider that supports embeddings
	for _, provider := range s.providers {
		embedding, err := provider.GenerateEmbedding(ctx, text)
		if err == nil {
			return embedding, nil
		}
		slog.Warn("Provider failed to generate embedding", "provider", provider.Name(), "error", err)
	}

	return nil, fmt.Errorf("no provider could generate embedding")
}

// GenerateCompetitionEmbedding generates an embedding for a competition
func (s *Service) GenerateCompetitionEmbedding(ctx context.Context, comp *footballdata.Competition) ([]float32, error) {
	// Create text representation of the competition for embedding
	text := fmt.Sprintf("Competition: %s (%s)\nType: %s\nArea: %s (%s)",
		comp.Name,
		comp.Code,
		comp.Type,
		comp.Area.Name,
		comp.Area.Code,
	)

	// Use the first available provider that supports embeddings
	for _, provider := range s.providers {
		embedding, err := provider.GenerateEmbedding(ctx, text)
		if err == nil {
			return embedding, nil
		}
		slog.Warn("Provider failed to generate embedding", "provider", provider.Name(), "error", err)
	}

	return nil, fmt.Errorf("no provider could generate embedding")
}

// SaveTeamEmbedding saves a team embedding to the database
func (s *Service) SaveTeamEmbedding(ctx context.Context, teamID int, embedding []float32) error {
	query := `UPDATE teams SET embedding = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, embedding, teamID)
	return err
}

// SaveMatchEmbedding saves a match embedding to the database
func (s *Service) SaveMatchEmbedding(ctx context.Context, matchID int, embedding []float32) error {
	query := `UPDATE matches SET embedding = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, embedding, matchID)
	return err
}

// SaveCompetitionEmbedding saves a competition embedding to the database
func (s *Service) SaveCompetitionEmbedding(ctx context.Context, compID int, embedding []float32) error {
	query := `UPDATE competitions SET embedding = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, embedding, compID)
	return err
}

// SearchSimilarTeams finds teams similar to the given query text
func (s *Service) SearchSimilarTeams(ctx context.Context, queryText string, limit int) ([]footballdata.Team, error) {
	// Generate embedding for query
	var queryEmbedding []float32
	for _, provider := range s.providers {
		embedding, err := provider.GenerateEmbedding(ctx, queryText)
		if err == nil {
			queryEmbedding = embedding
			break
		}
	}

	if queryEmbedding == nil {
		return nil, fmt.Errorf("failed to generate query embedding")
	}

	// Search for similar teams using cosine similarity
	query := `
		SELECT id, name, short_name, tla, crest, address, website, founded, club_colors, venue, last_updated
		FROM teams
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, queryEmbedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search teams: %w", err)
	}
	defer rows.Close()

	var teams []footballdata.Team
	for rows.Next() {
		var team footballdata.Team
		err := rows.Scan(
			&team.ID,
			&team.Name,
			&team.ShortName,
			&team.TLA,
			&team.Crest,
			&team.Address,
			&team.Website,
			&team.Founded,
			&team.ClubColors,
			&team.Venue,
			&team.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}
		teams = append(teams, team)
	}

	return teams, nil
}

// FindSimilarTeam finds teams similar to a given team
func (s *Service) FindSimilarTeam(ctx context.Context, teamID int, limit int) ([]footballdata.Team, error) {
	// Get the team's embedding
	var embedding []float32
	query := `SELECT embedding FROM teams WHERE id = $1 AND embedding IS NOT NULL`
	err := s.db.QueryRowContext(ctx, query, teamID).Scan(&embedding)
	if err != nil {
		return nil, fmt.Errorf("failed to get team embedding: %w", err)
	}

	// Search for similar teams
	query = `
		SELECT id, name, short_name, tla, crest, address, website, founded, club_colors, venue, last_updated
		FROM teams
		WHERE id != $1 AND embedding IS NOT NULL
		ORDER BY embedding <=> $2
		LIMIT $3
	`

	rows, err := s.db.QueryContext(ctx, query, teamID, embedding, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar teams: %w", err)
	}
	defer rows.Close()

	var teams []footballdata.Team
	for rows.Next() {
		var team footballdata.Team
		err := rows.Scan(
			&team.ID,
			&team.Name,
			&team.ShortName,
			&team.TLA,
			&team.Crest,
			&team.Address,
			&team.Website,
			&team.Founded,
			&team.ClubColors,
			&team.Venue,
			&team.LastUpdated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}
		teams = append(teams, team)
	}

	return teams, nil
}
