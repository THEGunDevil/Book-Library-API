package handlers

import (
	"net/http"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetUserNotificationByUserIDHandler(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "userID not found in context"})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid userID type"})
		return
	}

	notifications, err := db.Q.GetUserNotificationsByUserID(
		c.Request.Context(),
		pgtype.UUID{Bytes: userID, Valid: true},
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var response []models.Notification
	for _, n := range notifications {
		// var metadata map[string]any
		// if n.Metadata != nil {
		// 	if err := json.Unmarshal(n.Metadata, &metadata); err != nil {
		// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unmarshal metadata"})
		// 		return
		// 	}
		// }

		var objectID *uuid.UUID
		if n.ObjectID.Valid {
			id := n.ObjectID.Bytes
			uuidVal, err := uuid.FromBytes(id[:])
			if err == nil {
				objectID = &uuidVal
			}
		}

		response = append(response, models.Notification{
			ID:                n.ID.Bytes,
			UserID:            n.UserID.Bytes,
			UserName:          n.UserName.String,
			ObjectID:          objectID,
			ObjectTitle:       n.ObjectTitle.String,
			Type:              n.Type,
			NotificationTitle: n.NotificationTitle,
			Message:           n.Message,
			// Metadata:          n.Metadata,
			IsRead:    n.IsRead.Bool,
			CreatedAt: n.CreatedAt.Time,
		})
	}

	c.JSON(http.StatusOK, response)
}

func MarkNotificationAsReadByUserID(c *gin.Context) {
	// ✅ Get the authenticated user's ID from middleware context
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userUUID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID in context"})
		return
	}

	// ✅ Call DB query
	err := db.Q.MarkNotificationAsReadByUserID(
		c.Request.Context(),
		pgtype.UUID{Bytes: userUUID, Valid: true},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notifications marked as read successfully"})
}

