-- name: CreateUser :one
INSERT INTO users (first_name, last_name, email, password_hash, phone_number)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUserByID :one
UPDATE users
SET
  first_name = COALESCE($2, first_name),
  last_name = COALESCE($3, last_name),
  phone_number = COALESCE($4, phone_number)
WHERE id = $1
RETURNING *;