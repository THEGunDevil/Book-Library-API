package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func GetProfileData(c *gin.Context) {
	idStr := c.Param("id")

	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx := c.Request.Context()

	// ✅ Get user info
	user, err := db.Q.GetUserByID(ctx, pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	// ✅ Get reviews
	dbReviews, err := db.Q.GetReviewsByUserID(ctx, pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get reviews"})
		return
	}

	// ✅ Get borrows
	dbBorrows, err := db.Q.ListBorrowByUserID(ctx, pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get borrows"})
		return
	}

	// ✅ Prepare user response
	userResp := models.UserResponse{
		ID:          user.ID.String(),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber.String,
		CreatedAt:   user.CreatedAt.Time,
	}

	// ✅ Map borrows
	var borrowResponses []models.BorrowResponse
	for _, b := range dbBorrows {
		var returnedAt *time.Time
		if b.ReturnedAt.Valid {
			returnedAt = &b.ReturnedAt.Time
		}

		borrowResponses = append(borrowResponses, models.BorrowResponse{
			ID:         b.ID.Bytes,
			UserID:     b.UserID.Bytes,
			BookID:     b.BookID.Bytes,
			BorrowedAt: b.BorrowedAt.Time,
			DueDate:    b.DueDate.Time,
			ReturnedAt: returnedAt,
		})
	}

	// ✅ Map reviews
	var reviewResponses []models.ReviewResponse
	for _, r := range dbReviews {
		reviewResponses = append(reviewResponses, models.ReviewResponse{
			ID:        r.ID.Bytes,
			UserID:    r.UserID.Bytes,
			BookID:    r.BookID.Bytes,
			Rating:    int(r.Rating.Int32),
			Comment:   r.Comment.String,
			CreatedAt: r.CreatedAt.Time,
		})
	}

	// ✅ Build profile object
	profile := models.Profile{
		FullName: user.FirstName + " " + user.LastName,
		User:     []models.UserResponse{userResp},
		Reviews:  reviewResponses,
		Borrows:  borrowResponses,
	}


	c.JSON(http.StatusOK, profile)
}
