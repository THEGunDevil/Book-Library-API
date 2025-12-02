package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/THEGunDevil/GoForBackend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Helper function to unmarshal features safely
func unmarshalFeatures(featuresStr string) map[string]interface{} {
	var features map[string]interface{}
	
	if len(featuresStr) == 0 || featuresStr == "{}" {
		return make(map[string]interface{})
	}
	
	if err := json.Unmarshal([]byte(featuresStr), &features); err != nil {
		log.Println("❌ Features unmarshal error:", err)
		return make(map[string]interface{})
	}
	
	return features
}

// Helper function to build subscription plan response
func buildSubscriptionPlanResponse(sub gen.SubscriptionPlan) models.SubscriptionPlan {
	features := unmarshalFeatures(sub.Features.String)
	
	return models.SubscriptionPlan{
		ID:           sub.ID.Bytes,
		Name:         sub.Name,
		Price:        sub.Price,
		DurationDays: int(sub.DurationDays),
		Description:  sub.Description.String,
		Features:     features,
		CreatedAt:    sub.CreatedAt.Time,
		UpdatedAt:    sub.UpdatedAt.Time,
	}
}

// CreateSubscriptionPlanHandler creates a new subscription plan
func CreateSubscriptionPlanHandler(c *gin.Context) {
	log.Println("🔹 CreateSubscriptionPlanHandler called")
	var req models.SubscriptionPlan

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("❌ JSON bind error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("🔹 Request received: Name=%s, Price=%f, Duration=%d\n", req.Name, req.Price, req.DurationDays)

	// Validations
	if len(req.Name) == 0 || len(req.Name) > 100 {
		log.Println("❌ Validation failed: invalid name")
		c.JSON(http.StatusBadRequest, gin.H{"error": "name must be between 1 and 100 characters"})
		return
	}

	if req.Price <= 0 {
		log.Println("❌ Validation failed: invalid price")
		c.JSON(http.StatusBadRequest, gin.H{"error": "price must be greater than 0"})
		return
	}

	if req.DurationDays <= 0 {
		log.Println("❌ Validation failed: invalid duration")
		c.JSON(http.StatusBadRequest, gin.H{"error": "duration days must be greater than 0"})
		return
	}

	if len(req.Description) > 255 {
		log.Println("❌ Validation failed: description too long")
		c.JSON(http.StatusBadRequest, gin.H{"error": "description must not exceed 255 characters"})
		return
	}

	// Initialize features if nil
	if req.Features == nil {
		req.Features = make(map[string]interface{})
	}

	featuresBytes, err := json.Marshal(req.Features)
	if err != nil {
		log.Println("❌ Features marshal error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid features format"})
		return
	}

	log.Printf("🔹 Features JSON: %s\n", string(featuresBytes))

	params := gen.CreateSubscriptionPlanParams{
		Name:         req.Name,
		Price:        float64(req.Price),
		DurationDays: int32(req.DurationDays),
		Description:  service.StringToPGText(req.Description),
		Features:     pgtype.Text{String: string(featuresBytes), Valid: true},
	}
	log.Println("🔹 SQLC params prepared")

	sub, err := db.Q.CreateSubscriptionPlan(c.Request.Context(), params)
	if err != nil {
		log.Println("❌ DB insert error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create subscription plan",
			"details": err.Error(),
		})
		return
	}
	log.Println("✅ Subscription plan created with ID:", sub.ID)

	resp := buildSubscriptionPlanResponse(sub)

	c.JSON(http.StatusOK, gin.H{
		"message": "subscription plan created successfully",
		"plan":    resp,
	})
}

