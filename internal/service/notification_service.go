package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
)

// SendNotificationHandler handles sending flexible notifications
func NotificationService(c context.Context, req models.SendNotificationRequest) error {

	// Fetch user from DB to get name
	u, err := db.Q.GetUserByID(c, pgtype.UUID{Bytes: req.UserID, Valid: true})
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)

	}
	userName := u.FirstName + " " + u.LastName

	metadataBytes, err := json.Marshal(req.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	var objectID pgtype.UUID
	if req.ObjectID != nil {
		objectID = pgtype.UUID{Bytes: *req.ObjectID, Valid: true}
	}
	arg := gen.CreateNotificationParams{
		UserID:            pgtype.UUID{Bytes: req.UserID, Valid: true},
		UserName:          pgtype.Text{String: userName, Valid: true},
		ObjectID:          objectID,
		ObjectTitle:       pgtype.Text{String: req.ObjectTitle, Valid: req.ObjectTitle != ""},
		Type:              req.Type,
		NotificationTitle: req.NotificationTitle,
		Message:           req.Message,
		Metadata:          metadataBytes,
	}
	if err := db.Q.CreateNotification(c, arg); err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}
