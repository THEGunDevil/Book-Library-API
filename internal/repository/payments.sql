-- name: CreatePayment :one
INSERT INTO payments (
    id, user_id, subscription_id, amount, currency, transaction_id, payment_gateway, status, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, NOW()
) RETURNING *;

-- name: GetPaymentByID :one
SELECT * FROM payments
WHERE id = $1;

-- name: ListPaymentsByUser :many
SELECT * FROM payments
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeletePayment :exec
DELETE FROM payments
WHERE id = $1;


-- ===============================
-- Refunds
-- ===============================

-- name: CreateRefund :one
INSERT INTO refunds (
    id, payment_id, amount, reason, status, requested_at, processed_at
) VALUES (
    $1, $2, $3, $4, $5, NOW(), $6
) RETURNING *;

-- name: GetRefundByID :one
SELECT * FROM refunds
WHERE id = $1;

-- name: ListRefundsByPayment :many
SELECT * FROM refunds
WHERE payment_id = $1
ORDER BY requested_at DESC;

-- name: ListRefundsByStatus :many
SELECT * FROM refunds
WHERE status = $1
ORDER BY requested_at DESC;

-- name: UpdateRefundStatus :one
UPDATE refunds
SET status = $2,
    processed_at = $3
WHERE id = $1
RETURNING *;

-- name: DeleteRefund :exec
DELETE FROM refunds
WHERE id = $1;
