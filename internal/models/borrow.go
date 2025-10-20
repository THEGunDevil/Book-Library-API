package models

import (
	"time"

	"github.com/google/uuid"
)

type CreateBorrowRequest struct {
	BookID  uuid.UUID `json:"book_id"`
	DueDate string    `json:"due_date"`
}
type UpdateBorrowRequest struct {
	UserID     uuid.UUID `json:"user_id"`
	BookID     uuid.UUID `json:"book_id"`
	ReturnedAt string    `json:"returned_at"`
}
type BorrowResponse struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	BookID     uuid.UUID  `json:"book_id"`
	BorrowedAt time.Time  `json:"borrowed_at"`
	DueDate    time.Time  `json:"due_date"`
	ReturnedAt *time.Time `json:"returned_at,omitempty"`
}
