-- Add materialized path column to entries for hierarchical organization.
-- Each user has their own path namespace; default path is '/' (root).
ALTER TABLE entries ADD COLUMN path TEXT NOT NULL DEFAULT '/';

-- Ensure each user can only have one entry per path.
ALTER TABLE entries ADD CONSTRAINT entries_user_path_unique UNIQUE (user_id, path);

-- B-tree index with text_pattern_ops enables efficient prefix queries
-- such as: WHERE path LIKE '/notes/%'
CREATE INDEX idx_entries_user_path_pattern ON entries (user_id, path text_pattern_ops);
