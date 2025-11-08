package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func CreateReservationHandler(c *gin.Context) {
	// Get the logged-in user ID from context
	requestingUserID, _ := c.Get("userID")
	userUUID := requestingUserID.(uuid.UUID)

	var req struct {
		BookID string `json:"book_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bookUUID, err := uuid.Parse(req.BookID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID format"})
		return
	}

	// Check if book exists
	book, err := db.Q.GetBookByID(c.Request.Context(), pgtype.UUID{Bytes: bookUUID, Valid: true})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Book not found"})
		return
	}

	if book.AvailableCopies.Int32 > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Book is available, please borrow instead"})
		return
	}

	// Create reservation
	r, err := db.Q.CreateReservation(c.Request.Context(), gen.CreateReservationParams{
		UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
		BookID: pgtype.UUID{Bytes: bookUUID, Valid: true},
	})
	if err != nil {
		// Handle unique constraint violation (already reserved)
		if strings.Contains(err.Error(), "unique_user_book") {
			c.JSON(http.StatusConflict, gin.H{"error": "You already have a reservation for this book"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
		return
	}

	// Build response
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

	// ---- Parse pagination query params ----
	pageQuery := c.DefaultQuery("page", "1")
	limitQuery := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageQuery)
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(limitQuery)
	if err != nil || limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var reservations []models.ReservationResponse

	ctx := c.Request.Context()

	if role == "admin" {
		// Use sqlc struct for admin pagination
		params := gen.GetAllReservationsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		}

		adminRes, err := db.Q.GetAllReservations(ctx, params)
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
		// Use sqlc struct for user pagination
		params := gen.GetUserReservationsParams{
			UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
			Limit:  int32(limit),
			Offset: int32(offset),
		}

		userRes, err := db.Q.GetUserReservations(ctx, params)
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

	// Return reservations along with pagination info
	c.JSON(http.StatusOK, gin.H{
		"reservations": reservations,
		"page":         page,
		"limit":        limit,
		"count":        len(reservations),
	})
}

// CreateReservationHandler creates a new book reservation

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
func GetReservationsByBookIDAndUserIDHandler(c *gin.Context) {
	bookIDStr := c.Param("id")      // from /reservations/book/:id
	userIDStr := c.Query("user_id") // from ?user_id=<UUID>

	if bookIDStr == "" || userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing book_id or user_id"})
		return
	}

	bookUUID, err := uuid.Parse(bookIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// ✅ Call a new query that filters by BOTH
	r, err := db.Q.GetReservationsByBookIDAndUserID(
		c.Request.Context(),
		gen.GetReservationsByBookIDAndUserIDParams{
			BookID: pgtype.UUID{Bytes: bookUUID, Valid: true},
			UserID: pgtype.UUID{Bytes: userUUID, Valid: true},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
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
func GetReservationsByReservationID(c *gin.Context) {
	idStr := c.Param("id") // Correct: Param, not Params
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	// Fetch raw reservations from DB
	r, err := db.Q.GetReservationsByReservationID(c.Request.Context(), pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation ID"})
		return
	}
	res := models.ReservationResponse{
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

	c.JSON(http.StatusOK, res)
}
func UpdateBookCopiesToDeleteReservations(c *gin.Context) {
	// 1️⃣ Parse book ID from URL
	bookIDStr := c.Param("id")
	bookID, err := uuid.Parse(bookIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	// 2️⃣ Bind JSON request
	var req struct {
		TotalCopies     *int32 `json:"total_copies"`
		AvailableCopies *int32 `json:"available_copies"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 3️⃣ Prepare params
	params := gen.UpdateBookByIDParams{
		ID:              pgtype.UUID{Bytes: bookID, Valid: true},
		TotalCopies:     pgtype.Int4{Valid: false},
		AvailableCopies: pgtype.Int4{Valid: false},
	}

	if req.TotalCopies != nil {
		params.TotalCopies = pgtype.Int4{Int32: *req.TotalCopies, Valid: true}
	}
	if req.AvailableCopies != nil {
		params.AvailableCopies = pgtype.Int4{Int32: *req.AvailableCopies, Valid: true}
	}

	// 4️⃣ Update book
	updatedBook, err := db.Q.UpdateBookByID(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update book"})
		return
	}

	// 5️⃣ Cleanup reservations if AvailableCopies is now 0
		if err := db.Q.CleanupReservations(c.Request.Context(), pgtype.UUID{Bytes: bookID, Valid: true}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup reservations"})
			return
		}

	// 6️⃣ Return success
	c.JSON(http.StatusOK, gin.H{
		"message":          "Book updated successfully",
		"available_copies": updatedBook.AvailableCopies.Int32,
		"total_copies":     updatedBook.TotalCopies,
	})
}
