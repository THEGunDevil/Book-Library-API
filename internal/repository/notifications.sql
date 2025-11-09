-- name: CreateNotification :exec
INSERT INTO notifications (
    user_id,
    user_name,
    object_id,
    object_title,
    type,
    notification_title,
    message,
    metadata,
    is_read,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, false, NOW()
)
RETURNING *;

-- name: GetUserNotificationsByUserID :many
SELECT *
FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: MarkNotificationAsRead :exec
UPDATE notifications
SET is_read = true
WHERE id = $1;
