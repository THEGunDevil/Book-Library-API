package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetUserHandler(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	user, err := db.Q.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
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

	c.JSON(http.StatusOK, resp)

}
func GetUserByIDHandler(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "userID not found in context"})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid userID type"})
		return
	}

	user, err := db.Q.GetUserByID(c.Request.Context(), pgtype.UUID{Bytes: userID, Valid: true})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	resp := models.UserResponse{
		ID:          user.ID.String(),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber.String,
		Role:        user.Role.String,
		CreatedAt:   user.CreatedAt.Time,
	}

	c.JSON(http.StatusOK, resp)
}
func GetAllUsersHandler(c *gin.Context) {
	users, err := db.Q.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	// Convert users to response format
	var resp []models.UserResponse
	for _, u := range users {
		resp = append(resp, models.UserResponse{
			ID:          u.ID.String(),
			FirstName:   u.FirstName,
			LastName:    u.LastName,
			Email:       u.Email,
			PhoneNumber: u.PhoneNumber.String,
			Role:        u.Role.String,
			CreatedAt:   u.CreatedAt.Time,
		})
	}

	c.JSON(http.StatusOK, resp)
}

func UpdateUserByIDHandler(c *gin.Context) {
	// 1️⃣ Parse UUID from URL
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// 2️⃣ Bind JSON to request struct
	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 3️⃣ Validate non-empty strings for NOT NULL fields
	if req.FirstName != nil && *req.FirstName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "first name cannot be empty"})
		return
	}
	if req.LastName != nil && *req.LastName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "last name cannot be empty"})
		return
	}

	// 4️⃣ Map request to sqlc-generated params
	params := gen.UpdateUserByIDParams{
		ID: pgtype.UUID{Bytes: parsedID, Valid: true},
	}

	if req.FirstName != nil {
		params.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		params.LastName = *req.LastName
	}
	if req.PhoneNumber != nil {
		params.PhoneNumber = pgtype.Text{String: *req.PhoneNumber, Valid: true}
	} else {
		params.PhoneNumber = pgtype.Text{Valid: false} // leave existing value
	}

	// 5️⃣ Execute update
	updatedUser, err := db.Q.UpdateUserByID(c.Request.Context(), params)
	if err != nil {
		log.Printf("UpdateUserByID error: %v", err)
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	// 6️⃣ Return updated user
	c.JSON(http.StatusOK, updatedUser)
}
