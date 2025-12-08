package footballdata

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pgvector/pgvector-go"
)

// Repository handles database operations for football data
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SaveCompetition saves or updates a competition in the database
func (r *Repository) SaveCompetition(ctx context.Context, comp *Competition) error {
	areaJSON, err := json.Marshal(comp.Area)
	if err != nil {
		return fmt.Errorf("failed to marshal area: %w", err)
	}

	currentSeasonJSON, err := json.Marshal(comp.CurrentSeason)
	if err != nil {
		return fmt.Errorf("failed to marshal current season: %w", err)
	}

	seasonsJSON, err := json.Marshal(comp.Seasons)
	if err != nil {
		return fmt.Errorf("failed to marshal seasons: %w", err)
	}

	query := `
		INSERT INTO competitions (id, code, name, type, emblem, area, current_season, seasons, updated_at, cached_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			code = EXCLUDED.code,
			name = EXCLUDED.name,
			type = EXCLUDED.type,
			emblem = EXCLUDED.emblem,
			area = EXCLUDED.area,
			current_season = EXCLUDED.current_season,
			seasons = EXCLUDED.seasons,
			updated_at = EXCLUDED.updated_at,
			cached_at = EXCLUDED.cached_at
	`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query,
		comp.ID,
		comp.Code,
		comp.Name,
		comp.Type,
		comp.Emblem,
		areaJSON,
		currentSeasonJSON,
		seasonsJSON,
		now,
		now, // cached_at
	)
	if err != nil {
		return fmt.Errorf("failed to save competition: %w", err)
	}

	return nil
}

// GetCompetition retrieves a competition by ID
func (r *Repository) GetCompetition(ctx context.Context, id int) (*Competition, error) {
	query := `
		SELECT id, code, name, type, emblem, area, current_season, seasons
		FROM competitions
		WHERE id = $1
	`

	var comp Competition
	var areaJSON, currentSeasonJSON, seasonsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comp.ID,
		&comp.Code,
		&comp.Name,
		&comp.Type,
		&comp.Emblem,
		&areaJSON,
		&currentSeasonJSON,
		&seasonsJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("competition not found")
		}
		return nil, fmt.Errorf("failed to get competition: %w", err)
	}

	if err := json.Unmarshal(areaJSON, &comp.Area); err != nil {
		return nil, fmt.Errorf("failed to unmarshal area: %w", err)
	}

	if err := json.Unmarshal(currentSeasonJSON, &comp.CurrentSeason); err != nil {
		return nil, fmt.Errorf("failed to unmarshal current season: %w", err)
	}

	if err := json.Unmarshal(seasonsJSON, &comp.Seasons); err != nil {
		return nil, fmt.Errorf("failed to unmarshal seasons: %w", err)
	}

	return &comp, nil
}

// SaveTeam saves or updates a team in the database
func (r *Repository) SaveTeam(ctx context.Context, team *Team) error {
	// Create a minimal area representation for the team
	areaJSON, err := json.Marshal(map[string]any{})
	if err != nil {
		return fmt.Errorf("failed to marshal area: %w", err)
	}

	query := `
		INSERT INTO teams (id, name, short_name, tla, crest, address, website, founded, club_colors, venue, area, updated_at, cached_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			short_name = EXCLUDED.short_name,
			tla = EXCLUDED.tla,
			crest = EXCLUDED.crest,
			address = EXCLUDED.address,
			website = EXCLUDED.website,
			founded = EXCLUDED.founded,
			club_colors = EXCLUDED.club_colors,
			venue = EXCLUDED.venue,
			area = EXCLUDED.area,
			updated_at = EXCLUDED.updated_at,
			cached_at = EXCLUDED.cached_at
	`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query,
		team.ID,
		team.Name,
		team.ShortName,
		team.TLA,
		team.Crest,
		team.Address,
		team.Website,
		team.Founded,
		team.ClubColors,
		team.Venue,
		areaJSON,
		now,
		now, // cached_at
	)
	if err != nil {
		return fmt.Errorf("failed to save team: %w", err)
	}

	return nil
}

