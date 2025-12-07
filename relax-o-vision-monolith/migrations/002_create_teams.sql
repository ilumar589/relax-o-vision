-- Teams table
CREATE TABLE teams (
    id INTEGER PRIMARY KEY,
    name VARCHAR(255),
    short_name VARCHAR(100),
    tla VARCHAR(10),
    crest TEXT,
    address TEXT,
    website TEXT,
    founded INTEGER,
    club_colors VARCHAR(100),
    venue VARCHAR(255),
    area JSONB,
    embedding vector(1536),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for faster lookups
CREATE INDEX idx_teams_name ON teams(name);
CREATE INDEX idx_teams_tla ON teams(tla);
