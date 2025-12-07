-- Add vector extension if not exists
CREATE EXTENSION IF NOT EXISTS vector;

-- Add embedding column to teams table
ALTER TABLE teams ADD COLUMN IF NOT EXISTS embedding vector(1536);

-- Create index for vector similarity search
CREATE INDEX IF NOT EXISTS idx_teams_embedding ON teams USING ivfflat (embedding vector_cosine_ops);
