package handlers

import (
	"net/http"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// =======================
// Refunds Handlers
// =======================

func CreateRefundHandler(c *gin.Context) {
	var req models.CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload", "details": err.Error()})
		return
	}

	// Validate required fields
	if req.PaymentID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment_id is required"})
		return
	}
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than 0"})
		return
	}

	// Set default status if empty
	if req.Status == "" {
		req.Status = "requested"
	}

	params := gen.CreateRefundParams{
		PaymentID:   pgtype.UUID{Bytes: req.PaymentID, Valid: true},
		Amount:      req.Amount,
		Reason:      pgtype.Text{String: req.Reason, Valid: req.Reason != ""},
		Status:      req.Status,
		ProcessedAt: pgtype.Timestamp{Valid: false}, // initially NULL
	}

	refund, err := db.Q.CreateRefund(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refund", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "refund created", "refund": refund})
}

func GetRefundHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refund ID"})
		return
	}

	refund, err := db.Q.GetRefundByID(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get refund", "details": err.Error()})
		return
	}

	// Convert pgtype fields to Go types
	var processedAt *time.Time
	if refund.ProcessedAt.Valid {
		processedAt = &refund.ProcessedAt.Time
	}

	var reason string
	if refund.Reason.Valid {
		reason = refund.Reason.String
	}

	response := models.Refund{
		ID:          refund.ID.Bytes,
		PaymentID:   refund.PaymentID.Bytes,
		Amount:      refund.Amount,
		Reason:      reason,
		Status:      refund.Status,
		RequestedAt: refund.RequestedAt.Time,
		ProcessedAt: processedAt,
	}

	c.JSON(http.StatusOK, gin.H{"refund": response})
}

func ListRefundsByPaymentHandler(c *gin.Context) {
	paymentIDStr := c.Param("payment_id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment ID"})
		return
	}

	refunds, err := db.Q.ListRefundsByPayment(c.Request.Context(), pgtype.UUID{Bytes: paymentID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list refunds", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"refunds": refunds})
}

func ListRefundsByStatusHandler(c *gin.Context) {
	status := c.Query("status")
	validStatuses := map[string]bool{"requested": true, "processed": true, "rejected": true}
	if !validStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status filter"})
		return
	}

	refunds, err := db.Q.ListRefundsByStatus(c.Request.Context(), status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list refunds", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"refunds": refunds})
}

func UpdateRefundStatusHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refund ID"})
		return
	}

	var req struct {
		Status      string    `json:"status"`
		ProcessedAt time.Time `json:"processed_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload", "details": err.Error()})
		return
	}

	validStatuses := map[string]bool{"requested": true, "processed": true, "rejected": true}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
		return
	}

	params := gen.UpdateRefundStatusParams{
		ID:          pgtype.UUID{Bytes: id, Valid: true},
		Status:      req.Status,
		ProcessedAt: pgtype.Timestamp{Time: req.ProcessedAt, Valid: true},
	}

	refund, err := db.Q.UpdateRefundStatus(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update refund", "details": err.Error()})
		return
	}

	var processedAt *time.Time
	if refund.ProcessedAt.Valid {
		processedAt = &refund.ProcessedAt.Time
	}

	var reason string
	if refund.Reason.Valid {
		reason = refund.Reason.String
	}

	response := models.Refund{
		ID:          refund.ID.Bytes,
		PaymentID:   refund.PaymentID.Bytes,
		Amount:      refund.Amount,
		Reason:      reason,
		Status:      refund.Status,
		RequestedAt: refund.RequestedAt.Time,
		ProcessedAt: processedAt,
	}

	c.JSON(http.StatusOK, gin.H{"refund": response})
}

func DeleteRefundHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refund ID"})
		return
	}

	if err := db.Q.DeleteRefund(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete refund", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "refund deleted"})
}
