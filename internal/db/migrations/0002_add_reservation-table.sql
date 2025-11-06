-- +goose Up
CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    book_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    notified_at TIMESTAMPTZ,
    fulfilled_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_book FOREIGN KEY(book_id) REFERENCES books(id) ON DELETE CASCADE
);

-- Unique constraint to prevent duplicate pending reservations
CREATE UNIQUE INDEX unique_user_book_pending
ON reservations(user_id, book_id)
WHERE status = 'pending';

-- Index for fast queue retrieval per book
CREATE INDEX idx_book_created_at
ON reservations(book_id, created_at);

-- +goose Down
DROP TABLE IF EXISTS reservations;
