-- +goose Up
-- Adds user banning columns to the users table

ALTER TABLE users
ADD COLUMN is_banned BOOLEAN DEFAULT FALSE,
ADD COLUMN ban_reason TEXT,
ADD COLUMN ban_until TIMESTAMP,
ADD COLUMN is_permanent_ban BOOLEAN DEFAULT FALSE;

-- +goose Down
-- Reverts the banning columns

ALTER TABLE users
DROP COLUMN is_banned,
DROP COLUMN ban_reason,
DROP COLUMN ban_until,
DROP COLUMN is_permanent_ban;
