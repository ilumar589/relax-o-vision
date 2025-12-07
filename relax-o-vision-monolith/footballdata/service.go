package footballdata

import (
	"context"
	"fmt"
	"log/slog"
)

// Service handles business logic for football data
type Service struct {
	client *Client
	repo   *Repository
}

// NewService creates a new service instance
func NewService(client *Client, repo *Repository) *Service {
	return &Service{
		client: client,
		repo:   repo,
	}
}

// SyncCompetitions fetches and saves all competitions from the API
func (s *Service) SyncCompetitions(ctx context.Context) error {
	slog.Info("Starting competitions sync")
	
	competitions, err := s.client.GetCompetitions(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch competitions: %w", err)
	}

	for _, comp := range competitions {
		if err := s.repo.SaveCompetition(ctx, &comp); err != nil {
			slog.Error("Failed to save competition", "id", comp.ID, "error", err)
			continue
		}
		slog.Debug("Saved competition", "id", comp.ID, "name", comp.Name)
	}

	slog.Info("Completed competitions sync", "count", len(competitions))
	return nil
}

// SyncCompetitionMatches fetches and saves matches for a specific competition
func (s *Service) SyncCompetitionMatches(ctx context.Context, competitionCode string) error {
	slog.Info("Starting matches sync", "competition", competitionCode)
	
	matches, err := s.client.GetMatches(ctx, competitionCode)
	if err != nil {
		return fmt.Errorf("failed to fetch matches: %w", err)
	}

	for _, match := range matches {
		// Save home team
		if err := s.repo.SaveTeam(ctx, &match.HomeTeam); err != nil {
			slog.Error("Failed to save home team", "id", match.HomeTeam.ID, "error", err)
		}

		// Save away team
		if err := s.repo.SaveTeam(ctx, &match.AwayTeam); err != nil {
			slog.Error("Failed to save away team", "id", match.AwayTeam.ID, "error", err)
		}

		// Save match
		if err := s.repo.SaveMatch(ctx, &match); err != nil {
			slog.Error("Failed to save match", "id", match.ID, "error", err)
			continue
		}
		slog.Debug("Saved match", "id", match.ID)
	}

	slog.Info("Completed matches sync", "competition", competitionCode, "count", len(matches))
	return nil
}

// GetCompetition retrieves a competition by ID
func (s *Service) GetCompetition(ctx context.Context, id int) (*Competition, error) {
	return s.repo.GetCompetition(ctx, id)
}

// GetTeam retrieves a team by ID
func (s *Service) GetTeam(ctx context.Context, id int) (*Team, error) {
	return s.repo.GetTeam(ctx, id)
}

// GetMatch retrieves a match by ID
func (s *Service) GetMatch(ctx context.Context, id int) (*Match, error) {
	return s.repo.GetMatch(ctx, id)
}

// GetAllCompetitions fetches all competitions from the API
func (s *Service) GetAllCompetitions(ctx context.Context) ([]Competition, error) {
	slog.Info("Fetching all competitions from API")
	return s.client.GetCompetitions(ctx)
}

// GetClient returns the API client (for scheduler use)
func (s *Service) GetClient() *Client {
	return s.client
}

// GetRepository returns the repository (for scheduler use)
func (s *Service) GetRepository() *Repository {
	return s.repo
}
