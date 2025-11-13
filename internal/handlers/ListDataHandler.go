package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)
type ListReservationPaginatedParams struct {
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
	Status string `json:"status"` // ← Should be string, not []string
}

func timestampToPtr(ts pgtype.Timestamptz) *time.Time {
	if ts.Valid {
		return &ts.Time
	}
	return nil
}
func ListDataByStatusHandler(c *gin.Context) {
	status := strings.ToLower(strings.TrimSpace(c.Query("status"))) // ✅ normalize
	log.Printf("📥 Received status: '%s'", status)                   // ✅ debug log

	// role, exists := c.Get("role")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
	// 	return
	// }

	// roleStr, ok := role.(string)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid role type"})
	// 	return
	// }
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

	// if roleStr == "admin" {
		switch status {
		// =====================
		// Case: Reservations
		// =====================
		case "pending", "notified", "fulfilled", "cancelled":
			params := gen.ListReservationPaginatedByStatusesParams{
				Limit:   int32(limit),
				Offset:  int32(offset),
				Column3: []string{status},
			}

			reservations, err := db.Q.ListReservationPaginatedByStatuses(c.Request.Context(), params)
			if err != nil {
				log.Printf("❌ Failed to fetch reservations: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
				return
			}

			var reservationResp []models.ReservationListResponse // ← Use new model
			for _, r := range reservations {
				reservationResp = append(reservationResp, models.ReservationListResponse{
					ID:          r.ID.Bytes,
					UserID:      r.UserID.Bytes,
					BookID:      r.BookID.Bytes,
					Status:      r.Status,
					CreatedAt:   r.CreatedAt.Time,
					NotifiedAt:  timestampToPtr(r.NotifiedAt),
					FulfilledAt: timestampToPtr(r.FulfilledAt),
					CancelledAt: timestampToPtr(r.CancelledAt),
					UserName:    r.UserName,
					UserEmail:   r.Email,
					BookTitle:   r.BookTitle,
					BookAuthor:  r.Author,
					BookImage:   r.ImageUrl,
				})
			}

			c.JSON(http.StatusOK, gin.H{
				"reservations": reservationResp,
				"page":         page,
				"limit":        limit,
				"count":        len(reservationResp),
			})

		// =====================
		// Case: Borrowed Books
		// =====================
		case "borrowed_at":
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
			c.JSON(http.StatusOK, gin.H{
				"borrows": borrowResp,
				"page":         page,
				"limit":        limit,
				"count":        len(borrowResp),
			})
		case "returned_at":
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
			c.JSON(http.StatusOK, gin.H{
				"borrows": borrowResp,
				"page":         page,
				"limit":        limit,
				"count":        len(borrowResp),
			})
		case "not_returned":
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
			c.JSON(http.StatusOK, gin.H{
				"borrows": borrowResp,
				"page":         page,
				"limit":        limit,
				"count":        len(borrowResp),
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		}
	// } else {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
	// }
}
