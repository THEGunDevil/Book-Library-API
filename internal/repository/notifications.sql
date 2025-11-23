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

-- name: AssignNotificationToUser :exec
-- Use this for TARGETED notifications (Direct Messages). It forces is_read = false.
INSERT INTO user_notification_status (
    user_id, event_id, is_read, read_at, created_at
) VALUES (
    $1, $2, false, NULL, NOW()
)
ON CONFLICT (user_id, event_id) DO NOTHING;

-- name: MarkNotificationAsRead :exec
-- Use this when a user clicks a notification.
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
JOIN users u ON u.id = $1
LEFT JOIN user_notification_status uns 
    ON e.id = uns.event_id AND uns.user_id = $1
WHERE 
    e.created_at >= u.created_at
    OR uns.user_id IS NOT NULL
ORDER BY e.created_at DESC
LIMIT $2 OFFSET $3;

-- name: MarkAllNotificationsAsRead :exec
INSERT INTO user_notification_status (user_id, event_id, is_read, read_at, created_at)
SELECT 
    $1, e.id, true, NOW(), NOW()
FROM events e
JOIN users u ON u.id = $1
LEFT JOIN user_notification_status uns ON e.id = uns.event_id AND uns.user_id = $1
WHERE 
    (e.created_at >= u.created_at OR uns.user_id IS NOT NULL)
    AND COALESCE(uns.is_read, false) = false
ON CONFLICT (user_id, event_id) 
DO UPDATE SET is_read = true, read_at = NOW();