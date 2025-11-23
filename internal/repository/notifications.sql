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
-- 1. Join users to get the signup date
JOIN users u ON u.id = $1
LEFT JOIN user_notification_status uns 
  ON e.id = uns.event_id AND uns.user_id = $1
WHERE 
  COALESCE(uns.is_read, false) = false
  -- 2. Apply the same time filter
  AND (e.created_at >= u.created_at OR uns.user_id IS NOT NULL);


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
-- 1. Join the Users table to get the current user's signup date
JOIN users u ON u.id = $1
LEFT JOIN user_notification_status uns 
    ON e.id = uns.event_id 
    AND uns.user_id = $1
WHERE 
    -- 2. The Logic: Show event ONLY IF it happened after the user signed up
    e.created_at >= u.created_at
    -- 3. Optional: OR if it was specifically sent to them (handles edge cases)
    OR uns.user_id IS NOT NULL
ORDER BY e.created_at DESC
LIMIT $2 OFFSET $3;


