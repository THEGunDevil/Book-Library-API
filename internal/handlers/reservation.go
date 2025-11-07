package handlers

import (
	"net/http"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
)

// CreateReservationHandler creates a new reservation
func CreateReservationHandler(c *gin.Context) {
	var req models.CreateReservationParams
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the UUIDs are valid
	if req.UserID == uuid.Nil || req.BookID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UserID or BookID"})
		return
	}

	reservation, err := db.Q.CreateReservation(c.Request.Context(), gen.CreateReservationParams{
		UserID: pgtype.UUID{Bytes: req.UserID, Valid: true},
		BookID: pgtype.UUID{Bytes: req.BookID, Valid: true},
	})
	if err != nil {
		// Handle unique constraint violation (duplicate pending reservation)
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "You already have a pending reservation for this book"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Reservation created successfully",
		"reservation": reservation,
	})
}

// GetNextReservationHandler fetches the next pending reservation for a book
func GetNextReservationHandler(c *gin.Context) {
	bookID := c.Param("id")
	if bookID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Book ID is required"})
		return
	}

	bookUUID, err := uuid.Parse(bookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	reservation, err := db.Q.GetNextReservationForBook(c.Request.Context(), pgtype.UUID{Bytes: bookUUID, Valid: true})
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			c.JSON(http.StatusOK, gin.H{"message": "No pending reservations for this book"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch next reservation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Next reservation fetched successfully",
		"reservation": reservation,
	})
}

// UpdateReservationStatusHandler updates the status of a reservation
// UpdateReservationStatusHandler updates the status of a reservation (PATCH)
func UpdateReservationStatusHandler(c *gin.Context) {
	reservationID := c.Param("id")
	if reservationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reservation ID is required"})
		return
	}

	var req models.UpdateReservationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resUUID, err := uuid.Parse(reservationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation ID"})
		return
	}

	err = db.Q.UpdateReservationStatus(c.Request.Context(), gen.UpdateReservationStatusParams{
		Status: req.Status,
		ID:     pgtype.UUID{Bytes: resUUID, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reservation status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Reservation status updated successfully",
		"status":  req.Status,
		"id":      reservationID,
	})
}

// GetAllReservationsHandler fetches all reservations for admin or the user's own reservations
func GetAllReservationsHandler(c *gin.Context) {
    // Extract user info from context (assuming middleware.AuthMiddleware set it)
    user, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    appUser, ok := user.(models.User)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context"})
        return
    }

    var reservations []gen.Reservation
    var err error

    // Admin can see all reservations
    if appUser.Role == "admin" {
        reservations, err = db.Q.GetAllReservations(c.Request.Context())
    } else {
        // Regular user can only see their own
        reservations, err = db.Q.GetReservationsByUser(c.Request.Context(),
            pgtype.UUID{Bytes: appUser.ID, Valid: true},
        )
    }

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
        return
    }

    c.JSON(http.StatusOK, reservations)
}
