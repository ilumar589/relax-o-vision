package footballdata

import (
	"time"
)

// Test fixtures and helper functions

// NewTestCompetition creates a test competition
func NewTestCompetition(id int, code, name string) *Competition {
	return &Competition{
		ID:   id,
		Code: code,
		Name: name,
		Type: "LEAGUE",
		Area: Area{
			ID:   1,
			Name: "England",
			Code: "ENG",
		},
		CurrentSeason: Season{
			ID:              1,
			StartDate:       "2024-08-01",
			EndDate:         "2025-05-31",
			CurrentMatchday: 15,
		},
		Emblem: "https://example.com/emblem.png",
	}
}

// NewTestTeam creates a test team
func NewTestTeam(id int, name string) *Team {
	shortName := name
	tla := name
	if len(name) >= 3 {
		shortName = name[:3]
		tla = name[:3]
	}
	
	return &Team{
		ID:         id,
		Name:       name,
		ShortName:  shortName,
		TLA:        tla,
		Crest:      "https://example.com/crest.png",
		Founded:    1900,
		ClubColors: "Red / White",
		Venue:      "Test Stadium",
	}
}

// NewTestMatch creates a test match
func NewTestMatch(id int, homeTeam, awayTeam *Team) *Match {
	homeScore := 2
	awayScore := 1
	
	return &Match{
		ID: id,
		Competition: Competition{
			ID:   1,
			Code: "PL",
			Name: "Premier League",
		},
		Season: Season{
			ID: 1,
		},
		UTCDate:  time.Now().Add(24 * time.Hour),
		Status:   "SCHEDULED",
		Matchday: 15,
		HomeTeam: *homeTeam,
		AwayTeam: *awayTeam,
		Score: Score{
			Winner:   "HOME_TEAM",
			Duration: "REGULAR",
			FullTime: ScoreData{
				Home: &homeScore,
				Away: &awayScore,
			},
		},
	}
}

// NewTestStanding creates a test standing
func NewTestStanding(code string) *Standing {
	return &Standing{
		Competition: Competition{
			ID:   1,
			Code: code,
			Name: "Premier League",
		},
		Season: Season{
			ID:              1,
			StartDate:       "2024-08-01",
			EndDate:         "2025-05-31",
			CurrentMatchday: 15,
		},
		Standings: []StandingTable{
			{
				Stage: "REGULAR_SEASON",
				Type:  "TOTAL",
				Table: []TeamStanding{
					{
						Position: 1,
						Team: Team{
							ID:   1,
							Name: "Test Team 1",
						},
						PlayedGames: 15,
						Won:         12,
						Draw:        2,
						Lost:        1,
						Points:      38,
					},
				},
			},
		},
	}
}

// CompetitionsEqual compares two competitions for equality
func CompetitionsEqual(a, b *Competition) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.ID == b.ID && a.Code == b.Code && a.Name == b.Name
}

// TeamsEqual compares two teams for equality
func TeamsEqual(a, b *Team) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.ID == b.ID && a.Name == b.Name
}

// MatchesEqual compares two matches for equality
func MatchesEqual(a, b *Match) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.ID == b.ID && a.Status == b.Status
}
