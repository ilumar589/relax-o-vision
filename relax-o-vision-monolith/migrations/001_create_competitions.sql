-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "pgvector";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Competitions table
CREATE TABLE competitions (
    id INTEGER PRIMARY KEY,
    code VARCHAR(10),
    name VARCHAR(255),
    type VARCHAR(50),
    emblem TEXT,
    area JSONB,
    current_season JSONB,
    seasons JSONB,
    embedding vector(1536),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create index on code for faster lookups
CREATE INDEX idx_competitions_code ON competitions(code);
