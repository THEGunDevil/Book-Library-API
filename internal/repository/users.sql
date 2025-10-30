-- name: CreateUser :one
INSERT INTO users (first_name, last_name, email, password_hash, phone_number, token_version)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetAllUsers :many

SELECT * FROM users;


-- name: UpdateUserByID :one
UPDATE users
SET
  first_name   = COALESCE(sqlc.narg('first_name'), first_name),
  last_name    = COALESCE(sqlc.narg('last_name'), last_name),
  phone_number = COALESCE(sqlc.narg('phone_number'), phone_number),
  bio          = COALESCE(sqlc.narg('bio'), bio)
WHERE id = sqlc.arg('id')
RETURNING *;


-- name: UpdateUserBan :one
UPDATE users
SET is_banned = $1,
    ban_reason = $2,
    ban_until = $3,
    is_permanent_ban = $4
WHERE id = $5
RETURNING *;
