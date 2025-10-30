-- name: ListBooksPaginated :many
SELECT *
FROM books
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;


-- name: GetBookByID :one
SELECT * FROM books
WHERE id = $1;

-- name: FilterBooksByGenre :many
SELECT * FROM books
WHERE genre = $1;
-- name: CountBooks :one
SELECT COUNT(*) FROM books;

-- name: CreateBook :one
INSERT INTO books (title, author, published_year, isbn, total_copies, available_copies, image_url, genre, description)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateBookByID :one
-- name: UpdateBookByID :one
UPDATE books
SET
    title = COALESCE(sqlc.narg('title'), title),
    author = COALESCE(sqlc.narg('author'), author),
    published_year = COALESCE(sqlc.narg('published_year'), published_year),
    isbn = COALESCE(sqlc.narg('isbn'), isbn),
    available_copies = COALESCE(sqlc.narg('available_copies'), available_copies),
    total_copies = COALESCE(sqlc.narg('total_copies'), total_copies),
    genre = COALESCE(sqlc.narg('genre'), genre),
    description = COALESCE(sqlc.narg('description'), description),
    image_url = COALESCE(sqlc.narg('image_url'), image_url),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
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
-- name: SearchBooks :many
SELECT
    id,
    title,
    author,
    genre,
    published_year,
    isbn,
    available_copies,
    total_copies,
    description,
    image_url,
    created_at,
    updated_at
FROM books
WHERE
    ($1::text IS NULL OR genre ILIKE '%' || $1 || '%')
    AND ($2::text IS NULL OR title ILIKE '%' || $2 || '%' OR author ILIKE '%' || $2 || '%')
ORDER BY title;

-- name: ListGenres :many
SELECT DISTINCT genre
FROM books
WHERE genre IS NOT NULL AND genre <> ''
ORDER BY genre;