// UpdateSubscriptionPlanHandler updates an existing subscription plan
func UpdateSubscriptionPlanHandler(c *gin.Context) {
	log.Println("🔹 UpdateSubscriptionPlanHandler called")
	var req models.SubscriptionPlan

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("❌ JSON bind error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate ID is provided
	if req.ID == uuid.Nil {
		log.Println("❌ Validation failed: ID is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription plan ID is required"})
		return
	}

	log.Printf("🔹 Request received: ID=%s, Name=%s, Price=%f, Duration=%d\n", req.ID, req.Name, req.Price, req.DurationDays)

	// Validations
	if len(req.Name) == 0 || len(req.Name) > 100 {
		log.Println("❌ Validation failed: invalid name")
		c.JSON(http.StatusBadRequest, gin.H{"error": "name must be between 1 and 100 characters"})
		return
	}

	if req.Price <= 0 {
		log.Println("❌ Validation failed: invalid price")
		c.JSON(http.StatusBadRequest, gin.H{"error": "price must be greater than 0"})
		return
	}

	if req.DurationDays <= 0 {
		log.Println("❌ Validation failed: invalid duration")
		c.JSON(http.StatusBadRequest, gin.H{"error": "duration days must be greater than 0"})
		return
	}

	if len(req.Description) > 255 {
		log.Println("❌ Validation failed: description too long")
		c.JSON(http.StatusBadRequest, gin.H{"error": "description must not exceed 255 characters"})
		return
	}

	// Initialize features if nil
	if req.Features == nil {
		req.Features = make(map[string]interface{})
	}

	featuresBytes, err := json.Marshal(req.Features)
	if err != nil {
		log.Println("❌ Features marshal error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid features format"})
		return
	}

	log.Printf("🔹 Features JSON: %s\n", string(featuresBytes))

	params := gen.UpdateSubscriptionPlanParams{
		ID:           service.UUIDToPGType(req.ID),
		Name:         req.Name,
		Price:        float64(req.Price),
		DurationDays: int32(req.DurationDays),
		Description:  service.StringToPGText(req.Description),
		Features:     pgtype.Text{String: string(featuresBytes), Valid: true},
	}
	log.Println("🔹 SQLC params prepared")

	sub, err := db.Q.UpdateSubscriptionPlan(c.Request.Context(), params)
	if err != nil {
		log.Println("❌ DB update error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update subscription plan",
			"details": err.Error(),
		})
		return
	}
	log.Println("✅ Subscription plan updated with ID:", sub.ID)

	resp := buildSubscriptionPlanResponse(sub)

	c.JSON(http.StatusOK, gin.H{
		"message": "subscription plan updated successfully",
		"plan":    resp,
	})
}

// GetSubscriptionsPlanHandler retrieves all subscription plans
func GetSubscriptionsPlanHandler(c *gin.Context) {
	log.Println("🔹 GetSubscriptionsPlanHandler called")

	sublists, err := db.Q.ListSubscriptionPlans(c.Request.Context())
	if err != nil {
		log.Println("❌ DB fetch error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get subscription plans",
			"details": err.Error(),
		})
		return
	}
	log.Printf("🔹 %d plans fetched\n", len(sublists))

	var response []models.SubscriptionPlan
	for _, r := range sublists {
		response = append(response, buildSubscriptionPlanResponse(r))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "subscription plans retrieved successfully",
		"plans":   response,
	})
}

// GetSubscriptionPlanByIDHandler retrieves a single subscription plan by ID
func GetSubscriptionPlanByIDHandler(c *gin.Context) {
	log.Println("🔹 GetSubscriptionPlanByIDHandler called")
	idStr := c.Param("id")
	
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		log.Println("❌ Invalid UUID:", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription plan ID format"})
		return
	}

	sub, err := db.Q.GetSubscriptionPlanByID(c.Request.Context(), pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Println("❌ Subscription plan not found:", parsedID)
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription plan not found"})
			return
		}
		log.Println("❌ DB fetch error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get subscription plan",
			"details": err.Error(),
		})
		return
	}

	resp := buildSubscriptionPlanResponse(sub)

	c.JSON(http.StatusOK, gin.H{
		"message": "subscription plan retrieved successfully",
		"plan":    resp,
	})
}

// DeleteSubscriptionPlanByID deletes a subscription plan by ID
func DeleteSubscriptionPlanByID(c *gin.Context) {
	log.Println("🔹 DeleteSubscriptionPlanByID called")
	idStr := c.Param("id")

	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		log.Println("❌ Invalid UUID:", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription plan ID format"})
		return
	}

	err = db.Q.DeleteSubscriptionPlan(c.Request.Context(), pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil {
		log.Println("❌ DB delete error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to delete subscription plan",
			"details": err.Error(),
		})
		return
	}

	log.Println("✅ Subscription plan deleted:", parsedID)
	c.JSON(http.StatusOK, gin.H{"message": "subscription plan deleted successfully"})
}