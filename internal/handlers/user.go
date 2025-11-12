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

	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// 3️⃣ Build response
	var response []models.UserResponse
	for _, user := range users {
		activeBorrowsCount, err := db.Q.CountActiveBorrowsByUserID(
			c.Request.Context(),
			pgtype.UUID{Bytes: user.ID.Bytes, Valid: true},
		)
		if err != nil {
			log.Printf("Failed to count active borrows for user %v: %v", user.ID, err)
			activeBorrowsCount = 0
		}

		allBorrowsCount, err := db.Q.CountBorrowedBooksByUserID(
			c.Request.Context(),
			pgtype.UUID{Bytes: user.ID.Bytes, Valid: true},
		)
		if err != nil {
			log.Printf("Failed to count all borrows for user %v: %v", user.ID, err)
			allBorrowsCount = 0
		}

		response = append(response, models.UserResponse{
			ID:                 user.ID.Bytes,
			FirstName:          user.FirstName,
			LastName:           user.LastName,
			Bio:                user.Bio,
			Email:              user.Email,
			PhoneNumber:        user.PhoneNumber,
			Role:               user.Role.String,
			IsBanned:           user.IsBanned.Bool,
			BanUntil:           &user.BanUntil.Time,
			BanReason:          user.BanReason.String,
			IsPermanentBan:     user.IsPermanentBan.Bool,
			AllBorrowsCount:    int(allBorrowsCount),
			ActiveBorrowsCount: int(activeBorrowsCount),
			CreatedAt:          user.CreatedAt.Time,

			// 👈 optional: add field to struct
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"page":        page,
		"limit":       limit,
		"count":       len(response),
		"total_count": totalCount,
		"total_pages": totalPages,
		"users":       response,
	})
}

// GetUserByIDHandler fetches user by ID (including banned ones)
// GetUserByIDHandler fetches user by ID (including banned ones)
func GetUserByIDHandler(c *gin.Context) {
    // 1️⃣ Determine which user to fetch
    userIDVal, exists := c.Get("userID")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "userID not found in context"})
        return
    }

    userUUID, ok := userIDVal.(uuid.UUID)
    if !ok {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid userID type"})
        return
    }

    // Optional: check if the middleware set banned_user flag
    // bannedUserFlag, _ := c.Get("banned_user")

    // 2️⃣ Fetch user from DB
    user, err := db.Q.GetUserByID(c.Request.Context(), pgtype.UUID{Bytes: userUUID, Valid: true})
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }

    // 3️⃣ Count borrows
    activeBorrowsCount, err := db.Q.CountActiveBorrowsByUserID(c.Request.Context(), pgtype.UUID{Bytes: user.ID.Bytes, Valid: true})
    if err != nil {
        log.Printf("Failed to count active borrows for user %v: %v", user.ID, err)
        activeBorrowsCount = 0
    }

    allBorrowsCount, err := db.Q.CountBorrowedBooksByUserID(c.Request.Context(), pgtype.UUID{Bytes: user.ID.Bytes, Valid: true})
    if err != nil {
        log.Printf("Failed to count all borrows for user %v: %v", user.ID, err)
        allBorrowsCount = 0
    }

    // 4️⃣ Build response
    resp := models.UserResponse{
        ID:                 user.ID.Bytes,
        FirstName:          user.FirstName,
        LastName:           user.LastName,
        Bio:                user.Bio,
        Email:              user.Email,
        PhoneNumber:        user.PhoneNumber,
        Role:               user.Role.String,
        IsBanned:           user.IsBanned.Bool,
        IsPermanentBan:     user.IsPermanentBan.Bool,
        BanReason:          user.BanReason.String,
        BanUntil:           &user.BanUntil.Time,
        AllBorrowsCount:    int(allBorrowsCount),
        ActiveBorrowsCount: int(activeBorrowsCount),
        CreatedAt:          user.CreatedAt.Time,
    }

    log.Printf("👤 Returning user data for user %v (banned: %v)", user.ID, user.IsBanned.Bool)
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
	} else if req.BanUntil != nil {
		banUntil = pgtype.Timestamp{Time: *req.BanUntil, Valid: true}
	} else {
		banUntil = pgtype.Timestamp{Valid: false}
	}

	// Update user ban
	params := gen.UpdateUserBanByUserIDParams{
		ID:             pgtype.UUID{Bytes: parsedID, Valid: true},
		IsBanned:       pgtype.Bool{Bool: req.IsBanned, Valid: true},
		BanReason:      pgtype.Text{String: req.BanReason, Valid: true},
		BanUntil:       banUntil,
		IsPermanentBan: pgtype.Bool{Bool: req.IsPermanentBan, Valid: true},
	}

	updatedUser, err := db.Q.UpdateUserBanByUserID(c.Request.Context(), params)
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
		BanUntil:       banUntilPtr, // properly handle null
		IsPermanentBan: updatedUser.IsPermanentBan.Bool,
	}

	c.JSON(http.StatusOK, resp)
}
