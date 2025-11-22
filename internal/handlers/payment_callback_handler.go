package handlers

import (
	"context" // Added context for consistency
	"encoding/json"
	"errors" // New import for errors.As
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
	"github.com/jackc/pgx/v5/pgconn" // New import for checking PG error codes
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

// Helper function to create an invalid UUID for error handling
func getInvalidUUID() pgtype.UUID {
	return pgtype.UUID{Bytes: uuid.Nil, Valid: false}
}

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
			log.Println("ℹ️ [Webhook] Payment already marked as paid. Skipping fulfillment.")
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
		plan, err := txQueries.GetSubscriptionPlanByID(ctx, payment.PlanID)
		if err != nil {
			log.Printf("❌ [Webhook] Subscription plan not found for ID: %s", payment.PlanID)
			c.Status(http.StatusInternalServerError)
			return
		}
		start := time.Now().UTC()
		end := start.Add(time.Duration(plan.DurationDays) * 24 * time.Hour)

		var sub gen.Subscription
		// Create Subscription
		sub, err = txQueries.CreateSubscription(ctx, gen.CreateSubscriptionParams{
			UserID:    payment.UserID,
			PlanID:    payment.PlanID,
			StartDate: pgtype.Timestamp{Time: start, Valid: true},
			EndDate:   pgtype.Timestamp{Time: end, Valid: true},
			Status:    "active",
		})
		
		// --- CRITICAL ERROR HANDLING BLOCK ---
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				// 23505 is the PostgreSQL code for unique_violation.
				// This usually means the user already has an active subscription.
				log.Printf("⚠️ [Webhook] DB Constraint Failed (23505): User %s already has an active subscription. Proceeding to mark payment as paid.", payment.UserID.String)
				
				// Set sub ID to invalid so we skip linking it below, but allow the payment status update to proceed.
				sub.ID = getInvalidUUID() 
			} else {
				// A real, unexpected DB error occurred (e.g., NOT NULL violation, connection lost).
				log.Printf("❌ [Webhook] DB: CreateSubscription failed unexpectedly: %v", err)
				c.Status(http.StatusInternalServerError)
				return
			}
		}
		// --- END CRITICAL BLOCK ---

		// Update Payment Status to PAID (IDEMPOTENCY)
		_, err = txQueries.UpdatePaymentStatus(ctx, gen.UpdatePaymentStatusParams{
			ID:     payment.ID,
			Status: "paid",
		})
		if err != nil {
			log.Printf("❌ [Webhook] DB: UpdatePaymentStatus failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		// Update Payment with Sub ID (ONLY if subscription creation succeeded)
		if sub.ID.Valid {
			_, err = txQueries.UpdatePaymentSubscriptionID(ctx, gen.UpdatePaymentSubscriptionIDParams{
				ID:             payment.ID,
				SubscriptionID: pgtype.UUID{Bytes: sub.ID.Bytes, Valid: true},
			})
			if err != nil {
				log.Printf("❌ [Webhook] DB: UpdatePaymentStatus with subscription ID failed: %v", err)
				c.Status(http.StatusInternalServerError)
				return
			}
			log.Printf("✅ [Webhook] SUCCESS! Payment %s updated to PAID, Sub %s created and linked.", transactionID, sub.ID.String())
		} else {
			log.Printf("✅ [Webhook] SUCCESS! Payment %s updated to PAID, Subscription creation skipped (pre-existing sub constraint).", transactionID)
		}
		
		if err := tx.Commit(ctx); err != nil {
			log.Printf("❌ [Webhook] DB: Commit failed: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

	default:
		log.Printf("ℹ️ [Webhook] Unhandled event type: %s", event.Type)
	}

	c.Status(http.StatusOK)
}