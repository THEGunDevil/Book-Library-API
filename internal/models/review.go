package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CreateReviewRequest struct {
	UserID  uuid.UUID `json:"user_id"`
	BookID  uuid.UUID `json:"book_id"`
	Rating  int       `json:"rating"`
	Comment string    `json:"comment"`
}
type Review struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	BookID    uuid.UUID `json:"book_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type UpdateReviewRequest struct {
	Rating  *int32       `json:"rating,omitempty"`
	Comment *string    `json:"comment,omitempty"`
}

// Optional: Validation helper
func (r *Review) Validate() error {
	if r.Rating < 1 || r.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}
	if r.Comment == "" {
		return fmt.Errorf("comment cannot be empty")
	}
	return nil
}
