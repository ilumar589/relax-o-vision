package footballdata

import "time"

// Area represents a geographical area (country/region)
type Area struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Flag string `json:"flag"`
}

// Team represents a football team with complete details
type Team struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	ShortName   string    `json:"shortName"`
	TLA         string    `json:"tla"`
	Crest       string    `json:"crest"`
	Address     string    `json:"address"`
	Website     string    `json:"website"`
	Founded     int       `json:"founded"`
	ClubColors  string    `json:"clubColors"`
	Venue       string    `json:"venue"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// Season represents a competition season with details
type Season struct {
	ID              int      `json:"id"`
	StartDate       string   `json:"startDate"`
	EndDate         string   `json:"endDate"`
	CurrentMatchday int      `json:"currentMatchday"`
	Winner          *Team    `json:"winner"`
	Stages          []string `json:"stages"`
}

// Competition represents a football competition/league
type Competition struct {
	Area          Area   `json:"area"`
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	Type          string `json:"type"`
	Emblem        string `json:"emblem"`
	CurrentSeason Season `json:"currentSeason"`
	Seasons       []Season `json:"seasons"`
}

// Score represents match score information
type Score struct {
	Winner   string    `json:"winner"`
	Duration string    `json:"duration"`
	FullTime ScoreData `json:"fullTime"`
	HalfTime ScoreData `json:"halfTime"`
}

// ScoreData represents score details
type ScoreData struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}

// Referee represents a match referee
type Referee struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Nationality string `json:"nationality"`
}

// Match represents a football match
type Match struct {
	ID            int       `json:"id"`
	CompetitionID int       `json:"-"`
	Competition   Competition `json:"competition"`
	Season        Season    `json:"season"`
	UTCDate       time.Time `json:"utcDate"`
	Status        string    `json:"status"`
	Matchday      int       `json:"matchday"`
	Stage         string    `json:"stage"`
	Group         *string   `json:"group"`
	LastUpdated   time.Time `json:"lastUpdated"`
	HomeTeam      Team      `json:"homeTeam"`
	AwayTeam      Team      `json:"awayTeam"`
	Score         Score     `json:"score"`
	Odds          *Odds     `json:"odds"`
	Referees      []Referee `json:"referees"`
}

// Odds represents betting odds (if available)
type Odds struct {
	HomeWin float64 `json:"homeWin"`
	Draw    float64 `json:"draw"`
	AwayWin float64 `json:"awayWin"`
}

// StandingTable represents a league table
type TeamStanding struct {
	Position       int     `json:"position"`
	Team           Team    `json:"team"`
	PlayedGames    int     `json:"playedGames"`
	Form           *string `json:"form"`
	Won            int     `json:"won"`
	Draw           int     `json:"draw"`
	Lost           int     `json:"lost"`
	Points         int     `json:"points"`
	GoalsFor       int     `json:"goalsFor"`
	GoalsAgainst   int     `json:"goalsAgainst"`
	GoalDifference int     `json:"goalDifference"`
}

// StandingTable represents a league table
type StandingTable struct {
	Stage string          `json:"stage"`
	Type  string          `json:"type"`
	Group *string         `json:"group"`
	Table []TeamStanding  `json:"table"`
}

// TablePosition is an alias for TeamStanding for backwards compatibility
type TablePosition = TeamStanding

// Standing represents competition standings
type Standing struct {
	Competition Competition     `json:"competition"`
	Season      Season          `json:"season"`
	Standings   []StandingTable `json:"standings"`
}

// CompetitionsResponse wraps API response for competitions
type CompetitionsResponse struct {
	Count        int           `json:"count"`
	Competitions []Competition `json:"competitions"`
}

// MatchesResponse wraps API response for matches
type MatchesResponse struct {
	Count   int     `json:"count"`
	Matches []Match `json:"matches"`
}
