package handlers

import (
	"net/http"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// =======================
// Payments Handlers
// =======================

func CreatePaymentHandler(c *gin.Context) {
	var req models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload", "details": err.Error()})
		return
	}

	// Validate required fields
	if req.UserID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	if req.SubscriptionID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription_id is required"})
		return
	}
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return
	}

	// Set default values if empty
	if req.Currency == "" {
		req.Currency = "BDT"
	}
	if req.Status == "" {
		req.Status = "pending"
	}

	params := gen.CreatePaymentParams{
		UserID:         pgtype.UUID{Bytes: req.UserID, Valid: true},
		SubscriptionID: pgtype.UUID{Bytes: req.SubscriptionID, Valid: true},
		Amount:         req.Amount,
		Currency:       pgtype.Text{String: req.Currency, Valid: true},
		TransactionID:  pgtype.Text{String: req.TransactionID, Valid: req.TransactionID != ""},
		PaymentGateway: pgtype.Text{String: req.PaymentGateway, Valid: req.PaymentGateway != ""},
		Status:         pgtype.Text{String: req.Status, Valid: true},
	}

	payment, err := db.Q.CreatePayment(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment created", "payment": payment})
}

func GetPaymentHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	payment, err := db.Q.GetPaymentByID(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get payment", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment": payment})
}

func ListPaymentsByUserHandler(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	payments, err := db.Q.ListPaymentsByUser(c.Request.Context(), pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list payments", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payments": payments})
}

func UpdatePaymentStatusHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload", "details": err.Error()})
		return
	}

	// Validate status
	validStatuses := map[string]bool{"paid": true, "failed": true, "pending": true}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
		return
	}

	params := gen.UpdatePaymentStatusParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		Status: pgtype.Text{String: req.Status, Valid: true},
	}

	payment, err := db.Q.UpdatePaymentStatus(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update payment status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment": payment})
}

func DeletePaymentByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	if err := db.Q.DeletePayment(c.Request.Context(), pgtype.UUID{Bytes: id,Valid: true}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete payment", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "payment deleted"})
}
