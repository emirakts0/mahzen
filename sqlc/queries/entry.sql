-- name: InsertEntry :one
INSERT INTO entries (user_id, title, content, summary, s3_key, path, visibility, file_type, file_size, embedding)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, user_id, title, content, summary, s3_key, path, visibility, file_type, file_size, embedding, created_at, updated_at;

-- name: GetEntryByID :one
SELECT id, user_id, title, content, summary, s3_key, path, visibility, file_type, file_size, embedding, created_at, updated_at
FROM entries
WHERE id = $1;

-- name: UpdateEntry :one
UPDATE entries
SET title = $2, content = $3, summary = $4, s3_key = $5, path = $6, visibility = $7, file_type = $8, file_size = $9, embedding = $10, updated_at = now()
WHERE id = $1
RETURNING id, user_id, title, content, summary, s3_key, path, visibility, file_type, file_size, embedding, created_at, updated_at;

-- name: UpdateEntryEmbedding :exec
UPDATE entries SET embedding = $2, updated_at = now() WHERE id = $1;

-- name: DeleteEntry :exec
DELETE FROM entries WHERE id = $1;

-- name: ListEntriesByUser :many
SELECT id, user_id, title, content, summary, s3_key, path, visibility, file_type, file_size, embedding, created_at, updated_at
FROM entries
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountEntriesByUser :one
SELECT count(*) FROM entries WHERE user_id = $1;

-- name: ListAccessibleEntries :many
SELECT id, user_id, title, content, summary, s3_key, path, visibility, file_type, file_size, embedding, created_at, updated_at
FROM entries
WHERE (visibility = 'public' OR user_id = $1)
ORDER BY path ASC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAccessibleEntries :one
SELECT count(*) FROM entries
WHERE visibility = 'public' OR user_id = $1;

-- name: ListAccessibleEntriesByPath :many
SELECT id, user_id, title, content, summary, s3_key, path, visibility, file_type, file_size, embedding, created_at, updated_at
FROM entries
WHERE (visibility = 'public' OR user_id = $1)
  AND (path = $2 OR path LIKE $2 || '/%')
ORDER BY path ASC, created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountAccessibleEntriesByPath :one
SELECT count(*) FROM entries
WHERE (visibility = 'public' OR user_id = $1)
  AND (path = $2 OR path LIKE $2 || '/%');

-- name: ListDistinctPaths :many
SELECT DISTINCT path FROM entries
WHERE (visibility = 'public' OR user_id = sqlc.narg('user_id')::uuid)
ORDER BY path ASC;

-- name: ListEntriesInPath :many
SELECT id, user_id, title, content, summary, s3_key, path, visibility, file_type, file_size, embedding, created_at, updated_at
FROM entries
WHERE 
  CASE 
    WHEN sqlc.arg('own')::boolean THEN user_id = sqlc.narg('user_id')::uuid
    ELSE (visibility = 'public' OR user_id = sqlc.narg('user_id')::uuid)
  END
  AND path = sqlc.arg('path')::text
ORDER BY created_at DESC
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: CountEntriesInPath :one
SELECT count(*) FROM entries
WHERE 
  CASE 
    WHEN sqlc.arg('own')::boolean THEN user_id = sqlc.narg('user_id')::uuid
    ELSE (visibility = 'public' OR user_id = sqlc.narg('user_id')::uuid)
  END
  AND path = sqlc.arg('path')::text;

-- name: ListPathsUnderPrefix :many
SELECT DISTINCT path FROM entries
WHERE 
  CASE 
    WHEN sqlc.arg('own')::boolean THEN user_id = sqlc.narg('user_id')::uuid
    ELSE (visibility = 'public' OR user_id = sqlc.narg('user_id')::uuid)
  END
  AND path LIKE sqlc.arg('prefix')::text || '/%'
ORDER BY path ASC;

-- name: ListAllPaths :many
SELECT DISTINCT path FROM entries
WHERE 
  CASE 
    WHEN sqlc.arg('own')::boolean THEN user_id = sqlc.narg('user_id')::uuid
    ELSE (visibility = 'public' OR user_id = sqlc.narg('user_id')::uuid)
  END
ORDER BY path ASC;

-- name: ListAllEntries :many
SELECT id, user_id, title, content, summary, s3_key, path, visibility, file_type, file_size, embedding, created_at, updated_at
FROM entries
ORDER BY created_at ASC;
