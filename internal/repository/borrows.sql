-- name: ListBorrow :many
SELECT * FROM borrows
ORDER BY due_date DESC;

-- name: ListBorrowByUserID :many
SELECT * FROM borrows WHERE user_id = $1
ORDER BY due_date DESC;

-- name: FilterBorrowByUserAndBookID :one
SELECT * FROM borrows WHERE user_id = $1 AND book_id = $2 AND returned_at IS NULL;

-- name: CreateBorrow :one
INSERT INTO borrows (user_id,book_id,due_date,returned_at) VALUES ($1,$2,$3,$4)
RETURNING *;

-- name: UpdateBorrowByUserAndBookID :exec
UPDATE borrows
SET returned_at = NOW()
WHERE user_id = $1 AND book_id = $2 AND returned_at IS NULL;