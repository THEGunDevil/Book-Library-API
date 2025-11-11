-- name: ListBorrow :many
SELECT * FROM borrows
ORDER BY due_date DESC;

-- name: ListBorrowByUserID :many
SELECT brs.*, b.title
FROM borrows brs
JOIN books b ON b.id = brs.book_id  -- <- fix here
WHERE brs.user_id = $1
ORDER BY brs.due_date DESC;


-- name: ListBorrowByBookID :many
SELECT brs.*, b.title FROM borrows brs
JOIN books b ON b.id = brs.book_id
WHERE book_id = $1
ORDER BY due_date DESC;

-- name: FilterBorrowByUserAndBookID :one
SELECT * FROM borrows WHERE user_id = $1 AND book_id = $2 AND returned_at IS NULL;

-- name: CreateBorrow :one
INSERT INTO borrows (user_id,book_id,due_date,returned_at) VALUES ($1,$2,$3,$4)
RETURNING *;

-- name: UpdateBorrowReturnedAtByID :exec
UPDATE borrows
SET returned_at = NOW()
WHERE id = $1 AND returned_at IS NULL;

-- name: ListBorrowPaginated :many
SELECT b.id, b.user_id, b.book_id, b.borrowed_at, b.due_date, b.returned_at, bk.title AS book_title
FROM borrows b
JOIN books bk ON b.book_id = bk.id
LIMIT $1 OFFSET $2;

-- name: ListBorrowPaginatedByBorrowedAt :many
SELECT b.id, b.user_id, b.book_id, b.borrowed_at, b.due_date, b.returned_at, bk.title AS book_title, CAST(CONCAT(u.first_name, ' ', u.last_name) AS TEXT) AS user_name
FROM borrows b
JOIN books bk ON b.book_id = bk.id
JOIN users u on bk.user_id = u.id
ORDER BY b.borrowed_at ASC
LIMIT $1 OFFSET $2;

-- name: ListBorrowPaginatedByReturnedAt :many
SELECT b.id, b.user_id, b.book_id, b.borrowed_at, b.due_date, b.returned_at, bk.title AS book_title, CAST(CONCAT(u.first_name, ' ', u.last_name) AS TEXT) AS user_name
FROM borrows b
JOIN books bk ON b.book_id = bk.id
JOIN users u on bk.user_id = u.id
WHERE returned_at IS NOT NULL
ORDER BY b.borrowed_at ASC
LIMIT $1 OFFSET $2;

-- name: ListBorrowPaginatedByNotReturnedAt :many
SELECT b.id, b.user_id, b.book_id, b.borrowed_at, b.due_date, b.returned_at, bk.title AS book_title, CAST(CONCAT(u.first_name, ' ', u.last_name) AS TEXT) AS user_name
FROM borrows b
JOIN books bk ON b.book_id = bk.id
JOIN users u on bk.user_id = u.id
WHERE returned_at IS NULL
ORDER BY b.borrowed_at ASC
LIMIT $1 OFFSET $2;

-- name: CountBorrows :one
SELECT COUNT(*) FROM borrows;

-- name: CountBorrowedBooksByUserID :one
SELECT COUNT(*)
FROM borrows
WHERE user_id = $1;

-- name: CountActiveBorrowsByUserID :one
SELECT COUNT(*)
FROM borrows
WHERE user_id = $1
AND returned_at IS NULL;


