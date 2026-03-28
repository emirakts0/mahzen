-- name: CreateAccessToken :one
INSERT INTO access_tokens (user_id, name, token_hash, prefix, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, name, token_hash, prefix, status, expires_at, created_at;

-- name: GetAccessTokenByHash :one
SELECT id, user_id, name, token_hash, prefix, status, expires_at, created_at
FROM access_tokens
WHERE token_hash = $1;

-- name: ListAccessTokensByUserID :many
SELECT id, user_id, name, token_hash, prefix, status, expires_at, created_at
FROM access_tokens
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateAccessTokenStatus :exec
UPDATE access_tokens
SET status = $2
WHERE id = $1;

-- name: MarkAccessTokensExpiredBatch :exec
UPDATE access_tokens
SET status = 'expired'
WHERE id = ANY($1::uuid[]);

-- name: LoadAllActiveAccessTokens :many
SELECT id, user_id, name, token_hash, prefix, status, expires_at, created_at
FROM access_tokens
WHERE status = 'active';
