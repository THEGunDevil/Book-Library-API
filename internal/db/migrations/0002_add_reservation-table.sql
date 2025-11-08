-- +goose Up
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    book_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    notified_at TIMESTAMPTZ,
    fulfilled_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_book FOREIGN KEY (book_id) REFERENCES books (id) ON DELETE CASCADE
);

-- ✅ A user can only have one reservation per book (any status)
CREATE UNIQUE INDEX unique_user_book ON reservations (user_id, book_id);

-- ⚡ Makes queries faster when fetching/sorting by book
CREATE INDEX idx_book_created_at ON reservations (book_id, created_at);

-- +goose Down
DROP TABLE IF EXISTS reservations;
