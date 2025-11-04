package handlers

import (
	"encoding/csv"
	"fmt"
	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/gin-gonic/gin"
	"github.com/phpdave11/gofpdf"
	"net/http"
	"strconv"
)

func DownloadBooksHandler(c *gin.Context) {
	format := c.Query("format") // "csv" or "pdf"
	page := 1
	limit := 10

	// Parse page & limit
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

	books, err := db.Q.ListBooksPaginated(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	switch format {
	case "csv":
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=books_page_%d.csv", page))
		c.Header("Content-Type", "text/csv")

		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// CSV header
		_ = writer.Write([]string{"ID", "Title", "Author", "Genre", "Published Year", "Available Copies", "Total Copies"})

		// CSV rows
		for _, book := range books {
			record := []string{
				book.ID.String(),
				book.Title,
				book.Author,
				book.Genre,
				fmt.Sprintf("%d", book.PublishedYear.Int32),
				fmt.Sprintf("%d", book.AvailableCopies.Int32),
				fmt.Sprintf("%d", book.TotalCopies),
			}
			_ = writer.Write(record)
		}

	case "pdf":
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(40, 10, fmt.Sprintf("Books - Page %d", page))
		pdf.Ln(12)

		pdf.SetFont("Arial", "", 12)
		for _, book := range books {
			line := fmt.Sprintf("• %s by %s (%d) | %s | %d/%d available",
				book.Title, book.Author, book.PublishedYear.Int32,
				book.Genre, book.AvailableCopies.Int32, book.TotalCopies)
			pdf.MultiCell(0, 8, line, "", "", false)
		}

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=books_page_%d.pdf", page))
		c.Header("Content-Type", "application/pdf")
		if err := pdf.Output(c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format, use ?format=csv or ?format=pdf"})
	}
}

func DownloadUsersHandler(c *gin.Context) {
	format := c.Query("format")
	page := 1
	limit := 10

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
	params := gen.ListUsersPaginatedParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	users, err := db.Q.ListUsersPaginated(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	switch format {
	case "csv":
		c.Header("Content-Disposition", "attachment; filename=users.csv")
		c.Header("Content-Type", "text/csv")

		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// Header row
		writer.Write([]string{"ID", "First Name", "Last Name", "Email", "Phone", "Role", "Created At"})

		for _, u := range users {
			writer.Write([]string{
				u.ID.String(),
				u.FirstName,
				u.LastName,
				u.Email,
				u.PhoneNumber,
				u.Role.String,
				u.CreatedAt.Time.Format("2006-01-02 15:04:05"),
			})
		}
		return

	case "pdf":
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(40, 10, "Users List")
		pdf.Ln(12)
		pdf.SetFont("Arial", "", 11)

		for _, u := range users {
			line := fmt.Sprintf("%s %s (%s) — %s | %s | %s",
				u.FirstName, u.LastName, u.Email, u.PhoneNumber, u.Role.String,
				u.CreatedAt.Time.Format("2006-01-02"))
			pdf.MultiCell(0, 8, line, "0", "L", false)
			pdf.Ln(2)
		}

		c.Header("Content-Disposition", "attachment; filename=users.pdf")
		c.Header("Content-Type", "application/pdf")
		err := pdf.Output(c.Writer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		}
		return

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format, use ?format=csv or ?format=pdf"})
	}
}
func DownloadBorrowsHandler(c *gin.Context) {
	format := c.Query("format") // "pdf" or "csv"

	borrows, err := db.Q.ListBorrow(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch borrow records"})
		return
	}

	switch format {
	case "csv":
		c.Header("Content-Disposition", "attachment; filename=borrows.csv")
		c.Header("Content-Type", "text/csv")

		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// Header row
		writer.Write([]string{"Borrow ID", "User ID", "Book ID", "Borrowed At", "Due Date", "Returned At"})

		for _, b := range borrows {
			var returned string
			if b.ReturnedAt.Valid {
				returned = b.ReturnedAt.Time.Format("2006-01-02 15:04:05")
			}
			writer.Write([]string{
				b.ID.String(),
				b.UserID.String(),
				b.BookID.String(),
				b.BorrowedAt.Time.Format("2006-01-02 15:04:05"),
				b.DueDate.Time.Format("2006-01-02 15:04:05"),
				returned,
			})
		}
		return

	case "pdf":
		pdf := gofpdf.New("P", "mm", "A4", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(40, 10, "Borrow Records")
		pdf.Ln(12)
		pdf.SetFont("Arial", "", 11)

		for _, b := range borrows {
			returned := "Not Returned"
			if b.ReturnedAt.Valid {
				returned = b.ReturnedAt.Time.Format("2006-01-02")
			}
			line := fmt.Sprintf("User: %s | Book: %s | Borrowed: %s | Due: %s | Returned: %s",
				b.UserID.String(), b.BookID.String(),
				b.BorrowedAt.Time.Format("2006-01-02"),
				b.DueDate.Time.Format("2006-01-02"),
				returned)
			pdf.MultiCell(0, 8, line, "0", "L", false)
			pdf.Ln(2)
		}

		c.Header("Content-Disposition", "attachment; filename=borrows.pdf")
		c.Header("Content-Type", "application/pdf")
		err := pdf.Output(c.Writer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		}
		return

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format, use ?format=csv or ?format=pdf"})
	}
}
