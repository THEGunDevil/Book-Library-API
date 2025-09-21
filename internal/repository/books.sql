-- name: ListBooks :many
SELECT * FROM books
ORDER BY created_at DESC;

-- name: GetBookByID :one
SELECT * FROM books
WHERE id = $1;

-- name: CreateBook :one
INSERT INTO books (title, author, published_year, isbn, total_copies, available_copies,image_url)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateBookByID :one
UPDATE books
SET
  title = COALESCE($2, title),
  author = COALESCE($3, author),
  published_year = COALESCE($4, published_year),
  isbn = COALESCE($5, isbn),
  total_copies = COALESCE($6, total_copies),
  available_copies = COALESCE($7, available_copies),
  updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DecrementAvailableCopiesByID :one
UPDATE books
SET available_copies = available_copies - 1
WHERE id = $1 AND available_copies > 0
RETURNING available_copies;

-- name: IncrementAvailableCopiesByID :one
UPDATE books
SET available_copies = available_copies + 1
WHERE id = $1
RETURNING available_copies;


-- name: DeleteBookByID :one
DELETE FROM books
WHERE id = $1
RETURNING *;
