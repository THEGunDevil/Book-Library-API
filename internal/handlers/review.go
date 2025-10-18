package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func CreateReviewHandler(c *gin.Context) {
	userID := c.Query("userID")
	ratingStr := c.Query("rating")
	reviewComment := c.Query("comment")
	bookID := c.Query("bookID")

	// Validate input
	if reviewComment == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment can't be empty"})
		return
	}
	if bookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bookID can't be empty"})
		return
	}
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID can't be empty"})
		return
	}
	if ratingStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rating can't be empty"})
		return
	}

	// Parse and validate UUIDs
	bookUUID, err := uuid.Parse(bookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bookID"})
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userID"})
		return
	}

	// Convert rating to int32
	ratingInt, err := strconv.Atoi(ratingStr)
	if err != nil || ratingInt < 1 || ratingInt > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be a number between 1 and 5"})
		return
	}

	// Create review params
	review := gen.CreateReviewParams{
		UserID:  pgtype.UUID{Bytes: userUUID, Valid: true},
		BookID:  pgtype.UUID{Bytes: bookUUID, Valid: true},
		Rating:  pgtype.Int4{Int32: int32(ratingInt), Valid: true},
		Comment: pgtype.Text{String: reviewComment, Valid: true},
	}

	// Insert into DB
	createdReview, err := db.Q.CreateReview(db.Ctx, review)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create review",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "review created successfully",
		"review":  createdReview,
	})
}

func UpdateReviewByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	userIDStr := c.Query("userID") // 👈 assume userID is passed in query or from middleware

	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID is required"})
		return
	}

	reviewID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req models.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// ✅ 1. Verify ownership
	review, err := db.Q.GetReviewByID(c.Request.Context(), pgtype.UUID{Bytes: reviewID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		}
		return
	}

	if review.UserID.Bytes != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only update your own review"})
		return
	}

	// ✅ 2. Build update params
	params := gen.UpdateReviewByIDParams{
		ID: pgtype.UUID{Bytes: reviewID, Valid: true},
	}

	if req.Rating != nil {
		params.Rating = pgtype.Int4{Int32: int32(*req.Rating), Valid: true}
	}
	if req.Comment != nil {
		params.Comment = pgtype.Text{String: *req.Comment, Valid: true}
	}

	// ✅ 3. Perform update
	updatedReview, err := db.Q.UpdateReviewByID(c.Request.Context(), params)
	if err != nil {
		log.Printf("UpdateReviewByID error: %v", err)
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "review updated successfully",
		"review":  updatedReview,
	})
}

