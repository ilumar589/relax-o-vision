package footballdata

// ExtendedMatchStats represents extended match statistics
type ExtendedMatchStats struct {
	Possession    TeamStatPair `json:"possession"`
	Shots         TeamStatPair `json:"shots"`
	ShotsOnTarget TeamStatPair `json:"shotsOnTarget"`
	Corners       TeamStatPair `json:"corners"`
	Fouls         TeamStatPair `json:"fouls"`
	YellowCards   TeamStatPair `json:"yellowCards"`
	RedCards      TeamStatPair `json:"redCards"`
	Offsides      TeamStatPair `json:"offsides"`
	PassAccuracy  TeamStatPair `json:"passAccuracy"`
	ExpectedGoals TeamStatPair `json:"xG"`
}

// TeamStatPair represents a statistic for both teams
type TeamStatPair struct {
	Home float64 `json:"home"`
	Away float64 `json:"away"`
}
