package handlers

import (
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetUserHandler fetches user by email
func GetUsersHandler(c *gin.Context) {
	page := 1
	limit := 10

	// Read query params
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := (page - 1) * limit

	params := gen.ListUsersPaginatedParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	// 1️⃣ Fetch paginated users
	users, err := db.Q.ListUsersPaginated(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2️⃣ Count total users
	totalCount, err := db.Q.CountUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3️⃣ Compute total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// 4️⃣ Build response
	var response []models.UserResponse
	for _, user := range users {
		response = append(response, models.UserResponse{
			ID:          user.ID.Bytes,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Bio:         user.Bio,
			Email:       user.Email,
			PhoneNumber: user.PhoneNumber,
			Role:        user.Role.String,
			CreatedAt:   user.CreatedAt.Time,
		})
	}

	// 5️⃣ Return paginated data
	c.JSON(http.StatusOK, gin.H{
		"page":        page,
		"limit":       limit,
		"count":       len(response),
		"total_count": totalCount,
		"total_pages": totalPages,
		"users":       response,
	})
}


// GetUserByIDHandler fetches user by ID
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
		ID:          user.ID.Bytes,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Bio:         user.Bio, // added bio
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        user.Role.String,
		CreatedAt:   user.CreatedAt.Time,
	}

	c.JSON(http.StatusOK, resp)
}

// GetAllUsersHandler fetches all users
func GetAllUsersHandler(c *gin.Context) {
	users, err := db.Q.GetAllUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}

	var resp []models.UserResponse
	for _, u := range users {
		resp = append(resp, models.UserResponse{
		ID:            u.ID.Bytes,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Bio:           u.Bio,
		Email:         u.Email,
		PhoneNumber:   u.PhoneNumber,
		Role:          u.Role.String,
		IsBanned:      u.IsBanned.Bool,
		BanUntil:      &u.BanUntil.Time,
		BanReason:     u.BanReason.String,
		IsPermanentBan:u.IsPermanentBan.Bool,
		CreatedAt:     u.CreatedAt.Time,
		})
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateUserByIDHandler updates user by ID
func UpdateUserByIDHandler(c *gin.Context) {
	// Parse UUID
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.FirstName != nil && *req.FirstName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "first name cannot be empty"})
		return
	}
	if req.LastName != nil && *req.LastName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "last name cannot be empty"})
		return
	}

	params := gen.UpdateUserByIDParams{
		ID: pgtype.UUID{Bytes: parsedID, Valid: true},
	}
	if req.FirstName != nil {
		params.FirstName = pgtype.Text{String: *req.FirstName, Valid: true}
	} else {
		params.FirstName = pgtype.Text{Valid: false}
	}

	if req.LastName != nil {
		params.LastName = pgtype.Text{String: *req.LastName, Valid: true}
	} else {
		params.LastName = pgtype.Text{Valid: false}
	}

	if req.PhoneNumber != nil {
		params.PhoneNumber = pgtype.Text{String: *req.PhoneNumber, Valid: true}
	} else {
		params.PhoneNumber = pgtype.Text{Valid: false}
	}

	if req.Bio != nil {
		params.Bio = pgtype.Text{String: *req.Bio, Valid: true}
	} else {
		params.Bio = pgtype.Text{Valid: false}
	}

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

	resp := models.UserResponse{
		ID:             updatedUser.ID.Bytes,
		FirstName:      updatedUser.FirstName,
		LastName:       updatedUser.LastName,
		Bio:            updatedUser.Bio,
		Email:          updatedUser.Email,
		PhoneNumber:    updatedUser.PhoneNumber,
		Role:           updatedUser.Role.String,
		IsBanned:       updatedUser.IsBanned.Bool,
		BanUntil:       &updatedUser.BanUntil.Time,
		BanReason:      updatedUser.BanReason.String,
		IsPermanentBan: updatedUser.IsPermanentBan.Bool,
		CreatedAt:      updatedUser.CreatedAt.Time,
	}

	c.JSON(http.StatusOK, resp)
}

func BanUserByIDHandler(c *gin.Context) {
	// Parse UUID
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Bind request
	var req models.BanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle BanUntil
	var banUntil pgtype.Timestamp
	if req.IsPermanentBan {
		// Permanent ban has no expiration
		banUntil = pgtype.Timestamp{Valid: false}
	} else if req.BanUntil != "" {
		t, err := time.Parse(time.RFC3339, req.BanUntil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format for ban_until (expected RFC3339)"})
			return
		}
		banUntil = pgtype.Timestamp{Time: t, Valid: true}
	} else {
		banUntil = pgtype.Timestamp{Valid: false}
	}

	// Update user ban
	params := gen.UpdateUserBanByIDParams{
		ID:             pgtype.UUID{Bytes: parsedID, Valid: true},
		IsBanned:       pgtype.Bool{Bool: req.IsBanned, Valid: true},
		BanReason:      pgtype.Text{String: req.BanReason, Valid: true},
		BanUntil:       banUntil,
		IsPermanentBan: pgtype.Bool{Bool: req.IsPermanentBan, Valid: true},
	}

	updatedUser, err := db.Q.UpdateUserBanByID(c.Request.Context(), params)
	if err != nil {
		log.Printf("BanUser error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user ban status"})
		return
	}

	// Prepare response
	var banUntilPtr *time.Time
	if updatedUser.BanUntil.Valid {
		banUntilPtr = &updatedUser.BanUntil.Time
	} else {
		banUntilPtr = nil
	}

	resp := models.UserResponse{
		ID:             updatedUser.ID.Bytes,
		FirstName:      updatedUser.FirstName,
		LastName:       updatedUser.LastName,
		Bio:            updatedUser.Bio,
		Email:          updatedUser.Email,
		PhoneNumber:    updatedUser.PhoneNumber,
		CreatedAt:      updatedUser.CreatedAt.Time,
		Role:           updatedUser.Role.String,
		TokenVersion:   int(updatedUser.TokenVersion),
		IsBanned:       updatedUser.IsBanned.Bool,
		BanReason:      updatedUser.BanReason.String,
		BanUntil:       banUntilPtr,            // properly handle null
		IsPermanentBan: updatedUser.IsPermanentBan.Bool,
	}

	c.JSON(http.StatusOK, resp)
}
