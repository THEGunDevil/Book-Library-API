-- +goose Up
-- Add bio column
ALTER TABLE users
ADD COLUMN IF NOT EXISTS bio TEXT NOT NULL DEFAULT '';

-- +goose Down
-- Drop bio column
ALTER TABLE users
DROP COLUMN IF EXISTS bio;