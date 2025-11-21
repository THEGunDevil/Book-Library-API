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
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if endpointSecret == "" {
		log.Println("❌ STRIPE_WEBHOOK_SECRET not set")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "webhook secret not configured"})
		return
	}

	event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
	if err != nil {
		log.Println("❌ Webhook signature verification failed:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook signature"})
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		// FIX: Unmarshal the raw event data into the specific Stripe struct
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			log.Printf("❌ Error parsing webhook JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}

		// Extract metadata from the correctly typed 'session' object
		transactionID := session.Metadata["transaction_id"]
		if transactionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "transaction_id missing in metadata"})
			return
		}

		tranUUID, err := uuid.Parse(transactionID)
		if err != nil {
			log.Println("❌ Invalid transaction_id UUID:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction_id"})
			return
		}

		ctx := c.Request.Context()

		// Fetch payment record
		payment, err := db.Q.GetPaymentByTransactionID(ctx, pgtype.UUID{
			Bytes: tranUUID,
			Valid: true,
		})
		if err != nil {
			log.Println("❌ Payment record not found:", transactionID)
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}

		if payment.Status == "paid" { // Check pgtype.Text
			log.Println("ℹ Payment already processed:", transactionID)
			c.JSON(http.StatusOK, gin.H{"message": "payment already processed"})
			return
		}

		tx, err := db.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			log.Println("❌ Failed to begin DB transaction:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		txQueries := gen.New(tx)

		defer func() {
			if err := tx.Rollback(ctx); err != nil &&
				err.Error() != "pgx: transaction has already been committed or rolled back" {
				log.Println("⚠ Failed to rollback transaction:", err)
			}
		}()

		// Check the payment status from the Stripe session object
		if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
			log.Printf("💰 Stripe payment verified as PAID for tran_id=%s", transactionID)

			// Fetch subscription plan
			plan, err := txQueries.GetSubscriptionPlanByID(ctx, payment.PlanID)
			if err != nil {
				log.Println("❌ Failed to load subscription plan:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load plan"})
				return
			}

			start := time.Now().UTC()
			end := start.Add(time.Duration(plan.DurationDays) * 24 * time.Hour)

			// Create subscription
			subscription, err := txQueries.CreateSubscription(ctx, gen.CreateSubscriptionParams{
				UserID: payment.UserID,
				PlanID: payment.PlanID,
				StartDate: pgtype.Timestamp{
					Time:  start,
					Valid: true,
				},
				EndDate: pgtype.Timestamp{
					Time:  end,
					Valid: true,
				},
				Status: "active", // Ensure status is set
			})
			if err != nil {
				log.Println("❌ Failed to create subscription:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create subscription"})
				return
			}

			// Update payment with subscription ID
			_, err = txQueries.UpdatePaymentSubscriptionID(ctx, gen.UpdatePaymentSubscriptionIDParams{
				ID: payment.ID,
				SubscriptionID: pgtype.UUID{
					Bytes: subscription.ID.Bytes,
					Valid: true,
				},
			})
			if err != nil {
				log.Println("❌ Failed to update payment subscription ID:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update payment"})
				return
			}

			// Update payment status to PAID
			_, err = txQueries.UpdatePaymentStatus(ctx, gen.UpdatePaymentStatusParams{
				ID:     payment.ID,
				Status: "paid",
			})
			if err != nil {
				log.Println("❌ Failed to mark payment as paid:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update payment status"})
				return
			}

			if err := tx.Commit(ctx); err != nil {
				log.Println("❌ Failed to commit transaction:", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed"})
				return
			}

			log.Printf("✅ Payment successful & subscription created! transaction_id=%s subscription_id=%s", transactionID, subscription.ID)

		} else {
			// Payment failed or incomplete
			log.Println("❌ Stripe payment FAILED/INCOMPLETE for tran_id:", transactionID)
			_, err = txQueries.UpdatePaymentStatus(ctx, gen.UpdatePaymentStatusParams{
				ID:     payment.ID,
				Status: "failed",
			})
			if err != nil {
				log.Println("❌ Failed to mark payment as failed:", err)
			}
			_ = tx.Commit(ctx)
		}

	default:
		log.Println("ℹ Unhandled event type:", event.Type)
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
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

