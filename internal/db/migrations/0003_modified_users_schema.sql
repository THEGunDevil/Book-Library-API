-- +goose Up
-- Rename 'name' to 'first_name'
ALTER TABLE users
RENAME COLUMN name TO first_name;

-- Add 'last_name' and 'phone_number' columns
ALTER TABLE users
ADD COLUMN last_name TEXT NOT NULL DEFAULT '',
ADD COLUMN phone_number TEXT;

-- +goose Down
-- Revert changes
ALTER TABLE users
RENAME COLUMN first_name TO name;
ALTER TABLE users
DROP COLUMN last_name;
ALTER TABLE users
DROP COLUMN phone_number;
