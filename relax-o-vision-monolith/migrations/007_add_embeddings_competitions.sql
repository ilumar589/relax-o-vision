-- Add embedding column to competitions table
ALTER TABLE competitions ADD COLUMN IF NOT EXISTS embedding vector(1536);

-- Create index for vector similarity search
CREATE INDEX IF NOT EXISTS idx_competitions_embedding ON competitions USING ivfflat (embedding vector_cosine_ops);
