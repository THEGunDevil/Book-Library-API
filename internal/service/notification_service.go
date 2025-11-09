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

// SendNotificationHandler handles sending flexible notifications
func NotificationService(c context.Context, req models.SendNotificationRequest) error {
	log.Printf("🔔 [DEBUG] Starting NotificationService for UserID=%v | Type=%s | Title=%s",
		req.UserID, req.Type, req.NotificationTitle)

	// Fetch user info
	u, err := db.Q.GetUserByID(c, pgtype.UUID{Bytes: req.UserID, Valid: true})
	if err != nil {
		log.Printf("❌ [DEBUG] GetUserByID failed for UserID=%v: %v", req.UserID, err)
		return fmt.Errorf("invalid user ID: %w", err)
	}

	userName := fmt.Sprintf("%s %s", u.FirstName, u.LastName)
	log.Printf("👤 [DEBUG] Found user: %s", userName)

	// Marshal metadata to JSON
	var metadataBytes []byte
	if req.Metadata != nil {
		metadataBytes, err = json.Marshal(req.Metadata)
		if err != nil {
			log.Printf("❌ [DEBUG] Failed to marshal metadata for user=%v: %v", req.UserID, err)
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		if !json.Valid(metadataBytes) {
			log.Printf("❌ [DEBUG] Invalid JSON for metadata: %s", string(metadataBytes))
			return fmt.Errorf("invalid JSON metadata")
		}
		log.Printf("🧩 [DEBUG] Metadata JSON: %s", string(metadataBytes))
	} else {
		metadataBytes = []byte("{}")
		log.Printf("⚠️ [DEBUG] No metadata provided, using empty JSON object")
	}

	// Handle ObjectID safely
	var objectID pgtype.UUID
	if req.ObjectID != nil {
		objectID = pgtype.UUID{Bytes: *req.ObjectID, Valid: true}
		log.Printf("📘 [DEBUG] ObjectID: %v", *req.ObjectID)
	} else {
		log.Printf("⚠️ [DEBUG] No ObjectID provided.")
	}

	// Prepare params
	arg := gen.CreateNotificationParams{
		UserID:            pgtype.UUID{Bytes: req.UserID, Valid: true},
		UserName:          pgtype.Text{String: userName, Valid: true},
		ObjectID:          objectID,
		ObjectTitle:       pgtype.Text{String: req.ObjectTitle, Valid: req.ObjectTitle != ""},
		Type:              req.Type,
		NotificationTitle: req.NotificationTitle,
		Message:           req.Message,
		Column8:          metadataBytes, // Use []byte for JSONB
	}
	log.Printf("📦 [DEBUG] Inserting notification into DB: %+v", arg)

	notification, err := db.Q.CreateNotification(c, arg)
	if err != nil {
		log.Printf("❌ [DEBUG] Failed to create notification in DB: %v", err)
		return fmt.Errorf("failed to create notification: %w", err)
	}

	log.Printf("✅ [DEBUG] Notification created successfully: ID=%v | UserID=%v", notification.ID, req.UserID)
	return nil
}