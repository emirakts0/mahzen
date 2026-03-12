-- Re-add s3_key column for rollback.
ALTER TABLE entries ADD COLUMN s3_key text DEFAULT '';
