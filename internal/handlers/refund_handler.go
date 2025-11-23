package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/THEGunDevil/GoForBackend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

	ref, err := db.Q.CreateRefund(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refund", "details": err.Error()})
		return
	}

	res := models.Refund{
		ID:          ref.ID.Bytes,
		PaymentID:   ref.PaymentID.Bytes,
		Amount:      ref.Amount,
		Reason:      ref.Reason.String,
		RequestedAt: ref.RequestedAt.Time,
		ProcessedAt: &ref.ProcessedAt.Time,
		Status:      ref.Status,
	}

	c.JSON(http.StatusOK, gin.H{"message": "refund created", "refund": res})
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
	// 1. Validate Refund ID from URL
	idStr := c.Param("id")
	refundID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refund ID format"})
		return
	}

	// 2. Bind and validate request payload
	var req models.CreateRefundRequest // Assuming this struct contains the desired Status
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload", "details": err.Error()})
		return
	}

	// 3. Validate status value
	validStatuses := map[string]bool{"requested": true, "processed": true, "rejected": true}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
		return
	}

	// --- TRANSACTION START ---
	// 4. Begin a database transaction
	tx, err := db.DB.Begin(c.Request.Context()) // Assuming 'db.Pool' is your *pgxpool.Pool
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start database transaction", "details": err.Error()})
		return
	}
	defer func() {
		// Rollback is safe to call even if the transaction is already committed.
		// It only executes if the transaction is still active (e.g., due to an earlier error).
		_ = tx.Rollback(c.Request.Context())
	}()

	// Use the transaction-aware query object
	txQ := db.Q.WithTx(tx) // Assuming this method returns a query object bound to the transaction

	// 5. Fetch the existing Refund to get the PaymentID (Step: Refund -> Payment)
	originalRefund, err := txQ.GetRefundByID(c.Request.Context(), pgtype.UUID{Bytes: refundID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "refund not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch refund details", "details": err.Error()})
		return
	}

	// 6. Fetch the Payment using the PaymentID to get the SubscriptionID (Step: Payment -> Subscription)
	// NOTE: You must have a query function like GetPaymentByID that returns a struct containing SubscriptionID
	payment, err := txQ.GetPaymentByID(c.Request.Context(), originalRefund.PaymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found (data inconsistency)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch payment details", "details": err.Error()})
		return
	}

	// 7. Prepare and execute the Refund status update
	var processedAt pgtype.Timestamp
	if req.Status == "processed" {
		processedAt = pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	}

	updateRefundParams := gen.UpdateRefundStatusParams{
		ID:          pgtype.UUID{Bytes: refundID, Valid: true},
		Status:      req.Status,
		ProcessedAt: processedAt,
	}

	updatedRefund, err := txQ.UpdateRefundStatus(c.Request.Context(), updateRefundParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update refund status", "details": err.Error()})
		return
	}

	// 8. Conditionally update Subscription status (only if refund is processed)
	if updatedRefund.Status == "processed" {
		_, err = txQ.UpdateSubscription(c.Request.Context(), gen.UpdateSubscriptionParams{
			ID:     payment.SubscriptionID, // Use the ID fetched from the payment record
			Status: "cancelled",
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel associated subscription", "details": err.Error()})
			return
		}
	}

	// 9. Commit the transaction
	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction", "details": err.Error()})
		return
	}
	// --- TRANSACTION END ---
	if updatedRefund.Status == "processed" {
		// 1. Convert the database's raw UUID ([16]byte) to the standard uuid.UUID type
		refundUUID := uuid.UUID(updatedRefund.ID.Bytes)
		userUUID := uuid.UUID(payment.UserID.Bytes)

		// 2. Define the anonymous function to accept arguments (r, u, amount)
		go func(r gen.Refund, u uuid.UUID, amount float64) {
			// 3. Define the actual context
			bgCtx := context.Background()

			// 4. Use the passed arguments to construct the request
			notifReq := models.SendNotificationRequest{
				UserID:            u, // Use the passed 'u'
				Type:              "refund_processed",
				NotificationTitle: fmt.Sprintf("Refund Processed: %s", r.Status),                               // Use the passed 'r'
				Message:           fmt.Sprintf("Your refund of %.2f has been successfully processed.", amount), // Use the passed 'amount'
				ObjectID:          &refundUUID,
				ObjectTitle:       "Refund",
			}

			if err := service.NotificationService(bgCtx, notifReq); err != nil {
				log.Printf("⚠️ [RefundHandler] Failed to send refund notification: %v", err)
			} else {
				log.Printf("✅ [RefundHandler] Refund notification sent for user %s", u)
			}

			// 5. Pass the values when calling the anonymous function
		}(updatedRefund, userUUID, updatedRefund.Amount)
	}
	// 10. Construct final successful response
	response := models.Refund{
		ID:          updatedRefund.ID.Bytes,
		PaymentID:   updatedRefund.PaymentID.Bytes,
		Amount:      updatedRefund.Amount,
		Reason:      originalRefund.Reason.String, // Using original reason for response
		Status:      updatedRefund.Status,
		RequestedAt: updatedRefund.RequestedAt.Time,
		ProcessedAt: &updatedRefund.ProcessedAt.Time,
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
