-- name: CreateReservation :one
INSERT INTO reservations (user_id, book_id, status)
VALUES ($1, $2, 'pending')
RETURNING id, user_id, book_id, status, created_at, notified_at, fulfilled_at, cancelled_at;

-- name: UpdateReservationStatus :one
UPDATE reservations
SET status = $2,
    notified_at = CASE WHEN $2 = 'notified' AND notified_at IS NULL THEN now() ELSE notified_at END,
    fulfilled_at = CASE WHEN $2 = 'fulfilled' AND fulfilled_at IS NULL THEN now() ELSE fulfilled_at END,
    cancelled_at = CASE WHEN $2 = 'cancelled' AND cancelled_at IS NULL THEN now() ELSE cancelled_at END
WHERE id = $1
RETURNING id, user_id, book_id, status, created_at, notified_at, fulfilled_at, cancelled_at;

-- name: GetNextReservationForBook :one
SELECT id, user_id, book_id, status, created_at, notified_at, fulfilled_at, cancelled_at
FROM reservations
WHERE book_id = $1 AND status = 'pending'
ORDER BY created_at ASC
LIMIT 1;

-- name: GetUserReservations :many
SELECT 
    r.id,
    r.user_id,
    r.book_id,
    r.status,
    r.created_at,
    r.notified_at,
    r.fulfilled_at,
    r.cancelled_at,    
    CONCAT(u.first_name, ' ', u.last_name) as user_name,
    b.title as book_title,
    b.author as book_author,
    b.image_url as book_image
FROM reservations r
JOIN books b ON r.book_id = b.id
WHERE r.user_id = $1
ORDER BY r.created_at DESC;

-- name: GetAllReservations :many
SELECT 
    r.id,
    r.user_id,
    r.book_id,
    r.status,
    r.created_at,
    r.notified_at,
    r.fulfilled_at,
    r.cancelled_at,
    CONCAT(u.first_name, ' ', u.last_name) as user_name,
    u.email as user_email,
    b.title as book_title,
    b.author as book_author,
    b.image_url as book_image
FROM reservations r
JOIN users u ON r.user_id = u.id
JOIN books b ON r.book_id = b.id
ORDER BY r.created_at DESC;

-- name: CheckExistingReservation :one
SELECT COUNT(*) as count
FROM reservations
WHERE user_id = $1 AND book_id = $2 AND status IN ('pending', 'notified');