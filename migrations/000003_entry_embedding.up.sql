-- Add embedding column to entries table for storing OpenAI embeddings.
-- Embeddings are stored as JSON array of floats (e.g., [0.1, 0.2, 0.3, ...]).
ALTER TABLE entries ADD COLUMN embedding TEXT;

-- Create an index for faster queries when checking for missing embeddings.
CREATE INDEX idx_entries_embedding_null ON entries (id) WHERE embedding IS NULL;
