-- name: CreateReview :one
INSERT INTO reviews (user_id, book_id, rating, comment)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, book_id, rating, comment, created_at, updated_at;

-- name: GetReviewByID :one
SELECT id, user_id, book_id, rating, comment, created_at, updated_at
FROM reviews
WHERE id = $1;

-- name: GetReviewsByBook :many
SELECT id, user_id, book_id, rating, comment, created_at, updated_at
FROM reviews
WHERE book_id = $1
ORDER BY created_at DESC;

-- name: UpdateReviewByID :one
UPDATE reviews
SET
  rating = COALESCE($2, rating),
  comment = COALESCE($3, comment)
WHERE id = $1
RETURNING *;


-- name: DeleteReview :exec
DELETE FROM reviews
WHERE id = $1;
