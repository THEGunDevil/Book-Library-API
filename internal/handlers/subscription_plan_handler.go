package handlers

import (
	"encoding/json"
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

func CreateSubscriptionPlanHandler(c *gin.Context) {
	var req models.SubscriptionPlan

	// Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if len(req.Name) == 0 || len(req.Name) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name must be 1-100 characters"})
		return
	}

	if req.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price must be greater than 0"})
		return
	}

	if req.DurationDays <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "duration_days must be greater than 0"})
		return
	}

	// Optional: limit description length
	if len(req.Description) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description too long"})
		return
	}
	// TODO: Insert into DB
	featuresBytes, err := service.MapToJSONBBytes(req.Features)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid features"})
		return
	}
	params := gen.CreateSubscriptionPlanParams{
		Name:         req.Name,
		Price:        float64(req.Price),
		DurationDays: int32(req.DurationDays),
		Description:  service.StringToPGText(req.Description),
		Features:     featuresBytes,
	}
	sub, err := db.Q.CreateSubscriptionPlan(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription plans",
			"details": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"message": "subscription plan created", "plan": sub})
}
func UpdateSubscriptionPlanHandler(c *gin.Context) {
	var req models.SubscriptionPlan
	// Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate input
	if len(req.Name) == 0 || len(req.Name) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name must be 1-100 characters"})
		return
	}

	if req.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price must be greater than 0"})
		return
	}

	if req.DurationDays <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "duration_days must be greater than 0"})
		return
	}

	// Optional: limit description length
	if len(req.Description) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description too long"})
		return
	}
	// TODO: Insert into DB
	featuresBytes, err := service.MapToJSONBBytes(req.Features)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid features"})
		return
	}
	params := gen.UpdateSubscriptionPlanParams{
		ID:           service.UUIDToPGType(req.ID),
		Name:         req.Name,
		Price:        float64(req.Price),
		DurationDays: int32(req.DurationDays),
		Description:  service.StringToPGText(req.Description),
		Features:     featuresBytes,
	}
	sub, err := db.Q.UpdateSubscriptionPlan(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription plans",
			"details": err.Error()})
	}
	var response []models.SubscriptionPlan
	response = append(response, models.SubscriptionPlan{
		ID:           sub.ID.Bytes, // assuming r.ID is uuid.UUID or string
		Name:         sub.Name,
		Price:        sub.Price, // if using pgtype.Float8
		DurationDays: int(sub.DurationDays),
		Description:  sub.Description.String, // if pgtype.Text
		Features:     req.Features,             // []byte or map depending on SQLC
	})
	c.JSON(http.StatusOK, gin.H{"message": "subscription update created", "plan": response})
}
func GetSubscriptionsPlanHandler(c *gin.Context) {
	// Get subscription plans from DB
	sublists, err := db.Q.ListSubscriptionPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get subscription plans",
			"details": err.Error(),
		})
		return
	}

	// Map to your API response model if needed
	var response []models.SubscriptionPlan
	for _, r := range sublists {
		var features map[string]interface{}
		if len(r.Features) > 0 {
			if err := json.Unmarshal(r.Features, &features); err != nil {
				features = map[string]interface{}{} // fallback empty map
			}
		} else {
			features = map[string]interface{}{}
		}
		response = append(response, models.SubscriptionPlan{
			ID:           r.ID.Bytes, // assuming r.ID is uuid.UUID or string
			Name:         r.Name,
			Price:        r.Price, // if using pgtype.Float8
			DurationDays: int(r.DurationDays),
			Description:  r.Description.String, // if pgtype.Text
			Features:     features,             // []byte or map depending on SQLC
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "subscription plans retrieved",
		"plans":   response,
	})
}
func GetSubscriptionPlanByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}
	sub, err := db.Q.GetSubscriptionPlanByID(c.Request.Context(), pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil {
		if err == pgx.ErrNoRows { // important
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription", "details": err.Error()})
		return
	}

	// Map to your API response model if needed
	var response []models.SubscriptionPlan
	var features map[string]interface{}
	if len(sub.Features) > 0 {
		if err := json.Unmarshal(sub.Features, &features); err != nil {
			features = map[string]interface{}{} // fallback empty map
		}
	} else {
		features = map[string]interface{}{}
	}
	response = append(response, models.SubscriptionPlan{
		ID:           sub.ID.Bytes, // assuming sub.ID is uuid.UUID or string
		Name:         sub.Name,
		Price:        sub.Price, // if using pgtype.Float8
		DurationDays: int(sub.DurationDays),
		Description:  sub.Description.String, // if pgtype.Text
		Features:     features,               // []byte or map depending on SQLC
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "subscription plans retrieved",
		"plans":   response,
	})
}
func DeleteSubscriptionPlanByID(c *gin.Context) {
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}
	err = db.Q.DeleteSubscriptionPlan(c.Request.Context(), pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to delete subscription plan",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "subscription plan deleted successfully ",
	})
}
