-- +goose Up
ALTER TABLE books
    ADD COLUMN image_url TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE books
    DROP COLUMN image_url TEXT NOT NULL;
