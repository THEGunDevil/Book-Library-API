-- name: CreateEvent :one
INSERT INTO events (
    object_id,
    object_title,
    type,
    title,
    message,
    metadata,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, NOW()
)
RETURNING *;


-- name: SelectUnreadEventsForUser :many
SELECT e.id
FROM events e
LEFT JOIN user_notification_status uns
  ON e.id = uns.event_id AND uns.user_id = $1
WHERE COALESCE(uns.is_read, false) = false;


-- name: UpsertUserNotificationStatus :exec
INSERT INTO user_notification_status (user_id, event_id, is_read, read_at, created_at)
VALUES ($1, $2, true, NOW(), NOW())
ON CONFLICT (user_id, event_id)
DO UPDATE SET is_read = true, read_at = NOW();

-- name: GetUserNotificationsByUserID :many
SELECT
    e.id AS event_id,
    e.object_id,
    e.object_title,
    e.type,
    e.title AS notification_title,
    e.message,
    e.metadata,
    e.created_at,
    COALESCE(uns.is_read, false) AS is_read,
    uns.read_at
FROM events e
LEFT JOIN user_notification_status uns
    ON e.id = uns.event_id
    AND uns.user_id = $1
ORDER BY e.created_at DESC
LIMIT $2 OFFSET $3;


