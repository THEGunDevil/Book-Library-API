package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"

	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
)

// Helper function to wrap [16]byte UUIDs into pgtype.UUID

func UUIDToPGType(u [16]byte) pgtype.UUID {
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

func StringToPGText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  s != "",
	}
}

// NotificationService handles creating notifications
func NotificationService(ctx context.Context, req models.SendNotificationRequest) error {
	log.Printf("🔔 [DEBUG] NotificationService called for UserID=%v | Type=%s | Title=%s",
		req.UserID, req.Type, req.NotificationTitle)

	// Validate user ID
	if req.UserID == [16]byte{} {
		return fmt.Errorf("invalid UserID")
	}

	// Fetch user info
	u, err := db.Q.GetUserByID(ctx, UUIDToPGType(req.UserID))
	if err != nil {
		log.Printf("❌ [DEBUG] GetUserByID failed for UserID=%v: %v", req.UserID, err)
		return fmt.Errorf("invalid user ID: %w", err)
	}

	userName := fmt.Sprintf("%s %s", u.FirstName, u.LastName)
	log.Printf("👤 [DEBUG] Found user: %s", userName)

	// Marshal metadata to json.RawMessage
	var metadata json.RawMessage
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			log.Printf("❌ [DEBUG] Failed to marshal metadata: %v", err)
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadata = json.RawMessage(metadataBytes)
		log.Printf("🧩 [DEBUG] Metadata JSON: %s", string(metadataBytes))
	} else {
		metadata = json.RawMessage(`{}`)
		log.Printf("⚠️ [DEBUG] No metadata provided, using empty JSON object")
	}

	// Handle ObjectID safely
	var objectID *[16]byte
	if req.ObjectID != nil {
		tmp := [16]byte(*req.ObjectID) // convert uuid.UUID → [16]byte
		objectID = &tmp                // assign pointer
	} else {
		objectID = nil // store NULL in DB
	}

	// Prepare params for sqlc CreateNotification
	arg := gen.CreateNotificationParams{
		UserID:            UUIDToPGType(req.UserID),
		UserName:          StringToPGText(userName),
		ObjectID:          UUIDToPGType(*objectID),
		ObjectTitle:       StringToPGText(req.ObjectTitle),
		Type:              req.Type,
		NotificationTitle: req.NotificationTitle,
		Message:           req.Message,
		Column8:           metadata, // json.RawMessage
	}

	log.Printf("📦 [DEBUG] Inserting notification into DB: %+v", arg)

	// Create notification
	notification, err := db.Q.CreateNotification(ctx, arg)
	if err != nil {
		log.Printf("❌ [DEBUG] Failed to create notification: %v", err)
		return fmt.Errorf("failed to create notification: %w", err)
	}

	log.Printf("✅ [DEBUG] Notification created successfully: ID=%v | UserID=%v", notification.ID, req.UserID)
	return nil
}
