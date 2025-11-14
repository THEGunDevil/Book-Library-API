package models

import (
	"time"

	"github.com/google/uuid"
)

// ReservationResponse - existing model (keep as is)
type ReservationResponse struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	BookID      uuid.UUID   `json:"book_id"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	NotifiedAt  time.Time   `json:"notified_at"`
	FulfilledAt time.Time   `json:"fulfilled_at"`
	CancelledAt time.Time   `json:"cancelled_at"`
	PickedUp    bool        `json:"picked_up"`
	UserName    interface{} `json:"user_name"`
	UserEmail   string      `json:"email"`
	BookTitle   string      `json:"book_title"`
	BookAuthor  string      `json:"author"`
	BookImage   string      `json:"image_url"`
}

// ReservationListResponse - new model for filtered lists
type ReservationListResponse struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	BookID      uuid.UUID   `json:"book_id"`
	Status      string      `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	NotifiedAt  *time.Time  `json:"notified_at,omitempty"`
	FulfilledAt *time.Time  `json:"fulfilled_at,omitempty"`
	CancelledAt *time.Time  `json:"cancelled_at,omitempty"`
	PickedUp    bool        `json:"picked_up"`
	UserName    interface{} `json:"user_name"`
	UserEmail   string      `json:"email"`
	BookTitle   string      `json:"book_title"`
	BookAuthor  string      `json:"author"`
	BookImage   string      `json:"image_url"`
}

type CreateReservationParams struct {
	UserID uuid.UUID
	BookID uuid.UUID
}
type UpdateReservationStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending notified fulfilled cancelled"`
}
