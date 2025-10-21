-- +goose Up
-- Add bio column
ALTER TABLE users
ADD COLUMN IF NOT EXISTS bio TEXT NOT NULL DEFAULT '';
-- Add description column
ALTER TABLE books
ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
-- Add genre column
ALTER TABLE books
ADD COLUMN IF NOT EXISTS genre TEXT NOT NULL DEFAULT '';

-- +goose Down
-- Drop bio column
ALTER TABLE users
DROP COLUMN IF EXISTS bio;
-- Drop description column
ALTER TABLE books
DROP COLUMN IF EXISTS description;

-- Drop genre column
ALTER TABLE books
DROP COLUMN IF EXISTS genre;