package handlers

import (
	"encoding/json" // Import this!
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

func StripeWebhookHandler(c *gin.Context) {
	// 1. Read Body
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("❌ [Webhook] Failed to read request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// 2. Verify Signature
	sigHeader := c.GetHeader("Stripe-Signature")
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	log.Printf("🔍 [Webhook] Received event. Signature: %s... Secret Length: %d",
		sigHeader[:10], len(endpointSecret))

	event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
	if err != nil {
		log.Printf("❌ [Webhook] Signature verification failed: %v", err)
		log.Println("💡 Tip: Ensure STRIPE_WEBHOOK_SECRET matches the one in Stripe Dashboard > Developers > Webhooks")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook signature"})
		return
	}

	log.Printf("✅ [Webhook] Event Verified. Type: %s", event.Type)

	// 3. Handle Event
	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			log.Printf("❌ [Webhook] JSON Unmarshal error: %v", err)
			c.Status(http.StatusBadRequest)
			return
		}

		// Log Metadata to debug missing ID
		log.Printf("🔍 [Webhook] Metadata Received: %+v", session.Metadata)

		transactionID := session.Metadata["transaction_id"]
		if transactionID == "" {
			log.Println("❌ [Webhook] transaction_id MISSING in metadata. InitializeStripePayment might be wrong.")
			c.Status(http.StatusBadRequest)
			return
		}

		tranUUID, err := uuid.Parse(transactionID)
		if err != nil {
			log.Printf("❌ [Webhook] Invalid UUID format: %s", transactionID)
			c.Status(http.StatusBadRequest)
			return
		}

		// Database Operations
		ctx := c.Request.Context()

		// Check if payment exists
		payment, err := db.Q.GetPaymentByTransactionID(ctx, pgtype.UUID{Bytes: tranUUID, Valid: true})
		if err != nil {
			log.Printf("❌ [Webhook] DB: Payment not found for ID: %s", transactionID)
			c.Status(http.StatusNotFound)
			return
		}

		// Check status
		if payment.Status == "paid" {
			log.Println("ℹ️ [Webhook] Payment already marked as paid. Skipping.")
			c.Status(http.StatusOK)
			return
		}

		// Begin Transaction
		tx, err := db.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			log.Printf("❌ [Webhook] DB: Failed to start transaction: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}
		defer tx.Rollback(ctx)
		txQueries := gen.New(tx)

		// Calculate Subscription Dates
		plan, _ := txQueries.GetSubscriptionPlanByID(ctx, payment.PlanID)
		start := time.Now().UTC()
		end := start.Add(time.Duration(plan.DurationDays) * 24 * time.Hour)

		// Create Subscription
		sub, err := txQueries.CreateSubscription(ctx, gen.CreateSubscriptionParams{
			UserID:    payment.UserID,
			PlanID:    payment.PlanID,
			StartDate: pgtype.Timestamp{Time: start, Valid: true},
			EndDate:   pgtype.Timestamp{Time: end, Valid: true},
			Status:    "active",
		})
		if err != nil {
			log.Printf("❌ [Webhook] DB: CreateSubscription failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		// Update Payment
		err = txQueries.UpdatePaymentStatus(ctx, gen.UpdatePaymentStatusParams{
			ID:     payment.ID,
			Status: "paid",
		})
		if err != nil {
			log.Printf("❌ [Webhook] DB: UpdatePaymentStatus failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		// Update Payment with Sub ID
		_, err = txQueries.UpdatePaymentSubscriptionID(ctx, gen.UpdatePaymentSubscriptionIDParams{
			ID:             payment.ID,
			SubscriptionID: pgtype.UUID{Bytes: sub.ID.Bytes, Valid: true},
		})

		// Commit
		if err := tx.Commit(ctx); err != nil {
			log.Printf("❌ [Webhook] DB: Commit failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		log.Printf("✅ [Webhook] SUCCESS! Payment %s updated to PAID, Sub %s created.", transactionID, sub.ID)

	default:
		log.Printf("ℹ️ [Webhook] Unhandled event type: %s", event.Type)
	}

	c.Status(http.StatusOK)
}
func StripeRedirectHandler(c *gin.Context) {
	tranID := c.Query("tran_id")
	tranUUID, err := uuid.Parse(tranID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction_id"})
		return
	}

	payment, err := db.Q.GetPaymentByTransactionID(c.Request.Context(), pgtype.UUID{
		Bytes: tranUUID,
		Valid: true,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}

	status := payment.Status
	c.JSON(http.StatusOK, gin.H{
		"message":        "Payment processed",
		"status":         status,
		"transaction_id": tranID,
	})
}
