package footballdata

import (
	"testing"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	client := NewClient("test-api-key")
	mockRepo := NewMockRepository()
	
	// Can't easily create service with mock repo due to concrete types
	// This would work better with interfaces
	_ = client
	_ = mockRepo
	
	t.Skip("Service requires concrete Repository type - needs refactoring for better testing")
}

func TestService_SyncCompetitions(t *testing.T) {
	t.Skip("Integration test - requires API and database")
}

func TestService_SyncCompetitionMatches(t *testing.T) {
	t.Skip("Integration test - requires API and database")
}

func TestService_GetCompetition(t *testing.T) {
	t.Skip("Integration test - requires database")
}

func TestService_GetTeam(t *testing.T) {
	t.Skip("Integration test - requires database")
}

func TestService_GetMatch(t *testing.T) {
	t.Skip("Integration test - requires database")
}

func TestService_GetAllCompetitions(t *testing.T) {
	t.Skip("Integration test - requires API")
}

func TestService_ErrorHandling(t *testing.T) {
	t.Skip("Integration test - requires API error simulation")
}

func TestService_SaveTeamsFromMatches(t *testing.T) {
	t.Skip("Integration test - verify teams are saved when syncing matches")
}

func TestService_LoggingOfSync(t *testing.T) {
	t.Skip("Integration test - verify logging of sync operations")
}
