-- name: CreateUser :one
INSERT INTO users (username, email, display_name, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING id, username, email, display_name, password_hash, created_at;

-- name: GetUserByID :one
SELECT id, username, email, display_name, password_hash, created_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, username, email, display_name, password_hash, created_at
FROM users
WHERE email = $1;

-- name: GetUserByIdentifier :one
SELECT id, username, email, display_name, password_hash, created_at
FROM users
WHERE email = $1 OR username = $1;
