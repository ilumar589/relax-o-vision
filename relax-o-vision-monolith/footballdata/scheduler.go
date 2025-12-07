package footballdata

import (
	"context"
	"log/slog"
	"time"
)

// Scheduler handles periodic data synchronization
type Scheduler struct {
	service           *Service
	competitionCodes  []string
	syncInterval      time.Duration
	stopChan          chan struct{}
}

// NewScheduler creates a new scheduler instance
func NewScheduler(service *Service, competitionCodes []string, syncInterval time.Duration) *Scheduler {
	return &Scheduler{
		service:          service,
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

// runSync executes the synchronization process
func (s *Scheduler) runSync(ctx context.Context) {
	slog.Info("Running scheduled sync")

	// Sync competitions
	if err := s.service.SyncCompetitions(ctx); err != nil {
		slog.Error("Failed to sync competitions", "error", err)
	}

	// Sync matches for each configured competition
	for _, code := range s.competitionCodes {
		if err := s.service.SyncCompetitionMatches(ctx, code); err != nil {
			slog.Error("Failed to sync matches", "competition", code, "error", err)
		}
		// Small delay between competitions to respect rate limits
		time.Sleep(2 * time.Second)
	}

	slog.Info("Completed scheduled sync")
}
