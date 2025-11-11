package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
)

const (
	borrowedStatus   = "borrowed_at"
	returnedStatus   = "returned_at"
	noReturnedStatus = "not_returned"
)

func ListDataByStatusHandler(c *gin.Context) {
	status := c.Query("status")
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Pagination
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

	if role == "admin" {
		switch status {
		// =====================
		// Case: Reservations
		// =====================
		case "pending", "notified", "fulfilled", "cancelled":

			params := gen.ListReservationPaginatedParams{
				Limit:  int32(limit),
				Offset: int32(offset),
				Status: string(status),
			}
			reservations, err := db.Q.ListReservationPaginated(c.Request.Context(), params)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to fetch reservations"})
				return
			}
			// map and respond
			var reservationResp []models.ReservationResponse
			for _, r := range reservations {
				reservationResp = append(reservationResp, models.ReservationResponse{
					ID:        r.ID.Bytes,
					UserID:    r.UserID.Bytes,
					BookID:    r.BookID.Bytes,
					Status:    r.Status,
					CreatedAt: r.CreatedAt.Time,
					UserName:  r.UserName,
					UserEmail: r.Email,
					BookTitle: r.BookTitle,
				})
			}
			c.JSON(http.StatusOK, reservationResp)

		// =====================
		// Case: Borrowed Books
		// =====================
		case borrowedStatus:
			// Example: use current date to filter borrowed books
			params := gen.ListBorrowPaginatedByBorrowedAtParams{
				Limit:  int32(limit),
				Offset: int32(offset),
			}

			borrows, err := db.Q.ListBorrowPaginatedByBorrowedAt(c.Request.Context(), params)
			if err != nil {
				log.Print("failed to fetch borrowed data", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "failed to fetch borrowed data"})
				return
			}
			var borrowResp []models.BorrowResponse
			for _, b := range borrows {
				borrowResp = append(borrowResp, models.BorrowResponse{
					ID:         b.ID.Bytes,
					UserID:     b.UserID.Bytes,
					UserName:   b.UserName,
					BookID:     b.BookID.Bytes,
					BorrowedAt: b.BorrowedAt.Time,
					DueDate:    b.DueDate.Time,
					ReturnedAt: &b.ReturnedAt.Time,
					BookTitle:  b.BookTitle,
				})
			}
			c.JSON(http.StatusOK, borrowResp)
		case returnedStatus:
			params := gen.ListBorrowPaginatedByReturnedAtParams{
				Limit:  int32(limit),
				Offset: int32(offset),
			}

			borrows, err := db.Q.ListBorrowPaginatedByReturnedAt(c.Request.Context(), params)
			if err != nil {
				log.Print("failed to fetch borrowed data", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "failed to fetch borrowed data"})
				return
			}
			var borrowResp []models.BorrowResponse
			for _, b := range borrows {
				borrowResp = append(borrowResp, models.BorrowResponse{
					ID:         b.ID.Bytes,
					UserID:     b.UserID.Bytes,
					UserName:   b.UserName,
					BookID:     b.BookID.Bytes,
					BorrowedAt: b.BorrowedAt.Time,
					DueDate:    b.DueDate.Time,
					ReturnedAt: &b.ReturnedAt.Time,
					BookTitle:  b.BookTitle,
				})
			}
			c.JSON(http.StatusOK, borrowResp)
		case noReturnedStatus:
			params := gen.ListBorrowPaginatedByNotReturnedAtParams{
				Limit:  int32(limit),
				Offset: int32(offset),
			}

			borrows, err := db.Q.ListBorrowPaginatedByNotReturnedAt(c.Request.Context(), params)
			if err != nil {
				log.Print("failed to fetch borrowed data", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "failed to fetch borrowed data"})
				return
			}
			var borrowResp []models.BorrowResponse
			for _, b := range borrows {
				borrowResp = append(borrowResp, models.BorrowResponse{
					ID:         b.ID.Bytes,
					UserID:     b.UserID.Bytes,
					UserName:   b.UserName,
					BookID:     b.BookID.Bytes,
					BorrowedAt: b.BorrowedAt.Time,
					DueDate:    b.DueDate.Time,
					ReturnedAt: &b.ReturnedAt.Time,
					BookTitle:  b.BookTitle,
				})
			}
			c.JSON(http.StatusOK, borrowResp)
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		}
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
	}
}
