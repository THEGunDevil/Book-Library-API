package service

import (
	"context"
	// "encoding/json"
	"fmt"
	"log"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ✅ Converts uuid.UUID → pgtype.UUID
func UUIDToPGType(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

// ✅ Converts string → pgtype.Text
func StringToPGText(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  s != "",
	}
}

// ✅ NotificationService handles creating notifications
func NotificationService(ctx context.Context, req models.SendNotificationRequest) error {
	log.Printf("🔔 [DEBUG] NotificationService called for UserID=%v | Type=%s | Title=%s",
		req.UserID, req.Type, req.NotificationTitle)

	// Validate user ID
	if req.UserID == uuid.Nil {
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

	// // ✅ Marshal metadata safely
	// var meta json.RawMessage
	// if len(req.Metadata) > 0 {
	// 	meta = req.Metadata
	// } else {
	// 	meta = json.RawMessage(`{}`)
	// }

	// ✅ Handle ObjectID safely (*uuid.UUID → *[16]byte)
	var pgObjectID pgtype.UUID
	if req.ObjectID != nil {
		pgObjectID = UUIDToPGType(*req.ObjectID)
	} else {
		pgObjectID = pgtype.UUID{Valid: false} // NULL in DB
	}

	// ✅ Prepare params for sqlc CreateNotification
	arg := gen.CreateNotificationParams{
		UserID:            UUIDToPGType(req.UserID),
		UserName:          StringToPGText(userName),
		ObjectID:          pgObjectID,
		ObjectTitle:       StringToPGText(req.ObjectTitle),
		Type:              req.Type,
		NotificationTitle: req.NotificationTitle,
		Message:           req.Message,
		// Column8:           meta, // ✅ correct type
	}

	log.Printf("📦 [DEBUG] Inserting notification into DB: %+v", arg)

	// ✅ Create notification
	notification, err := db.Q.CreateNotification(ctx, arg)
	if err != nil {
		log.Printf("❌ [DEBUG] Failed to create notification: %v", err)
		return fmt.Errorf("failed to create notification: %w", err)
	}

	log.Printf("✅ [DEBUG] Notification created successfully: ID=%v | UserID=%v", notification.ID, req.UserID)
	return nil
}
