-- name: UpsertUser :one
INSERT INTO users (kratos_id, email, display_name)
VALUES ($1, $2, $3)
ON CONFLICT (kratos_id)
DO UPDATE SET email = EXCLUDED.email, display_name = EXCLUDED.display_name
RETURNING id, kratos_id, email, display_name, created_at;

-- name: GetUserByID :one
SELECT id, kratos_id, email, display_name, created_at
FROM users
WHERE id = $1;

-- name: GetUserByKratosID :one
SELECT id, kratos_id, email, display_name, created_at
FROM users
WHERE kratos_id = $1;
