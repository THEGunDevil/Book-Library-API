package models

import (
	// "encoding/json"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID                uuid.UUID      `json:"id"`                  // unique notification ID
	UserID            uuid.UUID      `json:"user_id"`             // user who receives it
	UserName          string         `json:"user_name"`           // optional, for display
	ObjectID          *uuid.UUID     `json:"object_id,omitempty"` // optional, related object
	ObjectTitle       string         `json:"object_title"`        // optional, title of related object
	Type              string         `json:"type"`                // e.g., BOOK_AVAILABLE, REMINDER, SYSTEM_ALERT
	NotificationTitle string         `json:"notification_title"`  // short title for UI
	Message           string         `json:"message"`             // full message
	// Metadata          json.RawMessage `json:"metadata,omitempty"`  // optional extra info
	IsRead            bool           `json:"is_read"`             // read/unread status
	CreatedAt         time.Time      `json:"created_at"`          // creation timestamp
}
type SendNotificationRequest struct {
	UserID            uuid.UUID      `json:"user_id" binding:"required"`
	UserName          string         `json:"user_name,omitempty"`
	ObjectID          *uuid.UUID     `json:"object_id,omitempty"`
	ObjectTitle       string         `json:"object_title,omitempty"`
	Type              string         `json:"type" binding:"required"`               // e.g., BOOK_AVAILABLE, REMINDER
	NotificationTitle string         `json:"notification_title" binding:"required"` // short title
	Message           string         `json:"message" binding:"required"`            // full message
	// Metadata          json.RawMessage `json:"metadata,omitempty"`                    // optional extra info
}
