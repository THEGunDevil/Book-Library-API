package handlers

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/THEGunDevil/GoForBackend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func BorrowBookHandler(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userUUID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	var req models.CreateBorrowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	borrowRes, err := service.Borrow(userUUID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, borrowRes)
}

func ReturnBookHandler(c *gin.Context) {
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid borrow ID"})
		return
	}
	var req models.ReturnBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := service.Return(parsedID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
func GetAllBorrowsHandlers(c *gin.Context) {
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

	params := gen.ListBorrowPaginatedParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	// 1️⃣ Fetch paginated borrows
	borrows, err := db.Q.ListBorrowPaginated(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2️⃣ Total count
	totalCount, err := db.Q.CountBorrows(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3️⃣ Total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

	// 4️⃣ Build response
	var response []models.BorrowResponse
	for _, b := range borrows {
		var returnedAt *time.Time
		if b.ReturnedAt.Valid {
			returnedAt = &b.ReturnedAt.Time
		}
		user, err := db.Q.GetUserByID(c.Request.Context(), b.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response = append(response, models.BorrowResponse{
			ID:         b.ID.Bytes,
			UserID:     b.UserID.Bytes,
			UserName:   user.FirstName + " " + user.LastName,
			BookID:     b.BookID.Bytes,
			BookTitle:  b.BookTitle,
			BorrowedAt: b.BorrowedAt.Time,
			DueDate:    b.DueDate.Time,
			ReturnedAt: returnedAt,
		})
	}

	// 5️⃣ Return paginated response
	c.JSON(http.StatusOK, gin.H{
		"page":        page,
		"limit":       limit,
		"count":       len(response),
		"total_count": totalCount,
		"total_pages": totalPages,
		"borrows":     response,
	})
}
func GetBorrowsByUserIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	borrows, err := db.Q.ListBorrowByUserID(
		c.Request.Context(),
		pgtype.UUID{Bytes: parsedID, Valid: true},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "borrows not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	var response []models.BorrowResponse
	for _, b := range borrows {

		response = append(response, models.BorrowResponse{
			ID:         b.ID.Bytes,
			UserID:     b.UserID.Bytes,
			BookID:     b.BookID.Bytes,
			BorrowedAt: b.BorrowedAt.Time,
			DueDate:    b.DueDate.Time,
			ReturnedAt: &b.ReturnedAt.Time,
		})
	}

	c.JSON(http.StatusOK, response)
}
func GetBorrowsByBookIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}
	borrows, err := db.Q.ListBorrowByBookID(
		c.Request.Context(),
		pgtype.UUID{Bytes: parsedID, Valid: true},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "borrows not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	var response []models.BorrowResponse
	for _, b := range borrows {

		response = append(response, models.BorrowResponse{
			ID:         b.ID.Bytes,
			UserID:     b.UserID.Bytes,
			BookID:     b.BookID.Bytes,
			BorrowedAt: b.BorrowedAt.Time,
			DueDate:    b.DueDate.Time,
			ReturnedAt: &b.ReturnedAt.Time,
		})
	}

	c.JSON(http.StatusOK, response)
}
func GetBorrowByBookAndUserIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "userID not found in context"})
		return
	}

	// userIDVal is interface{}, convert to string first
	userIDStr, ok := userIDVal.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID is not a string"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	params := gen.FilterBorrowByUserAndBookIDParams{
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
		BookID: pgtype.UUID{Bytes: parsedID, Valid: true},
	}
	b, err := db.Q.FilterBorrowByUserAndBookID(
		c.Request.Context(),
		params,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "borrows not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	var response []models.BorrowResponse
		response = append(response, models.BorrowResponse{
			ID:         b.ID.Bytes,
			UserID:     b.UserID.Bytes,
			BookID:     b.BookID.Bytes,
			BorrowedAt: b.BorrowedAt.Time,
			DueDate:    b.DueDate.Time,
			ReturnedAt: &b.ReturnedAt.Time,
		})
	

	c.JSON(http.StatusOK, response)
}
