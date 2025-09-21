package handlers

import (
	"net/http"
	"regexp"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/THEGunDevil/GoForBackend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func RegisterHandler(c *gin.Context) {
	var req models.User

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}
	if len(req.FirstName) < 3 || len(req.FirstName) > 25 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "first name must be 3-50 characters"})
		return
	}
	if len(req.LastName) < 3 || len(req.LastName) > 25 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "last name must be 3-50 characters"})
		return
	}
	emailRegex := `^[\w._%+-]+@[\w.-]+\.[a-zA-Z]{2,}$`
	if len(req.Email) > 255 || !regexp.MustCompile(emailRegex).MatchString(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email"})
		return
	}

	hashed, err := service.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user, err := db.Q.CreateUser(c, gen.CreateUserParams{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PhoneNumber:  pgtype.Text{String: req.PhoneNumber, Valid: true},
		PasswordHash: hashed,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := models.UserResponse{
		ID:          user.ID.String(),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber.String,
		CreatedAt:   user.CreatedAt.Time,
	}

	c.JSON(http.StatusCreated, resp)
}

func LoginHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := db.Q.GetUserByEmail(c, req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := service.CheckPassword(req.Password, user.PasswordHash); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, _ := service.GenerateJWT(user.ID.String())

	c.JSON(http.StatusOK, gin.H{"token": token})
}
