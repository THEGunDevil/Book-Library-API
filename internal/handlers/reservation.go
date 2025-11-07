package handlers

import (
	"net/http"
	"strings"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetReservationsHandler gets reservations based on user role
func GetReservationsHandler(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	role, _ := c.Get("role")
	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var reservations []models.ReservationResponse

	if role == "admin" {
		adminRes, err := db.Q.GetAllReservations(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
			return
		}
		for _, r := range adminRes {
			reservations = append(reservations, models.ReservationResponse{
				ID:          r.ID.Bytes,
				BookID:      r.BookID.Bytes,
				UserID:      r.UserID.Bytes,
				Status:      r.Status,
				CreatedAt:   r.CreatedAt.Time,
				NotifiedAt:  r.NotifiedAt.Time,
				FulfilledAt: r.FulfilledAt.Time,
				CancelledAt: r.CancelledAt.Time,
				UserName:    r.UserName,
				UserEmail:   r.Email,
				BookTitle:   r.Title,
				BookAuthor:  r.Author,
				BookImage:   r.ImageUrl,
			})
		}
	} else {
		userRes, err := db.Q.GetUserReservations(c.Request.Context(),
			pgtype.UUID{Bytes: userUUID, Valid: true})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
			return
		}
		for _, r := range userRes {
			reservations = append(reservations, models.ReservationResponse{
				ID:          r.ID.Bytes,
				BookID:      r.BookID.Bytes,
				UserID:      r.UserID.Bytes,
				Status:      r.Status,
				CreatedAt:   r.CreatedAt.Time,
				NotifiedAt:  r.NotifiedAt.Time,
				FulfilledAt: r.FulfilledAt.Time,
				CancelledAt: r.CancelledAt.Time,
				UserName:    r.UserName,
				UserEmail:   r.Email,
				BookTitle:   r.Title,
				BookAuthor:  r.Author,
				BookImage:   r.ImageUrl,
			})
		}
	}

	c.JSON(http.StatusOK, reservations)

}

// CreateReservationHandler creates a new book reservation
func CreateReservationHandler(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		BookID string `json:"book_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the requesting user matches the reservation user (unless admin)
	requestingUserID, _ := c.Get("userID")
	role, _ := c.Get("role")

	if role != "admin" && requestingUserID.(uuid.UUID).String() != req.UserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot create reservation for another user"})
		return
	}

	// Parse UUIDs
	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	bookUUID, err := uuid.Parse(req.BookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID format"})
		return
	}

	// Check if book exists and is unavailable
	book, err := db.Q.GetBookByID(c.Request.Context(),
		pgtype.UUID{Bytes: bookUUID, Valid: true})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	if book.AvailableCopies.Int32 > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Book is available, please borrow instead"})
		return
	}

	// Check for existing active reservation (handled by DB unique constraint, but good to check)
	count, err := db.Q.CheckExistingReservation(c.Request.Context(), gen.CheckExistingReservationParams{
		UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
		BookID: pgtype.UUID{Bytes: bookUUID, Valid: true},
	})
	if err == nil && count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "You already have an active reservation for this book"})
		return
	} // Create reservation
	r, err := db.Q.CreateReservation(c.Request.Context(), gen.CreateReservationParams{
		UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
		BookID: pgtype.UUID{Bytes: bookUUID, Valid: true},
	})

	if err != nil {
		// Check if it's a unique constraint violation
		if strings.Contains(err.Error(), "unique_user_book_pending") {
			c.JSON(http.StatusConflict, gin.H{"error": "You already have a pending reservation for this book"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
		return
	}
	resp := models.ReservationResponse{
		ID:          r.ID.Bytes,
		BookID:      r.BookID.Bytes,
		UserID:      r.UserID.Bytes,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt.Time,
		NotifiedAt:  r.NotifiedAt.Time,
		FulfilledAt: r.FulfilledAt.Time,
		CancelledAt: r.CancelledAt.Time,
	}

	c.JSON(http.StatusCreated, resp)
}

// GetNextReservationHandler gets the next pending reservation for a book (admin only)
func GetNextReservationHandler(c *gin.Context) {
	bookIDStr := c.Param("id")
	bookUUID, err := uuid.Parse(bookIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	r, err := db.Q.GetNextReservationForBook(c.Request.Context(),
		pgtype.UUID{Bytes: bookUUID, Valid: true})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No pending reservations for this book"})
		return
	}
	resp := models.ReservationResponse{
		ID:          r.ID.Bytes,
		BookID:      r.BookID.Bytes,
		UserID:      r.UserID.Bytes,
		Status:      r.Status,
		CreatedAt:   r.CreatedAt.Time,
		NotifiedAt:  r.NotifiedAt.Time,
		FulfilledAt: r.FulfilledAt.Time,
		CancelledAt: r.CancelledAt.Time,
		UserName:    r.UserName,
		UserEmail:   r.Email,
		BookTitle:   r.Title,
		BookAuthor:  r.Author,
		BookImage:   r.ImageUrl,
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateReservationStatusHandler updates a reservation's status (admin only)
func UpdateReservationStatusHandler(c *gin.Context) {
	reservationIDStr := c.Param("id")
	reservationUUID, err := uuid.Parse(reservationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation ID"})
		return
	}

	var req models.UpdateReservationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reservation, err := db.Q.UpdateReservationStatus(c.Request.Context(), gen.UpdateReservationStatusParams{
		ID:     pgtype.UUID{Bytes: reservationUUID, Valid: true},
		Status: req.Status,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reservation status"})
		return
	}

	c.JSON(http.StatusOK, reservation)
}
func GetReservationsByBookIDHandler(c *gin.Context) {
	idStr := c.Param("id") // Correct: Param, not Params
	bookID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	// Fetch raw reservations from DB
	dbReservations, err := db.Q.GetReservationsByBookID(
		c.Request.Context(),
		pgtype.UUID{Bytes: bookID, Valid: true},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
		return
	}

	// Map to response objects
	var reservations []models.ReservationResponse
	for _, r := range dbReservations {
		reservations = append(reservations, models.ReservationResponse{
			ID:          r.ID.Bytes,
			BookID:      r.BookID.Bytes,
			UserID:      r.UserID.Bytes,
			Status:      r.Status,
			CreatedAt:   r.CreatedAt.Time,
			NotifiedAt:  r.NotifiedAt.Time,
			FulfilledAt: r.FulfilledAt.Time,
			CancelledAt: r.CancelledAt.Time,
			UserName:    r.UserName,
			UserEmail:   r.Email,
			BookTitle:   r.Title,
			BookAuthor:  r.Author,
			BookImage:   r.ImageUrl,
		})
	}

	c.JSON(http.StatusOK, reservations)
}

