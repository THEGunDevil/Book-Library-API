// package handlers

// import (
// 	"crypto/tls"
// 	"fmt"
// 	"html"
// 	"log"
// 	"net/http"
// 	"os"
// 	"regexp"

// 	"github.com/gin-gonic/gin"
// 	gomail "gopkg.in/mail.v2"
// )

// // ContactRequest defines the user-submitted contact form data
// type ContactRequest struct {
// 	Name    string `json:"name" binding:"required,max=100"`
// 	Email   string `json:"email" binding:"required,email"`
// 	Subject string `json:"subject" binding:"required,max=200"`
// 	Message string `json:"message" binding:"required,max=5000"`
// }

// // ContactEmailHandler sends contact form submissions to a configured email
// func ContactEmailHandler(c *gin.Context) {
// 	var req ContactRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
// 		return
// 	}

// 	// Additional validation for email format
// 	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
// 	if !emailRegex.MatchString(req.Email) {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
// 		return
// 	}

// 	// Load email configuration from environment variables
// 	smtpUser := os.Getenv("SMTP_USER")
// 	smtpPass := os.Getenv("SMTP_APP_PASSWORD")
// 	smtpHost := os.Getenv("SMTP_HOST")
// 	smtpPort := 587 // Default Gmail SMTP port
// 	toEmail := os.Getenv("CONTACT_EMAIL")

// 	if smtpUser == "" || smtpPass == "" || smtpHost == "" || toEmail == "" {
// 		log.Println("Missing SMTP configuration")
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server email not configured"})
// 		return
// 	}

// 	// Create email message
// 	m := gomail.NewMessage()
// 	m.SetHeader("From", smtpUser)
// 	m.SetHeader("To", toEmail)
// 	m.SetHeader("Subject", fmt.Sprintf("[Contact Form] %s", html.EscapeString(req.Subject)))
// 	// Escape inputs to prevent HTML injection
// 	m.SetBody("text/html", fmt.Sprintf(
// 		"<p><strong>Name:</strong> %s</p>"+
// 			"<p><strong>Email:</strong> %s</p>"+
// 			"<p><strong>Message:</strong><br/>%s</p>",
// 		html.EscapeString(req.Name),
// 		html.EscapeString(req.Email),
// 		html.EscapeString(req.Message),
// 	))

// 	// Configure SMTP dialer
// 	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
// 	d.TLSConfig = &tls.Config{ServerName: smtpHost} // Ensure proper TLS configuration

// 	if err := d.DialAndSend(m); err != nil {
// 		log.Printf("Failed to send email: %T %v", err, err) // log type and message
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

//		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Message sent successfully"})
//	}
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/resend/resend-go/v2"
)

type ContactRequest struct {
	Name    string `json:"name" binding:"required,max=100"`
	Email   string `json:"email" binding:"required,email"`
	Subject string `json:"subject" binding:"required,max=200"`
	Message string `json:"message" binding:"required,max=5000"`
}

func ContactEmailHandler(c *gin.Context) {
	var req ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	apiKey := os.Getenv("RESEND_API_KEY")
	toEmail := os.Getenv("CONTACT_EMAIL")

	if apiKey == "" || toEmail == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server email not configured"})
		return
	}

	// Initialize Resend client
	client := resend.NewClient(apiKey)

	// Build email parameters
	params := &resend.SendEmailRequest{
		From:    "BookLibrary <onboarding@resend.dev>", // Must be verified or use default
		To:      []string{toEmail},
		Subject: fmt.Sprintf("[Contact Form] %s", req.Subject),
		Html: fmt.Sprintf(
			"<p><strong>Name:</strong> %s</p>"+
				"<p><strong>Email:</strong> %s</p>"+
				"<p><strong>Message:</strong><br/>%s</p>",
			req.Name, req.Email, req.Message,
		),
	}

	// Send email
	sent, err := client.Emails.SendWithContext(context.Background(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Resend API error: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Message sent successfully",
		"id":      sent.Id, // optional: Resend email ID for debugging
	})
}
