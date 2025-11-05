package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ContactRequest struct {
	Name    string `json:"name" binding:"required,min=2,max=50"`
	Email   string `json:"email" binding:"required,email"`
	Subject string `json:"subject" binding:"required,max=200"`
	Message string `json:"message" binding:"required,min=10,max=1000"`
}

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ResendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
	ReplyTo string   `json:"reply_to"`
}

var (
	resendAPIKey = os.Getenv("RESEND_API_KEY")
	resendURL    = "https://api.resend.com/emails"
	fromEmail    = os.Getenv("FROM_EMAIL")
	toEmail      = os.Getenv("TO_EMAIL")
)

func ContactEmailHandler(c *gin.Context) {
	var req ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "Invalid JSON: " + err.Error(),
		})
		return
	}

	// Basic sanitization (remove potential <script> tags)
	req.Message = strings.ReplaceAll(req.Message, "<script>", "")
	req.Message = strings.ReplaceAll(req.Message, "</script>", "")
	req.Subject = strings.ReplaceAll(req.Subject, "<script>", "")
	req.Subject = strings.ReplaceAll(req.Subject, "</script>", "")

	// Use user's subject directly
	subject := req.Subject
	body := fmt.Sprintf(`New contact form submission:

Name: %s
Email: %s
Message: %s

Submitted: %s`, req.Name, req.Email, req.Message, time.Now().Format(time.RFC1123))

	// Send email via Resend
	if err := sendEmail(fromEmail, toEmail, subject, body, req.Email); err != nil {
		log.Printf("Failed to send email: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "Failed to send message. Try again later.",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Message sent successfully!",
	})
}

func sendEmail(from, to, subject, body, replyTo string) error {
	reqBody := ResendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Text:    body,
		ReplyTo: replyTo, // User's email for replies
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("JSON marshal: %w", err)
	}

	idempotencyKey := fmt.Sprintf("contact-%d", time.Now().UnixNano())
	httpReq, err := http.NewRequest("POST", resendURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("HTTP new request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+resendAPIKey)
	httpReq.Header.Set("Idempotency-Key", idempotencyKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("HTTP send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Resend API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	log.Printf("Email sent successfully via Resend to %s", to)
	return nil
}
