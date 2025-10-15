package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/THEGunDevil/GoForBackend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateBookHandler handles adding books
func CreateBookHandler(c *gin.Context) {
	// Get form values
	title := c.PostForm("title")
	author := c.PostForm("author")
	publishedYearStr := c.PostForm("published_year")
	isbn := c.PostForm("isbn")
	totalCopiesStr := c.PostForm("total_copies")

	// Validate required fields
	if len(title) == 0 || len(title) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must be 1-255 characters"})
		return
	}
	if len(author) == 0 || len(author) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "author must be 1-100 characters"})
		return
	}

	// Convert numeric fields
	publishedYear, err := strconv.Atoi(publishedYearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "published year must be a number"})
		return
	}

	totalCopies, err := strconv.Atoi(totalCopiesStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "total copies must be a number"})
		return
	}

	// Handle file upload
	var imageURL string
	file, err := c.FormFile("image")
	if err == nil {
		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open image"})
			return
		}
		defer f.Close()

		imageURL, err = service.UploadImageToCloudinary(f, file.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "image upload failed"})
			return
		}
	}

	// Call the service
	bookResp, err := service.AddBook(models.CreateBookRequest{
		Title:         title,
		Author:        author,
		PublishedYear: publishedYear,
		Isbn:          isbn,
		TotalCopies:   totalCopies,
	}, imageURL)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bookResp)
}

// GetBookHandler example: fetch all books
func GetBooksHandler(c *gin.Context) {
	books, err := db.Q.ListBooks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}
	// Prepare response
	var response []models.BookResponse
	for _, book := range books {
		response = append(response, models.BookResponse{
			ID:              book.ID.String(),
			Title:           book.Title,
			Author:          book.Author,
			PublishedYear:   book.PublishedYear.Int32,
			Isbn:            book.Isbn.String,
			AvailableCopies: book.AvailableCopies.Int32,
			TotalCopies:     book.TotalCopies,
			CreatedAt:       book.CreatedAt.Time,
			UpdatedAt:       book.UpdatedAt.Time,
			ImageURL:        book.ImageUrl,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetBookByIDHandler example: fetch a book with ID
func GetBookByIDHandler(c *gin.Context) {
	idStr := c.Param("id")

	parsedID, err := uuid.Parse(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
	}

	book, err := db.Q.GetBookByID(c.Request.Context(), pgtype.UUID{Bytes: parsedID,
		Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		} else {
			// Any other DB or server error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	// Prepare response

	response := models.BookResponse{
		ID:              book.ID.String(),
		Title:           book.Title,
		Author:          book.Author,
		PublishedYear:   book.PublishedYear.Int32,
		Isbn:            book.Isbn.String,
		AvailableCopies: book.AvailableCopies.Int32,
		TotalCopies:     book.TotalCopies,
		CreatedAt:       book.CreatedAt.Time,
		UpdatedAt:       book.UpdatedAt.Time,
	}

	c.JSON(http.StatusOK, response)
}

func DeleteBookHandler(c *gin.Context) {

	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
	}
	_, err = db.Q.DeleteBookByID(c.Request.Context(), pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		} else {
			// Any other DB or server error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}
}

func UpdateBookByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	var req models.UpdateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	params := gen.UpdateBookByIDParams{
		ID: pgtype.UUID{Bytes: parsedID, Valid: true},
	}
	// Only assign if the client sent a value
	if req.Title != nil {
		if len(*req.Title) == 0 || len(*req.Title) > 255 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "title must be 1-255 characters"})
			return
		}
		params.Title = *req.Title
	}

	if req.Author != nil {
		if len(*req.Author) == 0 || len(*req.Author) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "author must be 1-100 characters"})
			return
		}
		params.Author = *req.Author
	}

	if req.TotalCopies != nil {
		params.TotalCopies = *req.TotalCopies
	}
	if req.PublishedYear != nil {
		params.PublishedYear = pgtype.Int4{Int32: *req.PublishedYear, Valid: true}
	}
	if req.Isbn != nil {
		params.Isbn = pgtype.Text{String: *req.Isbn, Valid: true}
	}
	if req.AvailableCopies != nil {
		params.AvailableCopies = pgtype.Int4{Int32: *req.AvailableCopies, Valid: true}
	}

	updatedBook, err := db.Q.UpdateBookByID(c.Request.Context(), params)
	if err != nil {
		log.Printf("UpdateBookByID error: %v", err) // 👈 add this
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		} else {
			// Any other DB or server error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	c.JSON(http.StatusOK, updatedBook)
}
