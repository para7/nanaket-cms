-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: CreateAccessToken :one
INSERT INTO access_tokens (
    user_id, token, expires_at
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetUserByToken :one
SELECT u.* FROM users u
INNER JOIN access_tokens t ON u.id = t.user_id
WHERE t.token = $1
  AND (t.expires_at IS NULL OR t.expires_at > CURRENT_TIMESTAMP)
LIMIT 1;

-- name: DeleteAccessToken :exec
DELETE FROM access_tokens
WHERE token = $1;
