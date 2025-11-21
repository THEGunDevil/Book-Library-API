package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func CreateSubscriptionHandler(c *gin.Context) {
	// Get userID from middleware (secure)
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userUUID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	// Parse request
	var req models.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.PlanID.String() == "" {
		c.JSON(400, gin.H{"error": "plan_id is required"})
		return
	}

	planUUID := uuid.UUID(req.PlanID)

	// Fetch the plan
	plan, err := db.Q.GetSubscriptionPlanByID(c.Request.Context(), pgtype.UUID{
		Bytes: planUUID,
		Valid: true,
	})
	if err != nil {
		c.JSON(404, gin.H{"error": "plan not found"})
		return
	}

	// Check if user already has an active subscription
	_, err = db.Q.GetSubscriptionByID(c.Request.Context(), pgtype.UUID{
		Bytes: userUUID,
		Valid: true,
	})

	if err == nil {
		// user already has active subs
		c.JSON(403, gin.H{"error": "you already have an active subscription"})
		return
	}

	// Calculate dates from the plan
	start := time.Now().UTC()
	end := start.Add(time.Duration(plan.DurationDays) * 24 * time.Hour)

	// Create subscription
	params := gen.CreateSubscriptionParams{
		UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
		PlanID: pgtype.UUID{Bytes: planUUID, Valid: true},
		StartDate: pgtype.Timestamp{
			Time:  start,
			Valid: true,
		},
		EndDate: pgtype.Timestamp{
			Time:  end,
			Valid: true,
		},
		Status:    "active", // or "pending" if using payment
		AutoRenew: req.AutoRenew,
	}

	sub, err := db.Q.CreateSubscription(c.Request.Context(), params)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create subscription", "details": err.Error()})
		return
	}

	response := models.Subscription{
		ID:        sub.ID.Bytes,
		UserID:    sub.UserID.Bytes,
		PlanID:    sub.PlanID.Bytes,
		StartDate: sub.StartDate.Time,
		EndDate:   sub.EndDate.Time,
		CreatedAt: sub.CreatedAt.Time,
		UpdatedAt: sub.UpdatedAt.Time,
		Status:    sub.Status,
		AutoRenew: sub.AutoRenew,
	}

	c.JSON(200, gin.H{"message": "subscription created", "subscription": response})
}

func GetSubscriptionByIDHandler(c *gin.Context) {
	idStr := c.Param("id")

	subID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid subscription ID"})
		return
	}

	sub, err := db.Q.GetSubscriptionByID(c.Request.Context(), pgtype.UUID{
		Bytes: subID, Valid: true,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch subscription", "details": err.Error()})
		return
	}
	var response []models.Subscription
	response = append(response, models.Subscription{
		ID:        sub.ID.Bytes,
		UserID:    sub.PlanID.Bytes,
		PlanID:    sub.PlanID.Bytes,
		StartDate: sub.EndDate.Time,
		EndDate:   sub.EndDate.Time,
		CreatedAt: sub.CreatedAt.Time,
		UpdatedAt: sub.UpdatedAt.Time,
		Status:    sub.Status,
		AutoRenew: sub.AutoRenew,
	})
	c.JSON(200, gin.H{"subscription": response})
}
func GetSubscriptionByUserIDHandler(c *gin.Context) {
	idStr := c.Param("user_id")

	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription user ID"})
		return
	}

	sub, err := db.Q.GetSubscriptionByUserID(c.Request.Context(), pgtype.UUID{
		Bytes: userID,
		Valid: true,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"subscription": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch subscription", "details": err.Error()})
		}
		return
	}

	response := models.Subscription{
		ID:        sub.ID.Bytes,
		UserID:    sub.UserID.Bytes,
		PlanID:    sub.PlanID.Bytes,
		StartDate: sub.StartDate.Time,
		EndDate:   sub.EndDate.Time,
		CreatedAt: sub.CreatedAt.Time,
		UpdatedAt: sub.UpdatedAt.Time,
		Status:    sub.Status,
		AutoRenew: sub.AutoRenew,
	}

	c.JSON(200, gin.H{"subscription": response})
}

func ListSubscriptionsHandler(c *gin.Context) {
	subs, err := db.Q.ListSubscriptions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to fetch subscriptions",
			"details": err.Error(),
		})
		return
	}

	var response []models.Subscription

	for _, sub := range subs {
		response = append(response, models.Subscription{
			ID:        sub.ID.Bytes,     // UUID
			UserID:    sub.UserID.Bytes, // Corrected: UserID
			PlanID:    sub.PlanID.Bytes,
			StartDate: sub.StartDate.Time,
			EndDate:   sub.EndDate.Time,
			Status:    sub.Status,
			AutoRenew: sub.AutoRenew,
			CreatedAt: sub.CreatedAt.Time,
			UpdatedAt: sub.UpdatedAt.Time,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"subscriptions": response, // fixed typo
	})
}
func ListSubscriptionsByUserHandler(c *gin.Context) {
	userID := c.Param("user_id")

	parsed, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid user_id"})
		return
	}

	subs, err := db.Q.ListSubscriptionsByUser(c.Request.Context(), pgtype.UUID{
		Bytes: parsed, Valid: true,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch user subscriptions", "details": err.Error()})
		return
	}
	var response []models.Subscription

	for _, sub := range subs {
		response = append(response, models.Subscription{
			ID:        sub.ID.Bytes,     // UUID
			UserID:    sub.UserID.Bytes, // Corrected: UserID
			PlanID:    sub.PlanID.Bytes,
			StartDate: sub.StartDate.Time,
			EndDate:   sub.EndDate.Time,
			Status:    sub.Status,
			AutoRenew: sub.AutoRenew,
			CreatedAt: sub.CreatedAt.Time,
			UpdatedAt: sub.UpdatedAt.Time,
		})
	}
	c.JSON(200, gin.H{"subscriptions": response})
}
func UpdateSubscriptionHandler(c *gin.Context) {
	idStr := c.Param("id")
	subID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid subscription ID"})
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	start := time.Now().UTC()
	end := start.Add(time.Duration(req.DurationDays) * 24 * time.Hour)
	params := gen.UpdateSubscriptionParams{
		ID:        pgtype.UUID{Bytes: subID, Valid: true},
		PlanID:    pgtype.UUID{Bytes: req.PlanID, Valid: true},
		StartDate: pgtype.Timestamp{Time: start, Valid: true},
		EndDate:   pgtype.Timestamp{Time: end, Valid: true},
		AutoRenew: *req.AutoRenew,
	}

	updated, err := db.Q.UpdateSubscription(c.Request.Context(), params)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to update subscription", "details": err.Error()})
		return
	}
	var response []models.Subscription
	response = append(response, models.Subscription{
		ID:        updated.ID.Bytes,
		UserID:    updated.PlanID.Bytes,
		PlanID:    updated.PlanID.Bytes,
		StartDate: updated.EndDate.Time,
		EndDate:   updated.EndDate.Time,
		CreatedAt: updated.CreatedAt.Time,
		UpdatedAt: updated.UpdatedAt.Time,
		Status:    updated.Status,
		AutoRenew: updated.AutoRenew,
	})
	c.JSON(200, gin.H{"message": "subscription updated", "subscription": response})
}
func DeleteSubscriptionByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	subID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid subscription ID"})
		return
	}

	err = db.Q.DeleteSubscription(c.Request.Context(), pgtype.UUID{
		Bytes: subID, Valid: true,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "failed to delete subscription", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "subscription deleted"})
}
