package models

import (
	"time"

	"github.com/google/uuid"
)

type SubscriptionPlan struct {
	ID           uuid.UUID              `json:"id"`
	Name         string                 `json:"name"`
	Price        float64                `json:"price"`         // NUMERIC(10,2)
	DurationDays int                    `json:"duration_days"` // e.g., 30, 365
	Description  string                 `json:"description"`
	Features     map[string]interface{} `json:"features"` // JSONB
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type Subscription struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	PlanID    uuid.UUID `json:"plan_id"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Status    string    `json:"status"` // active, expired, cancelled
	AutoRenew bool      `json:"auto_renew"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type CreateSubscriptionRequest struct {
	UserID       uuid.UUID `json:"user_id"`
	PlanID       uuid.UUID `json:"plan_id"`
	DurationDays int       `json:"duration_days"`
	Status       string    `json:"status"`     // optional, "active", "expired", "cancelled"
	AutoRenew    bool      `json:"auto_renew"` // optional
}
type UpdateSubscriptionRequest struct {
	UserID       uuid.UUID `json:"user_id"`
	PlanID       uuid.UUID `json:"plan_id"`
	DurationDays int       `json:"duration_days"`
	AutoRenew    *bool     `json:"auto_renew"` // optional
}

type Payment struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	SubscriptionID *uuid.UUID `json:"subscription_id,omitempty"` // nullable link
	Amount         float64    `json:"amount"`
	Currency       string     `json:"currency"`
	TransactionID  uuid.UUID  `json:"transaction_id"` // FIXED: Should be uuid.UUID to match DB and server generation
	PaymentGateway string     `json:"payment_gateway"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
}
type Refund struct {
	ID          uuid.UUID  `json:"id"`
	PaymentID   uuid.UUID  `json:"payment_id"`
	Amount      float64    `json:"amount"`
	Reason      string     `json:"reason,omitempty"`
	Status      string     `json:"status binding:"required,oneof=requested processed rejected"` // requested, processed, rejected
	RequestedAt time.Time  `json:"requested_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"` // pointer to allow NULL
}
type CreateRefundRequest struct {
	PaymentID uuid.UUID `json:"payment_id"`       // required
	Amount    float64   `json:"amount"`           // required
	Reason    string    `json:"reason,omitempty"` // optional
	Status    string    `json:"status"`           // requested, processed, rejected
}
type CreatePaymentRequest struct {
	UserID uuid.UUID `json:"user_id"`
	PlanID uuid.UUID `json:"plan_id"`
}
