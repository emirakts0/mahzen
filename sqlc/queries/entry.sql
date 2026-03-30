-- name: InsertEntry :one
INSERT INTO entries (user_id, title, content, summary, path, visibility, file_type, file_size, embedding)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, user_id, title, content, summary, path, visibility, file_type, file_size, embedding, created_at, updated_at;

-- name: GetEntryByID :one
SELECT id, user_id, title, content, summary, path, visibility, file_type, file_size, embedding, created_at, updated_at
FROM entries
WHERE id = $1;

-- name: UpdateEntry :one
UPDATE entries
SET title = $2, content = $3, summary = $4, path = $5, visibility = $6, file_type = $7, file_size = $8, embedding = $9, updated_at = now()
WHERE id = $1
RETURNING id, user_id, title, content, summary, path, visibility, file_type, file_size, embedding, created_at, updated_at;

-- name: UpdateEntryEmbedding :exec
UPDATE entries SET embedding = $2, updated_at = now() WHERE id = $1;

-- name: DeleteEntry :exec
DELETE FROM entries WHERE id = $1;

-- name: ListEntriesByUser :many
SELECT id, user_id, title, content, summary, path, visibility, file_type, file_size, embedding, created_at, updated_at
FROM entries
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountEntriesByUser :one
SELECT count(*) FROM entries WHERE user_id = $1;

-- name: ListAccessibleEntries :many
SELECT id, user_id, title, content, summary, path, visibility, file_type, file_size, embedding, created_at, updated_at
FROM entries
WHERE (visibility = 'public' OR user_id = $1)
ORDER BY path ASC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAccessibleEntries :one
SELECT count(*) FROM entries
WHERE visibility = 'public' OR user_id = $1;

-- name: ListAccessibleEntriesByPath :many
SELECT id, user_id, title, content, summary, path, visibility, file_type, file_size, embedding, created_at, updated_at
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

-- name: ListEntriesInPathWithCount :many
SELECT e.id, e.user_id, e.title, e.content, e.summary, e.path, e.visibility, e.file_type, e.file_size, e.embedding, e.created_at, e.updated_at,
  COUNT(*) OVER() AS total_count
FROM entries e
WHERE
  CASE
    WHEN sqlc.arg('own')::boolean THEN e.user_id = sqlc.narg('user_id')::uuid
    ELSE (e.visibility = 'public' OR e.user_id = sqlc.narg('user_id')::uuid)
  END
  AND e.path = sqlc.arg('path')::text
  AND (
    sqlc.arg('filter_visibility')::text IS NULL OR
    sqlc.arg('filter_visibility')::text = '' OR
    e.visibility = sqlc.arg('filter_visibility')::text
  )
  AND (
    sqlc.arg('from_date')::timestamptz IS NULL OR
    e.created_at >= sqlc.arg('from_date')::timestamptz
  )
  AND (
    sqlc.arg('to_date')::timestamptz IS NULL OR
    e.created_at <= sqlc.arg('to_date')::timestamptz
  )
  AND (
    sqlc.arg('filter_tags')::text[] IS NULL OR
    sqlc.arg('filter_tags')::text[] = '{}' OR
    EXISTS (
      SELECT 1 FROM entry_tags et
      JOIN tags t ON t.id = et.tag_id
      WHERE et.entry_id = e.id AND t.name = ANY(sqlc.arg('filter_tags')::text[])
    )
  )
ORDER BY e.created_at DESC
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: CountPathsUnderPrefix :many
SELECT e.path, COUNT(*) AS count
FROM entries e
WHERE
  CASE
    WHEN sqlc.arg('own')::boolean THEN e.user_id = sqlc.narg('user_id')::uuid
    ELSE (e.visibility = 'public' OR e.user_id = sqlc.narg('user_id')::uuid)
  END
  AND e.path LIKE sqlc.arg('prefix')::text || '/%'
  AND (
    sqlc.arg('filter_visibility')::text IS NULL OR
    sqlc.arg('filter_visibility')::text = '' OR
    e.visibility = sqlc.arg('filter_visibility')::text
  )
  AND (
    sqlc.arg('from_date')::timestamptz IS NULL OR
    e.created_at >= sqlc.arg('from_date')::timestamptz
  )
  AND (
    sqlc.arg('to_date')::timestamptz IS NULL OR
    e.created_at <= sqlc.arg('to_date')::timestamptz
  )
  AND (
    sqlc.arg('filter_tags')::text[] IS NULL OR
    sqlc.arg('filter_tags')::text[] = '{}' OR
    EXISTS (
      SELECT 1 FROM entry_tags et
      JOIN tags t ON t.id = et.tag_id
      WHERE et.entry_id = e.id AND t.name = ANY(sqlc.arg('filter_tags')::text[])
    )
  )
GROUP BY e.path
ORDER BY e.path ASC;

-- name: CountAllPaths :many
SELECT e.path, COUNT(*) AS count
FROM entries e
WHERE
  CASE
    WHEN sqlc.arg('own')::boolean THEN e.user_id = sqlc.narg('user_id')::uuid
    ELSE (e.visibility = 'public' OR e.user_id = sqlc.narg('user_id')::uuid)
  END
  AND (
    sqlc.arg('filter_visibility')::text IS NULL OR
    sqlc.arg('filter_visibility')::text = '' OR
    e.visibility = sqlc.arg('filter_visibility')::text
  )
  AND (
    sqlc.arg('from_date')::timestamptz IS NULL OR
    e.created_at >= sqlc.arg('from_date')::timestamptz
  )
  AND (
    sqlc.arg('to_date')::timestamptz IS NULL OR
    e.created_at <= sqlc.arg('to_date')::timestamptz
  )
  AND (
    sqlc.arg('filter_tags')::text[] IS NULL OR
    sqlc.arg('filter_tags')::text[] = '{}' OR
    EXISTS (
      SELECT 1 FROM entry_tags et
      JOIN tags t ON t.id = et.tag_id
      WHERE et.entry_id = e.id AND t.name = ANY(sqlc.arg('filter_tags')::text[])
    )
  )
GROUP BY e.path
ORDER BY e.path ASC;

-- name: ListAllEntries :many
SELECT id, user_id, title, content, summary, path, visibility, file_type, file_size, embedding, created_at, updated_at
FROM entries
ORDER BY created_at ASC;

-- name: CountAllEntries :one
SELECT count(*) FROM entries;
