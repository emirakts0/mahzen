DROP INDEX IF EXISTS idx_entries_user_path_pattern;
ALTER TABLE entries DROP CONSTRAINT IF EXISTS entries_user_path_unique;
ALTER TABLE entries DROP COLUMN IF EXISTS path;
