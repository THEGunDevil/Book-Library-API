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

-- name: GetAllUser :many
SELECT * FROM users;
-- name: UpdateUserByID :one
UPDATE users
SET
  first_name            = COALESCE(sqlc.narg(first_name), first_name),
  last_name             = COALESCE(sqlc.narg(last_name), last_name),
  phone_number          = COALESCE(sqlc.narg(phone_number), phone_number),
  bio                   = COALESCE(sqlc.narg(bio), bio),
  profile_img           = COALESCE(sqlc.narg(profile_img), profile_img),
  profile_img_public_id = COALESCE(sqlc.narg(profile_img_public_id), profile_img_public_id),
  updated_at            = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;




-- name: UpdateUserBanByUserID :one
UPDATE users
SET is_banned = COALESCE(sqlc.narg('is_banned'), is_banned),
    ban_reason = COALESCE(sqlc.narg('ban_reason'), ban_reason),
    ban_until = sqlc.narg('ban_until'),
    is_permanent_ban = COALESCE(sqlc.narg('is_permanent_ban'), is_permanent_ban)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: IncrementTokenVersion :exec
UPDATE users
SET token_version = token_version + 1
WHERE id = $1;
-- name: ListUsersPaginated :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: SearchUsersByEmailWithPagination :many
SELECT
    id,
    first_name,
    last_name,
    email,
    role,
    created_at,
    updated_at
FROM users
WHERE
    (CASE 
        WHEN $1 = '' THEN TRUE
        ELSE email ILIKE '%' || $1 || '%'
    END)
ORDER BY email
LIMIT $2
OFFSET $3;
-- name: CountUsersByEmail :one
SELECT COUNT(*)
FROM users
WHERE
    (CASE 
        WHEN $1 = '' THEN TRUE
        ELSE email ILIKE '%' || $1 || '%'
    END);
