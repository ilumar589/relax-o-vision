package footballdata

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Scheduler handles periodic data synchronization
type Scheduler struct {
	service           *Service
	cacheManager      *CacheManager
	competitionCodes  []string
	syncInterval      time.Duration
	stopChan          chan struct{}
}

// NewScheduler creates a new scheduler instance
func NewScheduler(service *Service, cacheManager *CacheManager, competitionCodes []string, syncInterval time.Duration) *Scheduler {
	return &Scheduler{
		service:          service,
		cacheManager:     cacheManager,
		competitionCodes: competitionCodes,
		syncInterval:     syncInterval,
		stopChan:         make(chan struct{}),
	}
}

// Start begins the background synchronization process
func (s *Scheduler) Start(ctx context.Context) {
	slog.Info("Starting football data scheduler", "interval", s.syncInterval)

	// Run initial sync
	s.runSync(ctx)

	// Schedule periodic syncs
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.runSync(ctx)
		case <-s.stopChan:
			slog.Info("Stopping football data scheduler")
			return
		case <-ctx.Done():
			slog.Info("Context cancelled, stopping scheduler")
			return
		}
	}
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	close(s.stopChan)
}

// runSync executes the synchronization process with freshness checks
func (s *Scheduler) runSync(ctx context.Context) {
	slog.Info("Running scheduled sync")

	// Get all competitions (from API or cache)
	competitions, err := s.getAllCompetitions(ctx)
	if err != nil {
		slog.Error("Failed to get competitions", "error", err)
		return
	}

	slog.Info("Processing competitions", "count", len(competitions))

	// Process each competition
	for _, comp := range competitions {
		// Check if competition data needs refresh
		if s.needsRefresh(ctx, "competition", comp.Code) {
			slog.Info("Syncing competition", "code", comp.Code, "name", comp.Name)
			if err := s.syncCompetition(ctx, comp.Code); err != nil {
				slog.Error("Failed to sync competition", "code", comp.Code, "error", err)
			}
			time.Sleep(2 * time.Second) // Rate limiting
		}

		// Check if matches need refresh
		if s.needsRefresh(ctx, "matches", comp.Code) {
			slog.Info("Syncing matches", "code", comp.Code)
			if err := s.service.SyncCompetitionMatches(ctx, comp.Code); err != nil {
				slog.Error("Failed to sync matches", "code", comp.Code, "error", err)
			}
			
			// Update cache metadata for matches
			if s.cacheManager != nil {
				s.cacheManager.SetMetadata(ctx, "matches", comp.Code, "")
			}
			
			time.Sleep(2 * time.Second) // Rate limiting
		}

		// Check if standings need refresh
		if s.needsRefresh(ctx, "standings", comp.Code) {
			slog.Info("Syncing standings", "code", comp.Code)
			if err := s.syncStandings(ctx, comp.Code); err != nil {
				slog.Error("Failed to sync standings", "code", comp.Code, "error", err)
			}
			
			// Update cache metadata for standings
			if s.cacheManager != nil {
				s.cacheManager.SetMetadata(ctx, "standings", comp.Code, "")
			}
			
			time.Sleep(2 * time.Second) // Rate limiting
		}
	}

	slog.Info("Completed scheduled sync")
}

// getAllCompetitions fetches all available competitions
func (s *Scheduler) getAllCompetitions(ctx context.Context) ([]Competition, error) {
	// If competition codes are configured, use them
	if len(s.competitionCodes) > 0 {
		var competitions []Competition
		for _, code := range s.competitionCodes {
			comp := Competition{Code: code}
			competitions = append(competitions, comp)
		}
		return competitions, nil
	}

	// Otherwise, fetch all competitions from the service
	return s.service.GetAllCompetitions(ctx)
}

// needsRefresh checks if data needs refresh based on cache metadata
func (s *Scheduler) needsRefresh(ctx context.Context, entityType, entityKey string) bool {
	if s.cacheManager == nil {
		return true // No cache manager, always refresh
	}

	return s.cacheManager.NeedsRefresh(ctx, entityType, entityKey)
}

// syncCompetition syncs a specific competition
func (s *Scheduler) syncCompetition(ctx context.Context, code string) error {
	// Fetch competition data from API
	comp, err := s.service.GetClient().GetCompetition(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to fetch competition: %w", err)
	}

	// Save to database
	if err := s.service.GetRepository().SaveCompetition(ctx, comp); err != nil {
		return fmt.Errorf("failed to save competition: %w", err)
	}

	// Update cache metadata
	if s.cacheManager != nil {
		dataHash := ComputeDataHash(comp)
		return s.cacheManager.SetMetadata(ctx, "competition", code, dataHash)
	}

	return nil
}

// syncStandings syncs standings for a competition
func (s *Scheduler) syncStandings(ctx context.Context, code string) error {
	// Fetch standings data from API
	standings, err := s.service.GetClient().GetStandings(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to fetch standings: %w", err)
	}

	// Note: Standings are not persisted to a dedicated table currently.
	// They can be cached in Redis for fast access.
	// Update cache metadata
	if s.cacheManager != nil {
		dataHash := ComputeDataHash(standings)
		return s.cacheManager.SetMetadata(ctx, "standings", code, dataHash)
	}

	return nil
}
