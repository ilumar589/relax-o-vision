-- Add embedding column to matches table
ALTER TABLE matches ADD COLUMN IF NOT EXISTS embedding vector(1536);

-- Create index for vector similarity search
CREATE INDEX IF NOT EXISTS idx_matches_embedding ON matches USING ivfflat (embedding vector_cosine_ops);
