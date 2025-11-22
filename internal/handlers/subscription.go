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
		// Use http.StatusBadRequest (400) constant
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) 
		return
	}

	// Assuming models.CreateSubscriptionRequest uses uuid.UUID for PlanID
	if req.PlanID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plan_id is required"})
		return
	}

	planUUID := uuid.UUID(req.PlanID)

	// Fetch the plan
	plan, err := db.Q.GetSubscriptionPlanByID(c.Request.Context(), pgtype.UUID{
		Bytes: planUUID,
		Valid: true,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch plan", "details": err.Error()})
		}
		return
	}

	// Check if user already has an active subscription
	_, err = db.Q.GetSubscriptionByUserID(c.Request.Context(), pgtype.UUID{
		Bytes: userUUID,
		Valid: true,
	})

	if err == nil {
		// Subscription found (due to the previously recommended SQL fix)
		c.JSON(http.StatusForbidden, gin.H{"error": "you already have an active subscription"})
		return
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		// Handle unexpected database error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error checking subscription status", "details": err.Error()})
		return
	}
	
	// If the flow reaches here, err is pgx.ErrNoRows, and we proceed to create the subscription.

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
		Status:    "active", // Assuming this is for manual creation/admin use without a payment step
		AutoRenew: req.AutoRenew,
	}

	sub, err := db.Q.CreateSubscription(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription", "details": err.Error()})
		return
	}

	response := models.Subscription{
		ID:          sub.ID.Bytes,
		UserID:      sub.UserID.Bytes,
		PlanID:      sub.PlanID.Bytes,
		StartDate:   sub.StartDate.Time,
		EndDate:     sub.EndDate.Time,
		CreatedAt:   sub.CreatedAt.Time,
		UpdatedAt:   sub.UpdatedAt.Time,
		Status:      sub.Status,
		AutoRenew:   sub.AutoRenew,
	}

	// ✅ Use http.StatusCreated (201) for successful resource creation
	c.JSON(http.StatusCreated, gin.H{"message": "subscription created", "subscription": response})
}

func GetSubscriptionByIDHandler(c *gin.Context) {
    idStr := c.Param("id")

    subID, err := uuid.Parse(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID"})
        return
    }

    sub, err := db.Q.GetSubscriptionByID(c.Request.Context(), pgtype.UUID{
        Bytes: subID, Valid: true,
    })

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch subscription", "details": err.Error()})
        return
    }

    // Map to a single response struct
    response := models.Subscription{
        ID:          sub.ID.Bytes,
        UserID:      sub.UserID.Bytes,    // CORRECTED
        PlanID:      sub.PlanID.Bytes,
        StartDate:   sub.StartDate.Time,  // CORRECTED
        EndDate:     sub.EndDate.Time,
        CreatedAt:   sub.CreatedAt.Time,
        UpdatedAt:   sub.UpdatedAt.Time,
        Status:      sub.Status,
        AutoRenew:   sub.AutoRenew,
    }

    c.JSON(http.StatusOK, gin.H{"subscription": response}) // Return single object
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
    
    // Response should be a single object, not a slice
    response := models.Subscription{
        ID:          updated.ID.Bytes,
        UserID:      updated.UserID.Bytes,    // CORRECTED
        PlanID:      updated.PlanID.Bytes,
        StartDate:   updated.StartDate.Time,  // CORRECTED
        EndDate:     updated.EndDate.Time,
        CreatedAt:   updated.CreatedAt.Time,
        UpdatedAt:   updated.UpdatedAt.Time,
        Status:      updated.Status,
        AutoRenew:   updated.AutoRenew,
    }

    c.JSON(http.StatusOK, gin.H{"message": "subscription updated", "subscription": response}) // Return single object
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
