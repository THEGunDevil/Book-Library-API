package handlers

import (
	"errors"
	// "fmt"
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

	// Get user info
	user, err := db.Q.GetUserByID(ctx, pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	// Get reviews by this user
	dbReviews, err := db.Q.GetReviewsByUserID(ctx, pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get reviews"})
		return
	}

	// Get borrows by this user
	dbBorrows, err := db.Q.ListBorrowByUserID(ctx, pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get borrows"})
		return
	}

	// Prepare user response
	userResp := models.UserResponse{
		ID:          user.ID.Bytes,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Bio:         user.Bio,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber.String,
		CreatedAt:   user.CreatedAt.Time,
		Role:        user.Role.String,
	}

	// Map borrows
	var borrowResponses []models.BorrowResponse
	for _, b := range dbBorrows {
		borrowResponses = append(borrowResponses, models.BorrowResponse{
			ID:         b.ID.Bytes,
			UserID:     b.UserID.Bytes,
			BookID:     b.BookID.Bytes,
			BookTitle:  b.Title,
			BorrowedAt: b.BorrowedAt.Time,
			DueDate:    b.DueDate.Time,
			ReturnedAt: func(t pgtype.Timestamp) *time.Time {
				if t.Valid {
					return &t.Time
				}
				return nil
			}(b.ReturnedAt),
		})
	}

	// --- Debug: print what we actually got ---
	// fmt.Printf("\nDEBUG: userID=%s, borrows=%d\n", parsedID, len(borrowResponses))
	// for i, b := range borrowResponses {
	// 	fmt.Printf("DEBUG: Borrow[%d] => %+v\n", i, b)
	// }

	// Map reviews
	var reviewResponses []models.ReviewResponse
	for _, r := range dbReviews {
		// Fetch book info for this review
		book, err := db.Q.GetBookByID(ctx, pgtype.UUID{Bytes: r.BookID.Bytes, Valid: true})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get book for review"})
			return
		}

		reviewResponses = append(reviewResponses, models.ReviewResponse{
			ID:        r.ID.Bytes,
			UserID:    r.UserID.Bytes,
			BookID:    r.BookID.Bytes,
			BookTitle: book.Title,
			UserName:  user.FirstName + " " + user.LastName, // profile user
			Rating:    int(r.Rating.Int32),
			Comment:   r.Comment.String,
			CreatedAt: r.CreatedAt.Time,
			UpdatedAt: r.UpdatedAt.Time,
		})
	}

	// Build profile object
	profile := models.Profile{
		UserName: user.FirstName + " " + user.LastName,
		User:     []models.UserResponse{userResp},
		Reviews:  reviewResponses,
		Borrows:  borrowResponses,
	}

	c.JSON(http.StatusOK, profile)
}
