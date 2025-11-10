package handlers

import (
	"context"
	// "encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

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
	var req models.CreateBookRequest

	// Try binding JSON first
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[DEBUG] JSON bind failed: %v", err)

		// Fallback to form-data / multipart
		if err := c.ShouldBind(&req); err != nil {
			log.Printf("[DEBUG] Form-data bind failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
	}

	// Validate string lengths
	if len(req.Title) == 0 || len(req.Title) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must be 1-255 characters"})
		return
	}
	if len(req.Author) == 0 || len(req.Author) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "author must be 1-100 characters"})
		return
	}
	if len(req.Genre) == 0 || len(req.Genre) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "genre must be 1-100 characters"})
		return
	}
	if len(req.Description) == 0 || len(req.Description) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description must be 1-255 characters"})
		return
	}

	// Handle file upload if provided
	var imageURL string
	if req.Image != nil {
		f, err := req.Image.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open image"})
			return
		}
		defer f.Close()

		imageURL, err = service.UploadImageToCloudinary(f, req.Image.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "image upload failed"})
			return
		}
	}

	// Call the service to add book
	bookResp, err := service.AddBook(req, imageURL)
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
		"page":        page,
		"limit":       limit,
		"count":       len(response),
		"total_count": totalCount,
		"total_pages": totalPages,
		"books":       response,
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
	log.Printf("📝 [DEBUG] UpdateBookByIDHandler called with ID param: %s", idStr)

	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("❌ [DEBUG] Invalid UUID param: %s", idStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid book ID"})
		return
	}

	var req models.UpdateBookRequest
	if err := c.ShouldBind(&req); err != nil {
		log.Printf("❌ [DEBUG] Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log.Printf("📘 [DEBUG] Update request: %+v", req)

	params := gen.UpdateBookByIDParams{
		ID: pgtype.UUID{Bytes: parsedID, Valid: true},
	}

	setText := func(reqVal *string) pgtype.Text {
		if reqVal != nil && *reqVal != "" {
			return pgtype.Text{String: *reqVal, Valid: true}
		}
		return pgtype.Text{Valid: false}
	}

	setInt := func(reqVal *int32) pgtype.Int4 {
		if reqVal != nil {
			return pgtype.Int4{Int32: *reqVal, Valid: true}
		}
		return pgtype.Int4{Valid: false}
	}

	// Assign request values
	params.Title = setText(req.Title)
	params.Author = setText(req.Author)
	params.Genre = setText(req.Genre)
	params.Description = setText(req.Description)
	params.Isbn = setText(req.Isbn)
	params.PublishedYear = setInt(req.PublishedYear)
	params.TotalCopies = setInt(req.TotalCopies)
	params.AvailableCopies = setInt(req.AvailableCopies)

	log.Printf("🧩 [DEBUG] UpdateBookByIDParams ready: %+v", params)

	// Upload image if exists
	if req.Image != nil {
		f, err := req.Image.Open()
		if err == nil {
			defer f.Close()
			url, err := service.UploadImageToCloudinary(f, req.Image.Filename)
			if err == nil {
				params.ImageUrl = pgtype.Text{String: url, Valid: true}
				log.Printf("🖼️ [DEBUG] Image uploaded successfully: %s", url)
			} else {
				log.Printf("❌ [DEBUG] Image upload failed: %v", err)
				params.ImageUrl = pgtype.Text{Valid: false}
			}
		} else {
			log.Printf("❌ [DEBUG] Failed to open image file: %v", err)
			params.ImageUrl = pgtype.Text{Valid: false}
		}
	} else {
		log.Printf("⚠️ [DEBUG] No image provided in update request.")
		params.ImageUrl = pgtype.Text{Valid: false}
	}

	updatedBook, err := db.Q.UpdateBookByID(c.Request.Context(), params)
	if err != nil {
		log.Printf("❌ [DEBUG] DB update failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	log.Printf("✅ [DEBUG] Book updated successfully: %s | AvailableCopies=%d",
		updatedBook.Title, updatedBook.AvailableCopies.Int32)

	updatedBookID, err := uuid.FromBytes(updatedBook.ID.Bytes[:])
	if err != nil {
		log.Printf("⚠️ [DEBUG] Failed to parse book UUID: %v", err)
		updatedBookID = uuid.Nil
	}

	// Notify reservations
	if updatedBook.AvailableCopies.Valid && updatedBook.AvailableCopies.Int32 > 0 {
		reservations, err := db.Q.GetReservationsByBookID(c.Request.Context(), updatedBook.ID)
		if err == nil && len(reservations) > 0 {
			var wg sync.WaitGroup

			for _, r := range reservations {
				wg.Add(1)
				userID := r.UserID
				go func(userID pgtype.UUID) {
					defer wg.Done()
					ctx := context.Background()

					userUUID, _ := uuid.FromBytes(userID.Bytes[:])

					err := service.NotificationService(ctx, models.SendNotificationRequest{
						UserID:            userUUID,
						ObjectID:          &updatedBookID,
						ObjectTitle:       updatedBook.Title,
						Type:              "BOOK_AVAILABLE",
						NotificationTitle: "Your reserved book is now available!",
						Message:           fmt.Sprintf("The book '%s' you reserved is now available.", updatedBook.Title),
					})

					if err != nil {
						log.Printf("❌ Failed to send notification to user %s: %v", userUUID, err)
						return
					}

					// Update reservation status to 'notified'
					_, err = db.Q.UpdateReservationStatus(ctx, gen.UpdateReservationStatusParams{
						ID:     r.ID,
						Status: "notified",
					})
					if err != nil {
						log.Printf("❌ Failed to update reservation %s to 'notified': %v", r.ID, err)
						return
					}

					// Then mark as 'fulfilled' (if applicable)
					_, err = db.Q.UpdateReservationStatus(ctx, gen.UpdateReservationStatusParams{
						ID:     r.ID,
						Status: "fulfilled",
					})
					if err != nil {
						log.Printf("❌ Failed to mark reservation %s as 'fulfilled': %v", r.ID, err)
					}
				}(userID)
			}

			wg.Wait()
		}
	}

	log.Printf("✅ [DEBUG] Book update flow complete: ID=%v", updatedBookID)

	c.JSON(http.StatusOK, models.BookResponse{
		ID:              updatedBook.ID.Bytes,
		Title:           updatedBook.Title,
		Author:          updatedBook.Author,
		PublishedYear:   updatedBook.PublishedYear.Int32,
		Isbn:            updatedBook.Isbn.String,
		AvailableCopies: updatedBook.AvailableCopies.Int32,
		TotalCopies:     updatedBook.TotalCopies,
		Genre:           updatedBook.Genre,
		Description:     updatedBook.Description,
		ImageURL:        updatedBook.ImageUrl,
		CreatedAt:       updatedBook.CreatedAt.Time,
		UpdatedAt:       updatedBook.UpdatedAt.Time,
	})
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
