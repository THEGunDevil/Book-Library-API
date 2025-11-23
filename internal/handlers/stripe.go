package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/THEGunDevil/GoForBackend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
)

func processPaidCheckoutSession(ctx context.Context, checkoutSession *stripe.CheckoutSession) error {
	transactionIDStr, ok := checkoutSession.Metadata["transaction_id"]
	if !ok || transactionIDStr == "" {
		return fmt.Errorf("missing transaction_id in metadata")
	}

	tranUUID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		return fmt.Errorf("invalid transaction_id UUID: %v", err)
	}

	payment, err := db.Q.GetPaymentByTransactionID(ctx, pgtype.UUID{Bytes: tranUUID, Valid: true})
	if err != nil {
		return fmt.Errorf("payment not found: %v", err)
	}

	if payment.Status == "paid" {
		return nil // already processed — idempotent
	}

	tx, err := db.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txQueries := gen.New(tx)

	plan, err := txQueries.GetSubscriptionPlanByID(ctx, payment.PlanID)
	if err != nil {
		return err
	}

	start := time.Now().UTC()
	end := start.Add(time.Duration(plan.DurationDays) * 24 * time.Hour)

	sub, err := txQueries.CreateSubscription(ctx, gen.CreateSubscriptionParams{
		UserID:    payment.UserID,
		PlanID:    payment.PlanID,
		StartDate: pgtype.Timestamp{Time: start, Valid: true},
		EndDate:   pgtype.Timestamp{Time: end, Valid: true},
		Status:    "active",
	})
	if err != nil {
		return err
	}

	if _, err = txQueries.UpdatePaymentStatus(ctx, gen.UpdatePaymentStatusParams{
		ID:     payment.ID,
		Status: "paid",
	}); err != nil {
		return err
	}

	if _, err = txQueries.UpdatePaymentSubscriptionID(ctx, gen.UpdatePaymentSubscriptionIDParams{
		ID:             payment.ID,
		SubscriptionID: pgtype.UUID{Bytes: sub.ID.Bytes, Valid: true},
	}); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return err
	}
	// Send notification asynchronously
	go func() {
		// Use a fresh background context (don't use the request ctx – it may be cancelled)
		bgCtx := context.Background()

		notifReq := models.SendNotificationRequest{
			UserID:            payment.UserID.Bytes, // Adjust if needed
			Type:              "subscription_created",
			NotificationTitle: fmt.Sprintf("Subscription to %s Activated!", plan.Name),
			Message: fmt.Sprintf("Your subscription is active from %s to %s.",
				start.Format("January 02, 2006"), end.Format("January 02, 2006")),
			ObjectID:    nil, // or &sub.ID.UUID if you want to link it
			ObjectTitle: plan.Name,
		}

		if err := service.NotificationService(bgCtx, notifReq); err != nil {
			log.Printf("⚠️ [Webhook] Failed to send subscription notification: %v", err)
			// Optional: send to Sentry/Slack for monitoring
		} else {
			log.Printf("✅ [Webhook] Subscription notification sent for user %s", payment.UserID)
		}
	}()
	log.Printf("✅ [SuccessHandler] Payment %s marked as PAID, subscription %s created", transactionIDStr, sub.ID)
	return nil
}
func StripeSuccessHandler(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.String(http.StatusBadRequest, "Missing session_id")
		return
	}

	s, err := session.Get(sessionID, nil)
	if err != nil {
		log.Printf("❌ [Success] Failed to retrieve session %s: %v", sessionID, err)
		c.String(http.StatusBadRequest, "Invalid session")
		return
	}

	if s.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		// Redirect to cancel page if somehow not paid
		http.Redirect(c.Writer, c.Request, os.Getenv("PAYMENT_CANCEL_REDIRECT"), http.StatusSeeOther)
		return
	}

	if err := processPaidCheckoutSession(c.Request.Context(), s); err != nil {
		log.Printf("❌ [Success] Failed to process paid session %s: %v", sessionID, err)
		// You can show an error page or still redirect (user already paid)
		// Most people just redirect to success anyway
	}

	tranID := s.Metadata["transaction_id"]
	redirectURL := os.Getenv("PAYMENT_SUCCESS_REDIRECT")
	if tranID != "" {
		redirectURL += "?tran_id=" + tranID
	}

	http.Redirect(c.Writer, c.Request, redirectURL, http.StatusSeeOther)
}

func StripeCancelHandler(c *gin.Context) {
	// Optional but recommended: retrieve the session to confirm it was actually cancelled
	// and mark the payment as "cancelled" in your DB (idempotently
	sessionID := c.Query("session_id")

	if sessionID != "" {
		s, err := session.Get(sessionID, nil)
		if err != nil {
			// If we can't retrieve the session, just redirect anyway – user already cancelled
			log.Printf("⚠️ [Cancel] Could not retrieve session %s: %v", sessionID, err)
		} else if s.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
			// Only mark as cancelled if it really wasn't paid
			if tranID, ok := s.Metadata["transaction_id"]; ok && tranID != "" {
				if uuid, err := uuid.Parse(tranID); err == nil {
					if _, err := db.Q.UpdatePaymentStatusByTransactionID(c.Request.Context(), gen.UpdatePaymentStatusByTransactionIDParams{
						TransactionID: pgtype.UUID{Bytes: uuid, Valid: true},
						Status:        "cancelled",
					}); err != nil {
						log.Printf("❌ [Cancel] Failed to mark payment as cancelled: %v", err)
					} else {
						log.Printf("✅ [Cancel] Payment %s marked as cancelled", tranID)
					}
				}
			}
		}
	}

	redirectURL := os.Getenv("PAYMENT_CANCEL_REDIRECT")
	if redirectURL == "" {
		redirectURL = "https://book-library-web.vercel.app/payment/cancelled" // fallback
	}

	// You can append ?tran_id=... if you have it, but it's optional for cancel page
	http.Redirect(c.Writer, c.Request, redirectURL, http.StatusSeeOther)
}
