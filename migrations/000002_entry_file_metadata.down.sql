ALTER TABLE entries
    DROP COLUMN IF EXISTS file_type,
    DROP COLUMN IF EXISTS file_size;
