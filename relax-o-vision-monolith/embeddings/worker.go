package embeddings

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/edd/relaxovisionmonolith/footballdata"
)

// Worker handles background embedding population
type Worker struct {
	service         *Service
	db              *sql.DB
	footballService *footballdata.Service
	batchSize       int
	interval        time.Duration
}

// NewWorker creates a new embedding worker
func NewWorker(service *Service, db *sql.DB, footballService *footballdata.Service) *Worker {
	return &Worker{
		service:         service,
		db:              db,
		footballService: footballService,
		batchSize:       10,
		interval:        5 * time.Minute,
	}
}

// Start starts the background worker
func (w *Worker) Start(ctx context.Context) {
	slog.Info("Starting embedding worker")

	// Do initial population
	w.populateTeamEmbeddings(ctx)
	w.populateCompetitionEmbeddings(ctx)

	// Set up periodic updates
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping embedding worker")
			return
		case <-ticker.C:
			w.populateTeamEmbeddings(ctx)
			w.populateCompetitionEmbeddings(ctx)
		}
	}
}

// populateTeamEmbeddings populates embeddings for teams without them
func (w *Worker) populateTeamEmbeddings(ctx context.Context) {
	slog.Info("Populating team embeddings")

	query := `
		SELECT id, name, short_name, tla, crest, address, website, founded, club_colors, venue, last_updated
		FROM teams
		WHERE embedding IS NULL
		LIMIT $1
	`

	rows, err := w.db.QueryContext(ctx, query, w.batchSize)
	if err != nil {
		slog.Error("Failed to query teams for embedding", "error", err)
		return
	}
	defer rows.Close()

	count := 0
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
			slog.Error("Failed to scan team", "error", err)
			continue
		}

		embedding, err := w.service.GenerateTeamEmbedding(ctx, &team)
		if err != nil {
			slog.Error("Failed to generate team embedding", "teamId", team.ID, "error", err)
			continue
		}

		if err := w.service.SaveTeamEmbedding(ctx, team.ID, embedding); err != nil {
			slog.Error("Failed to save team embedding", "teamId", team.ID, "error", err)
			continue
		}

		count++
		slog.Info("Generated team embedding", "teamId", team.ID, "teamName", team.Name)
	}

	if count > 0 {
		slog.Info("Populated team embeddings", "count", count)
	}
}

// populateCompetitionEmbeddings populates embeddings for competitions without them
func (w *Worker) populateCompetitionEmbeddings(ctx context.Context) {
	slog.Info("Populating competition embeddings")

	query := `
		SELECT id, name, code, type, emblem, area, current_season, seasons
		FROM competitions
		WHERE embedding IS NULL
		LIMIT $1
	`

	rows, err := w.db.QueryContext(ctx, query, w.batchSize)
	if err != nil {
		slog.Error("Failed to query competitions for embedding", "error", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var comp footballdata.Competition
		var areaJSON, currentSeasonJSON, seasonsJSON []byte

		err := rows.Scan(
			&comp.ID,
			&comp.Name,
			&comp.Code,
			&comp.Type,
			&comp.Emblem,
			&areaJSON,
			&currentSeasonJSON,
			&seasonsJSON,
		)
		if err != nil {
			slog.Error("Failed to scan competition", "error", err)
			continue
		}

		embedding, err := w.service.GenerateCompetitionEmbedding(ctx, &comp)
		if err != nil {
			slog.Error("Failed to generate competition embedding", "compId", comp.ID, "error", err)
			continue
		}

		if err := w.service.SaveCompetitionEmbedding(ctx, comp.ID, embedding); err != nil {
			slog.Error("Failed to save competition embedding", "compId", comp.ID, "error", err)
			continue
		}

		count++
		slog.Info("Generated competition embedding", "compId", comp.ID, "compName", comp.Name)
	}

	if count > 0 {
		slog.Info("Populated competition embeddings", "count", count)
	}
}
