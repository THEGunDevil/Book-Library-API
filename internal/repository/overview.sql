-- name: GetStats :one
SELECT
    (SELECT COUNT(*) FROM books) AS total_books,
    (SELECT COUNT(*) FROM users WHERE role='member') AS active_users,
    (SELECT COUNT(*) FROM subscriptions WHERE status='active') AS total_subscriptions,
    (SELECT COALESCE(SUM(p.amount),0)
     FROM payments p
     WHERE p.status='paid'
       AND EXTRACT(MONTH FROM p.created_at) = CAST($1 AS INT)
       AND EXTRACT(YEAR FROM p.created_at) = CAST($2 AS INT)
    ) AS revenue_month;



-- name: GetBooksPerMonth :many
SELECT TO_CHAR(created_at, 'Mon') AS month,
       COUNT(*) AS books
FROM books
WHERE created_at >= NOW() - INTERVAL '6 months'
GROUP BY month
ORDER BY MIN(created_at);

-- name: GetCategoryData :many
SELECT genre AS name,
       COUNT(*) AS value
FROM books
GROUP BY genre;

-- name: GetTopBorrowedBooks :many
SELECT b.title AS name,
       COUNT(*) AS count
FROM borrows br
JOIN books b ON br.book_id = b.id
GROUP BY b.title
ORDER BY count DESC
LIMIT 5;

-- name: GetSubscriptionPlans :many
SELECT sp.name AS plan,
       COUNT(s.id) AS count
FROM subscriptions s
JOIN subscription_plans sp ON s.plan_id = sp.id
WHERE s.status='active'
GROUP BY sp.name;

-- name: GetSubscriptionHistory :many
SELECT TO_CHAR(start_date, 'Mon') AS month,
       COUNT(*) FILTER (WHERE status='active') AS active,
       COUNT(*) FILTER (WHERE status='cancelled') AS cancelled
FROM subscriptions
WHERE start_date >= NOW() - INTERVAL '6 months'
GROUP BY month
ORDER BY MIN(start_date);
