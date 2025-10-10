-- +goose Up
-- Add first_name column
ALTER TABLE users
ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT '';

-- Copy data from name to first_name (optional)
UPDATE users SET first_name = name;

-- Drop the old name column
ALTER TABLE users
DROP COLUMN IF EXISTS name;

-- +goose Down
-- Revert changes: add name back
ALTER TABLE users
ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';

-- Copy data back from first_name to name
UPDATE users SET name = first_name;

-- Drop first_name column
ALTER TABLE users
DROP COLUMN IF EXISTS first_name;
