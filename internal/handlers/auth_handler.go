package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/THEGunDevil/GoForBackend/internal/service"
)

// RegisterHandler creates a new user
func RegisterHandler(c *gin.Context) {
	var req models.User
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Password != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "passwords do not match"})
		return
	}

	if len(req.FirstName) < 3 || len(req.FirstName) > 25 || len(req.LastName) < 3 || len(req.LastName) > 25 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "first and last names must be 3-25 chars"})
		return
	}

	emailRegex := regexp.MustCompile(`^[\w.%+-]+@[\w.-]+\.[a-zA-Z]{2,}$`)
	if len(req.Email) == 0 || len(req.Email) > 255 || !emailRegex.MatchString(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email format"})
		return
	}

	hashed, err := service.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process password"})
		return
	}

	user, err := db.Q.CreateUser(c.Request.Context(), gen.CreateUserParams{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PhoneNumber:  req.PhoneNumber,
		PasswordHash: hashed,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	resp := models.UserResponse{
		ID:          user.ID.Bytes,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		CreatedAt:   user.CreatedAt.Time,
		Role:        user.Role.String,
	}

	c.JSON(http.StatusCreated, resp)
}

// LoginHandler logs in a user and issues tokens
func LoginHandler(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := db.Q.GetUserByEmail(c, body.Email)
	if err != nil || service.CheckPassword(body.Password, user.PasswordHash) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	accessToken, err := service.GenerateAccessToken(user.ID.String(), user.Role.String, user.TokenVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	refreshToken, err := service.GenerateRefreshToken(user.ID.String(), user.TokenVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	secure := true
	if c.Request.Host == "localhost" || c.Request.Host == "127.0.0.1" {
		secure = false
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   3600 * 24 * 7,
		Path:     "/",
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken, "role": user.Role.String})
}

// RefreshHandler issues a new access token
func RefreshHandler(c *gin.Context) {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing refresh token"})
		return
	}

	token, err := service.VerifyToken(cookie, true)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	userIDStr, ok := claims["sub"].(string)
	version, ok2 := claims["token_version"].(float64)
	if !ok || !ok2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token data"})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := db.Q.GetUserByID(c, pgtype.UUID{Bytes: userUUID,Valid: true})
	if err != nil || user.TokenVersion != int32(version) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired or invalid"})
		return
	}

	accessToken, err := service.GenerateAccessToken(userIDStr, user.Role.String, user.TokenVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

// LogoutHandler clears refresh token
func LogoutHandler(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}
