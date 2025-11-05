package handlers

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
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

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	sendgridAPIKey := os.Getenv("SENDGRID_API_KEY")
	toEmail := os.Getenv("CONTACT_EMAIL")
	verifiedSender := os.Getenv("VERIFIED_SENDER_EMAIL") // Add this env variable

	if sendgridAPIKey == "" || toEmail == "" || verifiedSender == "" {
		log.Println("Missing SendGrid configuration")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server email not configured"})
		return
	}

	from := mail.NewEmail("Book Library Contact Form", verifiedSender)
	to := mail.NewEmail("Admin", toEmail)
	subject := fmt.Sprintf("[Contact Form] %s", html.EscapeString(req.Subject))
	content := fmt.Sprintf(
		"<p><strong>Name:</strong> %s</p>"+
			"<p><strong>Email:</strong> %s</p>"+
			"<p><strong>Message:</strong><br/>%s</p>",
		html.EscapeString(req.Name),
		html.EscapeString(req.Email),
		html.EscapeString(req.Message),
	)
	message := mail.NewSingleEmail(from, subject, to, "", content)
	// Set the reply-to to the user’s email
	message.SetReplyTo(mail.NewEmail(req.Name, req.Email))

	client := sendgrid.NewSendClient(sendgridAPIKey)
	response, err := client.Send(message)
	if err != nil {
		log.Printf("SendGrid error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	if response.StatusCode >= 400 {
		log.Printf("SendGrid returned status %d: %s", response.StatusCode, response.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Message sent successfully"})
}
