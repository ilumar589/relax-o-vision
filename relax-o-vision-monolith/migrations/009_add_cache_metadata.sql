-- Add cached_at column to competitions table
ALTER TABLE competitions ADD COLUMN IF NOT EXISTS cached_at TIMESTAMP WITH TIME ZONE;

-- Add cached_at column to teams table
ALTER TABLE teams ADD COLUMN IF NOT EXISTS cached_at TIMESTAMP WITH TIME ZONE;

-- Add cached_at column to matches table
ALTER TABLE matches ADD COLUMN IF NOT EXISTS cached_at TIMESTAMP WITH TIME ZONE;

-- Create cache_metadata table for tracking cache state per championship
CREATE TABLE IF NOT EXISTS cache_metadata (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,  -- 'competition', 'team', 'match', 'standings'
    entity_key VARCHAR(100) NOT NULL,   -- competition code or entity ID
    cached_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    data_hash VARCHAR(64),              -- optional: to detect if data changed
    UNIQUE(entity_type, entity_key)
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_cache_metadata_type_key ON cache_metadata(entity_type, entity_key);
CREATE INDEX IF NOT EXISTS idx_cache_metadata_expires_at ON cache_metadata(expires_at);
