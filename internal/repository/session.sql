-- name: CreateSession :one
INSERT INTO sessions (user_id, refresh_token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions
WHERE refresh_token = $1
LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE refresh_token = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < NOW();
