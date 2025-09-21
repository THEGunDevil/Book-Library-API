package handlers

import (
	"errors"
	"net/http"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/THEGunDevil/GoForBackend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func BorrowBookHandler(c *gin.Context) {
	var req models.CreateBorrowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
	}

	borrowRes, err := service.Borrow(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, borrowRes)
}

func ReturnBookHandler(c *gin.Context) {
	var req models.ReturnBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
	}
	resp, err := service.Return(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)

}

func GetAllBorrowsHandlers(c *gin.Context) {
	borrows, err := db.Q.ListBorrow(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	var response []models.BorrowResponse
	for _, b := range borrows {
		response = append(response, models.BorrowResponse{
			ID:         b.ID.String(),
			UserID:     b.UserID.String(),
			BookID:     b.BookID.String(),
			BorrowedAt: b.BorrowedAt.Time,
			DueDate:    b.DueDate.Time,
			ReturnedAt: &b.ReturnedAt.Time,
		})
	}

	c.JSON(http.StatusOK, response)
}
func GetBorrowsByIDHandler(c *gin.Context) {
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
            ID:         b.ID.String(),
            UserID:     b.UserID.String(),
            BookID:     b.BookID.String(),
            BorrowedAt: b.BorrowedAt.Time,
            DueDate:    b.DueDate.Time,
            ReturnedAt: &b.ReturnedAt.Time,
        })
    }

    c.JSON(http.StatusOK, response)
}

