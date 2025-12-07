-- Matches table
CREATE TABLE matches (
    id INTEGER PRIMARY KEY,
    competition_id INTEGER REFERENCES competitions(id),
    season_id INTEGER,
    matchday INTEGER,
    status VARCHAR(50),
    utc_date TIMESTAMP,
    home_team JSONB,
    away_team JSONB,
    score JSONB,
    odds JSONB,
    referees JSONB,
    embedding vector(1536),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for faster lookups
CREATE INDEX idx_matches_competition_id ON matches(competition_id);
CREATE INDEX idx_matches_status ON matches(status);
CREATE INDEX idx_matches_utc_date ON matches(utc_date);
