-- Remove s3_key column as we now store all content directly in the database.
ALTER TABLE entries DROP COLUMN IF EXISTS s3_key;
