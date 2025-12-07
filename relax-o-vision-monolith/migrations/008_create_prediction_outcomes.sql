-- Prediction outcomes table for tracking accuracy
CREATE TABLE IF NOT EXISTS prediction_outcomes (
    id UUID PRIMARY KEY,
    prediction_id UUID REFERENCES predictions(id),
    match_id INTEGER REFERENCES matches(id),
    predicted_winner VARCHAR(10), -- 'home', 'away', 'draw'
    actual_winner VARCHAR(10),
    was_correct BOOLEAN,
    confidence_score DECIMAL(5,4),
    home_win_prob DECIMAL(5,4),
    draw_prob DECIMAL(5,4),
    away_win_prob DECIMAL(5,4),
    actual_home_score INTEGER,
    actual_away_score INTEGER,
    competition_id INTEGER,
    competition_name VARCHAR(255),
    provider VARCHAR(50),
    agent_type VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_outcomes_prediction_id ON prediction_outcomes(prediction_id);
CREATE INDEX IF NOT EXISTS idx_outcomes_match_id ON prediction_outcomes(match_id);
CREATE INDEX IF NOT EXISTS idx_outcomes_competition_id ON prediction_outcomes(competition_id);
CREATE INDEX IF NOT EXISTS idx_outcomes_was_correct ON prediction_outcomes(was_correct);
CREATE INDEX IF NOT EXISTS idx_outcomes_confidence ON prediction_outcomes(confidence_score);
CREATE INDEX IF NOT EXISTS idx_outcomes_provider ON prediction_outcomes(provider);
CREATE INDEX IF NOT EXISTS idx_outcomes_agent_type ON prediction_outcomes(agent_type);
