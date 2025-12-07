-- Predictions table
CREATE TABLE predictions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    match_id INTEGER REFERENCES matches(id),
    home_win_prob DECIMAL(5,4),
    draw_prob DECIMAL(5,4),
    away_win_prob DECIMAL(5,4),
    confidence DECIMAL(5,4),
    reasoning JSONB,
    agent_outputs JSONB,
    workflow_id VARCHAR(255),
    status VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for faster lookups
CREATE INDEX idx_predictions_match_id ON predictions(match_id);
CREATE INDEX idx_predictions_workflow_id ON predictions(workflow_id);
CREATE INDEX idx_predictions_status ON predictions(status);
