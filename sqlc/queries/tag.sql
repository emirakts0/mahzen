-- name: InsertTag :one
INSERT INTO tags (name, slug)
VALUES ($1, $2)
RETURNING id, name, slug, created_at;

-- name: GetTagByID :one
SELECT id, name, slug, created_at
FROM tags
WHERE id = $1;

-- name: GetTagBySlug :one
SELECT id, name, slug, created_at
FROM tags
WHERE slug = $1;

-- name: ListTags :many
SELECT id, name, slug, created_at
FROM tags
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: CountTags :one
SELECT count(*) FROM tags;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = $1;

-- name: AttachTagToEntry :exec
INSERT INTO entry_tags (entry_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DetachTagFromEntry :exec
DELETE FROM entry_tags
WHERE entry_id = $1 AND tag_id = $2;

-- name: ListTagsByEntry :many
SELECT t.id, t.name, t.slug, t.created_at
FROM tags t
INNER JOIN entry_tags et ON et.tag_id = t.id
WHERE et.entry_id = $1
ORDER BY t.name;
