-- name: CreateReview :one
INSERT INTO reviews (user_id, book_id, rating, comment)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, book_id, rating, comment, created_at, updated_at;

-- name: GetReviewsByUserID :many
SELECT r.id, r.user_id, b.title, r.book_id, r.rating, r.comment, r.created_at, r.updated_at
FROM reviews r
JOIN books b ON b.id = r.book_id
WHERE r.user_id = $1
ORDER BY r.created_at DESC;

-- name: GetReviewsByBookID :many
SELECT r.id, r.user_id, u.first_name,u.last_name, r.book_id, r.rating, r.comment, r.created_at, r.updated_at
FROM reviews r
JOIN users u ON u.id = r.user_id
WHERE r.book_id = $1
ORDER BY r.created_at DESC;

-- name: GetReviewsByReviewID :many
SELECT r.id, r.user_id, u.first_name, u.last_name, r.book_id, r.rating, r.comment, r.created_at, r.updated_at
FROM reviews r
JOIN users u ON u.id = r.user_id
WHERE r.id = $1
ORDER BY r.created_at DESC;



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
