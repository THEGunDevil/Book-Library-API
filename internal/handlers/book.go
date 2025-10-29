package handlers

import (
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

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
	genre := c.PostForm("genre")
	description := c.PostForm("description")

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
		Genre:         genre,
		Description:   description,
	}, imageURL)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bookResp)
}

func GetBooksHandler(c *gin.Context) {
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

    params := gen.ListBooksPaginatedParams{
        Limit:  int32(limit),
        Offset: int32(offset),
    }

    // 1️⃣ Fetch paginated books
    books, err := db.Q.ListBooksPaginated(c.Request.Context(), params)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 2️⃣ Fetch total count of all books
    totalCount, err := db.Q.CountBooks(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    // 3️⃣ Compute total pages
    totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

    // Convert to response model
    var response []models.BookResponse
    for _, book := range books {
        response = append(response, models.BookResponse{
            ID:              book.ID.Bytes,
            Title:           book.Title,
            Author:          book.Author,
            PublishedYear:   book.PublishedYear.Int32,
            Isbn:            book.Isbn.String,
            AvailableCopies: book.AvailableCopies.Int32,
            TotalCopies:     book.TotalCopies,
            Genre:           book.Genre,
            Description:     book.Description,
            CreatedAt:       book.CreatedAt.Time,
            UpdatedAt:       book.UpdatedAt.Time,
            ImageURL:        book.ImageUrl,
        })
    }

    // 4️⃣ Return all pagination info
    c.JSON(http.StatusOK, gin.H{
        "page":         page,
        "limit":        limit,
        "count":        len(response),
        "total_count":  totalCount,
        "total_pages":  totalPages,
        "books":        response,
    })
}

// GetBookByIDHandler fetches a book by its ID
func GetBookByIDHandler(c *gin.Context) {
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	book, err := db.Q.GetBookByID(c.Request.Context(), pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	response := models.BookResponse{
		ID:              book.ID.Bytes,
		Title:           book.Title,
		Author:          book.Author,
		PublishedYear:   book.PublishedYear.Int32,
		Isbn:            book.Isbn.String,
		AvailableCopies: book.AvailableCopies.Int32,
		TotalCopies:     book.TotalCopies,
		Genre:           book.Genre,
		Description:     book.Description,
		CreatedAt:       book.CreatedAt.Time,
		UpdatedAt:       book.UpdatedAt.Time,
		ImageURL:        book.ImageUrl,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteBookHandler deletes a book by ID
func DeleteBookHandler(c *gin.Context) {
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book id"})
		return
	}

	_, err = db.Q.DeleteBookByID(c.Request.Context(), pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "book deleted"})
}

// UpdateBookByIDHandler updates a book by ID
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
	if req.Genre != nil {
		params.Genre = *req.Genre
	}
	if req.Description != nil {
		params.Description = *req.Description
	}

	updatedBook, err := db.Q.UpdateBookByID(c.Request.Context(), params)
	if err != nil {
		log.Printf("UpdateBookByID error: %v", err)
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "book not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		}
		return
	}
		response := models.BookResponse{
		ID:              updatedBook.ID.Bytes,
		Title:           updatedBook.Title,
		Author:          updatedBook.Author,
		PublishedYear:   updatedBook.PublishedYear.Int32,
		Isbn:            updatedBook.Isbn.String,
		AvailableCopies: updatedBook.AvailableCopies.Int32,
		TotalCopies:     updatedBook.TotalCopies,
		Genre:           updatedBook.Genre,
		Description:     updatedBook.Description,
		CreatedAt:       updatedBook.CreatedAt.Time,
		UpdatedAt:       updatedBook.UpdatedAt.Time,
		ImageURL:        updatedBook.ImageUrl,
	}
	c.JSON(http.StatusOK, response)
}

// SearchBooksHandler searches books by title/author/genre

func SearchBooksHandler(c *gin.Context) {
	query := strings.TrimSpace(c.Query("query"))
	genre := strings.TrimSpace(c.Query("genre"))

	// Use empty string if not provided
	searchParam := query
	genreParam := genre

	// Call SQLC-generated query
	books, err := db.Q.SearchBooks(c.Request.Context(), gen.SearchBooksParams{
		Column1: genreParam,
		Column2: searchParam,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Map to response model
	var response []models.BookResponse
	for _, book := range books {
		response = append(response, models.BookResponse{
			ID:              book.ID.Bytes,
			Title:           book.Title,
			Author:          book.Author,
			PublishedYear:   book.PublishedYear.Int32,
			Isbn:            book.Isbn.String,
			AvailableCopies: book.AvailableCopies.Int32,
			TotalCopies:     book.TotalCopies,
			Genre:           book.Genre,
			Description:     book.Description,
			CreatedAt:       book.CreatedAt.Time,
			UpdatedAt:       book.UpdatedAt.Time,
			ImageURL:        book.ImageUrl,
		})
	}

	c.JSON(http.StatusOK, response)
}

func ListGenresHandler(c *gin.Context) {
	genres, err := db.Q.ListGenres(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, genres)
}
func ListBooksByGenreHandler(c *gin.Context) {
	genre := c.Param("genre")
	if genre == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "genre is required"})
		return
	}

	books, err := db.Q.FilterBooksByGenre(c.Request.Context(), genre)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Map to response model
	var response []models.BookResponse
	for _, book := range books {
		response = append(response, models.BookResponse{
			ID:              book.ID.Bytes,
			Title:           book.Title,
			Author:          book.Author,
			PublishedYear:   book.PublishedYear.Int32,
			Isbn:            book.Isbn.String,
			AvailableCopies: book.AvailableCopies.Int32,
			TotalCopies:     book.TotalCopies,
			Genre:           book.Genre,
			Description:     book.Description,
			CreatedAt:       book.CreatedAt.Time,
			UpdatedAt:       book.UpdatedAt.Time,
			ImageURL:        book.ImageUrl,
		})
	}

	c.JSON(http.StatusOK, response)
}
