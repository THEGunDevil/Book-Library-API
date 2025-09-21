package models

import "time"

type CreateBorrowRequest struct {
	UserID     string `json:"user_id"`
	BookID     string `json:"book_id"`
	DueDate    string `json:"due_date"`
}
type UpdateBorrowRequest struct {
	UserID     string `json:"user_id"`
	BookID     string `json:"book_id"`
	ReturnedAt    string `json:"returned_at"`
}
type BorrowResponse struct {
	ID         string `json:"id"`
	UserID     string `json:"user_id"`
	BookID     string `json:"book_id"`
	BorrowedAt time.Time `json:"borrowed_at"`
	DueDate    time.Time `json:"due_date"`
	ReturnedAt *time.Time `json:"returned_at,omitempty"`
}
