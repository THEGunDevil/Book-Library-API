package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	// "strings"
	"time"

	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
)

// InitializeSSLCommerzPayment calls SSLCommerz API and returns the redirect URL
func InitializeSSLCommerzPayment(payment *gen.Payment) (string, error) {
	payload := map[string]interface{}{
		"store_id":     "YOUR_STORE_ID",
		"store_passwd": "YOUR_STORE_PASSWORD",
		"total_amount": payment.Amount,
		"currency":     payment.Currency,
		"tran_id":      payment.TransactionID.Bytes,
		"success_url":  "https://yourapp.com/payments/callback",
		"fail_url":     "https://yourapp.com/payments/callback",
		"cancel_url":   "https://yourapp.com/payments/callback",
		"cus_name":     "Customer Name",        // replace with actual user name
		"cus_email":    "customer@example.com", // replace with actual user email
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://sandbox.sslcommerz.com/gwprocess/v4/api.php", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("gateway request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Println("SSLCommerz response:", string(body))

	var respData struct {
		GatewayPageURL string `json:"GatewayPageURL"`
		Status         string `json:"status"`
	}
	if err := json.Unmarshal(body, &respData); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if respData.Status != "SUCCESS" || respData.GatewayPageURL == "" {
		return "", fmt.Errorf("gateway initialization failed")
	}

	return respData.GatewayPageURL, nil
}

func InitializeStripePayment(payment *gen.Payment) (string, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		log.Fatal("STRIPE_SECRET_KEY is required")
	}
	log.Printf("Initializing Stripe payment: payment_id=%s, transaction_id=%s, amount=%f, currency=%s",
		payment.ID, payment.TransactionID.String(), payment.Amount, payment.Currency)

	// Basic validations
	if payment.Amount <= 0 {
		err := fmt.Errorf("payment amount must be greater than 0")
		log.Println("Error:", err)
		return "", err
	}
	if payment.Currency == "" {
		err := fmt.Errorf("payment currency is required")
		log.Println("Error:", err)
		return "", err
	}
	successURLBase := os.Getenv("PAYMENT_SUCCESS_URL")
	cancelURL := os.Getenv("PAYMENT_CANCEL_URL")
	successRedirect := os.Getenv("PAYMENT_SUCCESS_REDIRECT")
	cancelRedirect := os.Getenv("PAYMENT_CANCEL_REDIRECT")

	if successURLBase == "" || cancelURL == "" || successRedirect == "" || cancelRedirect == "" {
		return "", fmt.Errorf("missing Stripe URL env vars")
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(payment.Currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Subscription Plan"),
					},
					UnitAmount: stripe.Int64(int64(payment.Amount * 100)),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		// Stripe will replace {CHECKOUT_SESSION_ID} automatically
		SuccessURL: stripe.String(successURLBase + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(cancelURL + "?session_id={CHECKOUT_SESSION_ID}"),
	}
	params.AddMetadata("transaction_id", payment.TransactionID.String())
	params.AddMetadata("payment_id", payment.ID.String())
	params.AddMetadata("user_id", payment.UserID.String()) // safe way for pgtype.UUID
	s, err := session.New(params)
	if err != nil {
		log.Printf("[ERROR] Failed to create Stripe session: %v", err)
		return "", fmt.Errorf("failed to create Stripe session: %w", err)
	}

	log.Printf("Stripe Checkout session created – ID: %s, URL: %s", s.ID, s.URL)
	return s.URL, nil
}

// func InitializeStripePayment(payment *gen.Payment) (string, error) {
// 	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
// 	if stripe.Key == "" {
// 		log.Fatal("STRIPE_SECRET_KEY is required")
// 	}
// 	log.Printf("Initializing Stripe payment: payment_id=%s, transaction_id=%s, amount=%f, currency=%s",
// 		payment.ID, payment.TransactionID.String(), payment.Amount, payment.Currency)

// 	// Basic validations
// 	if payment.Amount <= 0 {
// 		err := fmt.Errorf("payment amount must be greater than 0")
// 		log.Println("Error:", err)
// 		return "", err
// 	}
// 	if payment.Currency == "" {
// 		err := fmt.Errorf("payment currency is required")
// 		log.Println("Error:", err)
// 		return "", err
// 	}
// 	// Convert amount to the smallest currency unit (Stripe expects integer)
// 	unitAmount := int64(payment.Amount * 100)

// 	// Prepare Stripe Checkout session parameters
// 	params := &stripe.CheckoutSessionParams{
// 		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
// 		LineItems: []*stripe.CheckoutSessionLineItemParams{
// 			{
// 				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
//             Currency: stripe.String(strings.ToLower(payment.Currency)),
// 					// Currency: stripe.String("usd"),
// 					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
// 						Name: stripe.String(fmt.Sprintf("Subscription Payment %s", payment.ID.String())),
// 					},
// 					UnitAmount: stripe.Int64(unitAmount),
// 				},
// 				Quantity: stripe.Int64(1),
// 			},
// 		},
// 		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
// 		SuccessURL: stripe.String(fmt.Sprintf(
// 			"https://yourapp.com/payments/callback?status=success&tran_id=%s",
// 			payment.TransactionID.Bytes,
// 		)),
// 		CancelURL: stripe.String(fmt.Sprintf(
// 			"https://yourapp.com/payments/callback?status=cancel&tran_id=%s",
// 			payment.TransactionID.Bytes,
// 		)),
// 	}

// 	log.Println("Creating Stripe Checkout session...")
// 	s, err := session.New(params)
// 	if err != nil {
// 		log.Println("Stripe session creation failed:", err)
// 		return "", fmt.Errorf("failed to create Stripe session: %w", err)
// 	}

// 	if s.URL == "" {
// 		err := fmt.Errorf("Stripe returned empty checkout URL")
// 		log.Println("Error:", err)
// 		return "", err
// 	}

// 	log.Printf("Stripe Checkout session created successfully: session_id=%s, redirect_url=%s", s.ID, s.URL)
// 	return s.URL, nil
// }
