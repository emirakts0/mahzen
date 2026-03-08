-- Remove embedding column from entries table.
DROP INDEX IF EXISTS idx_entries_embedding_null;
ALTER TABLE entries DROP COLUMN IF EXISTS embedding;