// GetTeam retrieves a team by ID
func (r *Repository) GetTeam(ctx context.Context, id int) (*Team, error) {
	query := `
		SELECT id, name, short_name, tla, crest, address, website, founded, club_colors, venue
		FROM teams
		WHERE id = $1
	`

	var team Team
	err := r.db.QueryRowContext(ctx, query, id).Scan(
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
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return &team, nil
}

// SaveMatch saves or updates a match in the database
func (r *Repository) SaveMatch(ctx context.Context, match *Match) error {
	homeTeamJSON, err := json.Marshal(match.HomeTeam)
	if err != nil {
		return fmt.Errorf("failed to marshal home team: %w", err)
	}

	awayTeamJSON, err := json.Marshal(match.AwayTeam)
	if err != nil {
		return fmt.Errorf("failed to marshal away team: %w", err)
	}

	scoreJSON, err := json.Marshal(match.Score)
	if err != nil {
		return fmt.Errorf("failed to marshal score: %w", err)
	}

	oddsJSON, err := json.Marshal(match.Odds)
	if err != nil {
		return fmt.Errorf("failed to marshal odds: %w", err)
	}

	refereesJSON, err := json.Marshal(match.Referees)
	if err != nil {
		return fmt.Errorf("failed to marshal referees: %w", err)
	}

	query := `
		INSERT INTO matches (id, competition_id, season_id, matchday, status, utc_date, home_team, away_team, score, odds, referees, updated_at, cached_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			competition_id = EXCLUDED.competition_id,
			season_id = EXCLUDED.season_id,
			matchday = EXCLUDED.matchday,
			status = EXCLUDED.status,
			utc_date = EXCLUDED.utc_date,
			home_team = EXCLUDED.home_team,
			away_team = EXCLUDED.away_team,
			score = EXCLUDED.score,
			odds = EXCLUDED.odds,
			referees = EXCLUDED.referees,
			updated_at = EXCLUDED.updated_at,
			cached_at = EXCLUDED.cached_at
	`

	now := time.Now()
	_, err = r.db.ExecContext(ctx, query,
		match.ID,
		match.Competition.ID,
		match.Season.ID,
		match.Matchday,
		match.Status,
		match.UTCDate,
		homeTeamJSON,
		awayTeamJSON,
		scoreJSON,
		oddsJSON,
		refereesJSON,
		now,
		now, // cached_at
	)
	if err != nil {
		return fmt.Errorf("failed to save match: %w", err)
	}

	return nil
}

// GetMatch retrieves a match by ID
func (r *Repository) GetMatch(ctx context.Context, id int) (*Match, error) {
	query := `
		SELECT id, competition_id, season_id, matchday, status, utc_date, home_team, away_team, score, odds, referees
		FROM matches
		WHERE id = $1
	`

	var match Match
	var homeTeamJSON, awayTeamJSON, scoreJSON, oddsJSON, refereesJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&match.ID,
		&match.CompetitionID,
		&match.Season.ID,
		&match.Matchday,
		&match.Status,
		&match.UTCDate,
		&homeTeamJSON,
		&awayTeamJSON,
		&scoreJSON,
		&oddsJSON,
		&refereesJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("match not found")
		}
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	if err := json.Unmarshal(homeTeamJSON, &match.HomeTeam); err != nil {
		return nil, fmt.Errorf("failed to unmarshal home team: %w", err)
	}

	if err := json.Unmarshal(awayTeamJSON, &match.AwayTeam); err != nil {
		return nil, fmt.Errorf("failed to unmarshal away team: %w", err)
	}

	if err := json.Unmarshal(scoreJSON, &match.Score); err != nil {
		return nil, fmt.Errorf("failed to unmarshal score: %w", err)
	}

	if len(oddsJSON) > 0 && string(oddsJSON) != "null" {
		if err := json.Unmarshal(oddsJSON, &match.Odds); err != nil {
			return nil, fmt.Errorf("failed to unmarshal odds: %w", err)
		}
	}

	if err := json.Unmarshal(refereesJSON, &match.Referees); err != nil {
		return nil, fmt.Errorf("failed to unmarshal referees: %w", err)
	}

	return &match, nil
}

// UpdateCompetitionEmbedding updates the embedding vector for a competition
func (r *Repository) UpdateCompetitionEmbedding(ctx context.Context, id int, embedding []float32) error {
	query := `UPDATE competitions SET embedding = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, pgvector.NewVector(embedding), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update competition embedding: %w", err)
	}
	return nil
}

// UpdateTeamEmbedding updates the embedding vector for a team
func (r *Repository) UpdateTeamEmbedding(ctx context.Context, id int, embedding []float32) error {
	query := `UPDATE teams SET embedding = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, pgvector.NewVector(embedding), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update team embedding: %w", err)
	}
	return nil
}

// UpdateMatchEmbedding updates the embedding vector for a match
func (r *Repository) UpdateMatchEmbedding(ctx context.Context, id int, embedding []float32) error {
	query := `UPDATE matches SET embedding = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, pgvector.NewVector(embedding), time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update match embedding: %w", err)
	}
	return nil
}
