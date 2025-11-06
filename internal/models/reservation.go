package models

import (
	"time"

	"github.com/google/uuid"
)

type ReservationResponse struct {
    ID          uuid.UUID      `json:"id"`           // Reservation UUID
    UserID      uuid.UUID     `json:"user_id"`      // User who made the reservation
    BookID      uuid.UUID     `json:"book_id"`      // Reserved book
    Status      string     `json:"status"`       // pending | notified | fulfilled | cancelled
    CreatedAt   time.Time  `json:"created_at"`   // When reservation was made
    NotifiedAt  *time.Time `json:"notified_at"`  // When user was notified (nullable)
    FulfilledAt *time.Time `json:"fulfilled_at"` // When reservation was converted to borrow (nullable)
    CancelledAt *time.Time `json:"cancelled_at"` // When reservation was cancelled (nullable)
}

type CreateReservationParams struct {
    UserID uuid.UUID
    BookID uuid.UUID
}
type UpdateReservationStatusRequest struct {
    Status string `json:"status" binding:"required,oneof=pending notified fulfilled cancelled"`
}
