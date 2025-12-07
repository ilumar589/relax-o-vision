package footballdata

import (
	"testing"
)

// Repository tests require a real PostgreSQL database
// These are integration tests and should be run with testcontainers

func TestRepository_SaveCompetition(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database")
}

func TestRepository_GetCompetition(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database")
}

func TestRepository_SaveTeam(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database")
}

func TestRepository_GetTeam(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database")
}

func TestRepository_SaveMatch(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database")
}

func TestRepository_GetMatch(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database")
}

func TestRepository_UpdateCompetitionEmbedding(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database")
}

func TestRepository_UpdateTeamEmbedding(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database")
}

func TestRepository_UpdateMatchEmbedding(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database")
}

func TestRepository_UpsertBehavior(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database - test INSERT ON CONFLICT")
}

func TestRepository_CachedAtUpdate(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database - verify cached_at is updated")
}

func TestRepository_JSONBHandling(t *testing.T) {
	t.Skip("Integration test - requires PostgreSQL database - test JSONB columns")
}
