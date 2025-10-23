package handlers

import (
	"errors"
	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"log"
	"net/http"
)

// CreateReviewHandler creates a new review
func CreateReviewHandler(c *gin.Context) {
	// Extract userID from context (set by AuthMiddleware)
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userUUID := userIDVal.(uuid.UUID)

	// Bind JSON
	var req struct {
		BookID  string `json:"bookId,omitempty"`
		Rating  int    `json:"rating,omitempty"`
		Comment string `json:"comment,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Validate input
	if req.BookID == "" || req.Comment == "" || req.Rating < 1 || req.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid fields"})
		return
	}

	bookUUID, err := uuid.Parse(req.BookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bookID"})
		return
	}

	// Insert review
	review := gen.CreateReviewParams{
		UserID:  pgtype.UUID{Bytes: userUUID, Valid: true},
		BookID:  pgtype.UUID{Bytes: bookUUID, Valid: true},
		Rating:  pgtype.Int4{Int32: int32(req.Rating), Valid: true},
		Comment: pgtype.Text{String: req.Comment, Valid: true},
	}

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

// UpdateReviewByIDHandler updates an existing review
func UpdateReviewByIDHandler(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userUUID := userIDVal.(uuid.UUID)

	reviewIDStr := c.Param("id")
	reviewID, err := uuid.Parse(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review ID"})
		return
	}

	var req models.UpdateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Verify ownership
	reviews, err := db.Q.GetReviewsByReviewID(c.Request.Context(), pgtype.UUID{Bytes: reviewID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		}
		return
	}

	if len(reviews) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		return
	}

	review := reviews[0] // take the first review
	if review.UserID.Bytes != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only update your own review"})
		return
	}

	// Build update params
	params := gen.UpdateReviewByIDParams{ID: pgtype.UUID{Bytes: reviewID, Valid: true}}
	if req.Rating != nil {
		params.Rating = pgtype.Int4{Int32: int32(*req.Rating), Valid: true}
	}
	if req.Comment != nil {
		params.Comment = pgtype.Text{String: *req.Comment, Valid: true}
	}

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

func GetReviewsByBookIDHandler(c *gin.Context) {
	bookIDStr := c.Param("id")
	bookID, err := uuid.Parse(bookIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	book, err := db.Q.GetBookByID(c.Request.Context(), pgtype.UUID{Bytes: bookID, Valid: true})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dbReviews, err := db.Q.GetReviewsByBookID(c.Request.Context(), pgtype.UUID{Bytes: bookID, Valid: true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var reviews []models.ReviewResponse
	for _, r := range dbReviews {
		user, err := db.Q.GetUserByID(c.Request.Context(), pgtype.UUID{Bytes: r.UserID.Bytes, Valid: true})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		reviews = append(reviews, models.ReviewResponse{
			ID:        r.ID.Bytes,
			UserID:    r.UserID.Bytes,
			BookID:    r.BookID.Bytes,
			BookTitle: book.Title,
			UserName:  user.FirstName + " " + user.LastName,
			Rating:    int(r.Rating.Int32),
			Comment:   r.Comment.String,
			CreatedAt: r.CreatedAt.Time,
			UpdatedAt: r.UpdatedAt.Time,
		})
	}

	c.JSON(http.StatusOK, reviews)
}

func DeleteReviewsByIDHandler(c *gin.Context) {
	reviewParam := c.Param("id")
	reviewId, err := uuid.Parse(reviewParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}
	err = db.Q.DeleteReview(db.Ctx, pgtype.UUID{Bytes: reviewId, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
		} else {
			// Any other DB or server error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}
}
