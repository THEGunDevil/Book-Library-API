package models

import (
	"time"

	"github.com/google/uuid"
)

type CreateBorrowRequest struct {
	BookID  uuid.UUID `json:"book_id"`
	DueDate string    `json:"due_date"`
}
type ReturnBookRequest struct {
	BookID uuid.UUID `json:"book_id"`
}
type BorrowResponse struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	UserName   string     `json:"user_name"`
	BookID     uuid.UUID  `json:"book_id"`
	BookTitle  string     `json:"book_title"`
	BorrowedAt time.Time  `json:"borrowed_at"`
	DueDate    time.Time  `json:"due_date"`
	ReturnedAt *time.Time `json:"returned_at"`
}
