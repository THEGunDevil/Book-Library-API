package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	gomail "gopkg.in/gomail.v2"
)

// SendEmail sends an email using Gmail SMTP, including sender's name
func SendEmail(name, to, subject, body string) error {
	from := os.Getenv("GMAIL_USER")
	password := os.Getenv("GMAIL_PASSWORD")
	
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", name, from))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	
	// Use port 465 with SSL instead of 587 with TLS
	d := gomail.NewDialer("smtp.gmail.com", 465, from, password)
	d.SSL = true  // Enable SSL
	
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	
	return nil
}

// ContactRequest represents the incoming contact form request
type ContactRequest struct {
	Name    string `json:"name" binding:"required,min=2,max=50"`
	Email   string `json:"email" binding:"required,email"`
	Subject string `json:"subject" binding:"required,min=2,max=100"`
	Message string `json:"message" binding:"required"`
}

// ContactHandler handles contact form submissions
func ContactHandler(c *gin.Context) {
	var req ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log environment variables (remove password logging in production!)
	fmt.Printf("GMAIL_USER: %s\n", os.Getenv("GMAIL_USER"))
	fmt.Printf("GMAIL_PASSWORD exists: %v\n", os.Getenv("GMAIL_PASSWORD") != "")

	body := fmt.Sprintf("From: %s <%s>\n\n%s", req.Name, req.Email, req.Message)
	recipient := "himelsd117@gmail.com"
	
	if err := SendEmail(req.Name, recipient, req.Subject, body); err != nil {
		// Log the actual error
		fmt.Printf("Email error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to send email",
			"details": err.Error(), // Include in development only
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully!"})
}
