-- name: CreateSubscriptionPlan :one
INSERT INTO subscription_plans (
    name, price, duration_days, description, features, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, NOW(), NOW()
) RETURNING *;

-- name: GetSubscriptionPlanByID :one
SELECT * FROM subscription_plans
WHERE id = $1;

-- name: ListSubscriptionPlans :many
SELECT * FROM subscription_plans
ORDER BY created_at DESC;

-- name: UpdateSubscriptionPlan :one
UPDATE subscription_plans
SET name = $2,
    price = $3,
    duration_days = $4,
    description = $5,
    features = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSubscriptionPlan :exec
DELETE FROM subscription_plans
WHERE id = $1;
-- name: CountSubsPerUser :exec
SELECT COUNT(*) 
FROM subscriptions 
WHERE user_id = $1 AND plan_id = $2 AND status = 'active';


-- name: CreateSubscription :one
INSERT INTO subscriptions (
user_id, plan_id, start_date, end_date, status, auto_renew, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, NOW(), NOW()
) RETURNING *;

-- name: GetSubscriptionByID :one
SELECT * FROM subscriptions
WHERE id = $1;
-- name: GetSubscriptionByUserID :one
SELECT * FROM subscriptions
WHERE user_id = $1
  AND status = 'active'
  AND end_date >= NOW()
ORDER BY end_date DESC
LIMIT 1;
-- name: ListSubscriptions :many
SELECT * FROM subscriptions
ORDER BY created_at DESC;
-- name: ListSubscriptionsByUser :many
SELECT * FROM subscriptions
WHERE user_id = $1
ORDER BY start_date DESC;

-- name: UpdateSubscription :one
UPDATE subscriptions
SET plan_id = $2,
    start_date = $3,
    end_date = $4,
    status = $5,
    auto_renew = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSubscription :exec
DELETE FROM subscriptions
WHERE id = $1;

