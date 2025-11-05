package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	gomail "gopkg.in/gomail.v2"
)

// ContactRequest defines the expected JSON structure from the frontend
type ContactRequest struct {
	Name    string `json:"name" binding:"required,min=2,max=50"`
	Email   string `json:"email" binding:"required,email"`
	Message string `json:"message" binding:"required"`
}

// ContactHandler handles the contact form submission and sends email via Gmail SMTP
func ContactHandler(c *gin.Context) {
	var req ContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	from := os.Getenv("GMAIL_USER")
	password := os.Getenv("GMAIL_PASSWORD")

	// Create a polished subject line
	subject := fmt.Sprintf("New Contact Form Message from %s <%s>", req.Name, req.Email)

	// Create an HTML body
	body := fmt.Sprintf(`
		<h2>New Contact Form Submission</h2>
		<p><strong>Name:</strong> %s</p>
		<p><strong>Email:</strong> %s</p>
		<p><strong>Message:</strong><br/>%s</p>
	`, req.Name, req.Email, req.Message)

	// Build the email
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", from) // send to yourself
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, from, password)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to send email: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email sent successfully!"})
}

func main() {
	r := gin.Default()

	// Route for contact form
	r.POST("/contact/send", ContactHandler)

	// Start server
	r.Run() // listens on :8080 by default
}
