-- name: CreateReservation :one
INSERT INTO reservations (user_id, book_id, status)
VALUES ($1, $2, 'pending')
RETURNING id, user_id, book_id, status, created_at, notified_at, fulfilled_at, cancelled_at;

-- name: UpdateReservationStatus :exec
UPDATE reservations
SET status = $1,
    notified_at = CASE WHEN $1 = 'notified' AND notified_at IS NULL THEN now() ELSE notified_at END,
    fulfilled_at = CASE WHEN $1 = 'fulfilled' AND fulfilled_at IS NULL THEN now() ELSE fulfilled_at END,
    cancelled_at = CASE WHEN $1 = 'cancelled' AND cancelled_at IS NULL THEN now() ELSE cancelled_at END
WHERE id = $2;
-- name: GetNextReservationForBook :one
SELECT id, user_id, book_id, status, created_at, notified_at, fulfilled_at, cancelled_at
FROM reservations
WHERE book_id = $1
  AND status = 'pending'
ORDER BY created_at ASC
LIMIT 1;
-- name: GetAllReservations :many
SELECT * FROM reservations
ORDER BY created_at DESC;

-- name: GetReservationsByUser :many
SELECT * FROM reservations
WHERE user_id = $1
ORDER BY created_at DESC;
