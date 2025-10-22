-- +goose Up
-- Add first_name column if it doesn't exist
ALTER TABLE users
ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT '';

-- Add token_version column if it doesn't exist
ALTER TABLE users
ADD COLUMN IF NOT EXISTS token_version INT NOT NULL DEFAULT 1;
-- Drop the old 'name' column
ALTER TABLE users
DROP COLUMN IF EXISTS name;


-- +goose Down
-- Revert changes: add name back
ALTER TABLE users
ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';

-- Drop first_name column
ALTER TABLE users
DROP COLUMN IF EXISTS first_name;
