package service

import (
	"context"
	"fmt"
	"log"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// Converts uuid.UUID → pgtype.UUID
func UUIDToPGType(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

// Converts string → pgtype.Text
func StringToPGText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  s != "",
	}
}

// NotificationService handles creating event-based notifications
func NotificationService(ctx context.Context, req models.SendNotificationRequest) error {
	log.Printf("🔔 [DEBUG] NotificationService called for Type=%s | Title=%s",
		req.Type, req.NotificationTitle)

	// Validate optional ObjectID
	var pgObjectID pgtype.UUID
	if req.ObjectID != nil {
		pgObjectID = UUIDToPGType(*req.ObjectID)
	} else {
		pgObjectID = pgtype.UUID{Valid: false} // NULL in DB
	}

	// Prepare params for CreateEvent
	eventArg := gen.CreateEventParams{
		ObjectID:    pgObjectID,
		ObjectTitle: StringToPGText(req.ObjectTitle),
		Type:        req.Type,
		Title:       req.NotificationTitle,
		Message:     req.Message,
		Metadata:    req.Metadata, // optional JSONB
	}

	// Insert event into events table
	event, err := db.Q.CreateEvent(ctx, eventArg)
	if err != nil {
		log.Printf("❌ [DEBUG] Failed to create event: %v", err)
		return fmt.Errorf("failed to create event: %w", err)
	}
	log.Printf("✅ [DEBUG] Event created successfully: ID=%v", event.ID)

	// Optional: create initial unread status for a specific user (if sending to one user)
	if req.UserID != uuid.Nil {
		statusArg := gen.UpsertUserNotificationStatusParams{
			UserID:  UUIDToPGType(req.UserID),
			EventID: event.ID,
		}
		err := db.Q.UpsertUserNotificationStatus(ctx, statusArg)
		if err != nil {
			log.Printf("❌ Failed to insert user notification status: %v", err)
			return fmt.Errorf("failed to create user notification status: %w", err)
		}
		log.Printf("✅ User notification status created for UserID=%v", req.UserID)

	}

	return nil
}
