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

func CreateSubscriptionPlanHandler(c *gin.Context) {
	log.Println("🔹 CreateSubscriptionPlanHandler called")
	var req models.SubscriptionPlan

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("❌ JSON bind error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("🔹 Request received: %+v\n", req)

	// Validations
	if len(req.Name) == 0 || len(req.Name) > 100 || req.Price <= 0 || req.DurationDays <= 0 || len(req.Description) > 255 {
		log.Println("❌ Validation failed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	featuresBytes, err := json.Marshal(req.Features)
	if err != nil {
		log.Println("❌ Features marshal error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid features"})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription plan", "details": err.Error()})
		return
	}
	log.Println("✅ Subscription plan created with ID:", sub.ID)

	var features map[string]interface{}
	if err := json.Unmarshal([]byte(sub.Features.String), &features); err != nil {
		log.Println("❌ Features unmarshal error:", err)
		features = make(map[string]interface{})
	}

	resp := models.SubscriptionPlan{
		ID:           sub.ID.Bytes,
		Name:         sub.Name,
		Price:        sub.Price,
		DurationDays: int(sub.DurationDays),
		Description:  sub.Description.String,
		Features:     features,
		CreatedAt:    sub.CreatedAt.Time,
		UpdatedAt:    sub.UpdatedAt.Time,
	}

	log.Println("🔹 Response prepared")
	c.JSON(http.StatusOK, gin.H{"message": "subscription plan created", "plan": resp})
}

// ---------------- Update ----------------
func UpdateSubscriptionPlanHandler(c *gin.Context) {
	log.Println("🔹 UpdateSubscriptionPlanHandler called")
	var req models.SubscriptionPlan
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("❌ JSON bind error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("🔹 Request received: %+v\n", req)

	featuresBytes, err := json.Marshal(req.Features)
	if err != nil {
		log.Println("❌ Features marshal error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid features"})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update subscription plans", "details": err.Error()})
		return
	}
	log.Println("✅ Subscription plan updated with ID:", sub.ID)

	var features map[string]interface{}
	if err := json.Unmarshal([]byte(sub.Features.String), &features); err != nil {
		log.Println("❌ Features unmarshal error:", err)
		features = make(map[string]interface{})
	}

	resp := models.SubscriptionPlan{
		ID:           sub.ID.Bytes,
		Name:         sub.Name,
		Price:        sub.Price,
		DurationDays: int(sub.DurationDays),
		Description:  sub.Description.String,
		Features:     features,
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription plan updated", "plan": resp})
}

// ---------------- Get All ----------------
func GetSubscriptionsPlanHandler(c *gin.Context) {
	log.Println("🔹 GetSubscriptionsPlanHandler called")

	sublists, err := db.Q.ListSubscriptionPlans(c.Request.Context())
	if err != nil {
		log.Println("❌ DB fetch error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription plans", "details": err.Error()})
		return
	}
	log.Printf("🔹 %d plans fetched\n", len(sublists))

	var response []models.SubscriptionPlan
	for _, r := range sublists {
		var features map[string]interface{}
		if len(r.Features.String) > 0 {
			if err := json.Unmarshal([]byte(r.Features.String), &features); err != nil {
				log.Println("❌ Features unmarshal error:", err)
				features = make(map[string]interface{})
			}
		} else {
			features = map[string]interface{}{}
		}
		response = append(response, models.SubscriptionPlan{
			ID:           r.ID.Bytes,
			Name:         r.Name,
			Price:        r.Price,
			DurationDays: int(r.DurationDays),
			Description:  r.Description.String,
			Features:     features,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription plans retrieved", "plans": response})
}

// ---------------- Get By ID ----------------
func GetSubscriptionPlanByIDHandler(c *gin.Context) {
	log.Println("🔹 GetSubscriptionPlanByIDHandler called")
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		log.Println("❌ Invalid UUID:", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription plan ID"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get subscription plan", "details": err.Error()})
		return
	}
	var features map[string]interface{}
	if len(sub.Features.String) > 0 {
		if err := json.Unmarshal([]byte(sub.Features.String), &features); err != nil {
			log.Println("❌ Features unmarshal error:", err)
			features = make(map[string]interface{})
		}
	} else {
		features = map[string]interface{}{}
	}
	resp := models.SubscriptionPlan{
		ID:           sub.ID.Bytes,
		Name:         sub.Name,
		Price:        sub.Price,
		DurationDays: int(sub.DurationDays),
		Description:  sub.Description.String,
		Features:     features,
	}

	c.JSON(http.StatusOK, gin.H{"message": "subscription plan retrieved", "plan": resp})
}

// ---------------- Delete ----------------
func DeleteSubscriptionPlanByID(c *gin.Context) {
	log.Println("🔹 DeleteSubscriptionPlanByID called")
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		log.Println("❌ Invalid UUID:", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription plan ID"})
		return
	}

	err = db.Q.DeleteSubscriptionPlan(c.Request.Context(), pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil {
		log.Println("❌ DB delete error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete subscription plan", "details": err.Error()})
		return
	}

	log.Println("✅ Subscription plan deleted:", parsedID)
	c.JSON(http.StatusOK, gin.H{"message": "subscription plan deleted successfully"})
}
