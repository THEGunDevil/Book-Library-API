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
	"time"
)

// setupPDF initializes a PDF with common settings, header, and footer
func setupPDF(title string) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Set fonts
	pdf.SetFont("Helvetica", "B", 16) // Use Helvetica for better readability
	pdf.SetTextColor(0, 0, 0)
	
	// Header
	pdf.CellFormat(0, 10, title, "", 1, "C", false, 0, "")
	pdf.Ln(5)
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf("Generated on %s", time.Now().Format("2006-01-02 15:04:05")), "", 1, "C", false, 0, "")
	pdf.Ln(5)

	// Footer
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(128, 128, 128)
		pdf.CellFormat(0, 10, fmt.Sprintf("Page %d", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	return pdf
}

// drawTableHeader draws a table header with consistent styling
func drawTableHeader(pdf *gofpdf.Fpdf, headers []string, widths []float64) {
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(100, 100, 100)
	
	for i, header := range headers {
		pdf.CellFormat(widths[i], 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)
}

// drawTableRow draws a table row with alternating background colors
func drawTableRow(pdf *gofpdf.Fpdf, row []string, widths []float64, rowIndex int) {
	pdf.SetFont("Helvetica", "", 9)
	if rowIndex%2 == 0 {
		pdf.SetFillColor(240, 240, 240) // Light gray for even rows
	} else {
		pdf.SetFillColor(255, 255, 255) // White for odd rows
	}
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(100, 100, 100)
	
	for i, cell := range row {
		pdf.CellFormat(widths[i], 8, cell, "1", 0, "L", true, 0, "")
	}
	pdf.Ln(-1)
}

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
		pdf := setupPDF(fmt.Sprintf("Books Report - Page %d", page))
		
		// Table headers
		headers := []string{"ID", "Title", "Author", "Genre", "Year", "Available", "Total"}
		widths := []float64{30, 50, 40, 30, 20, 20, 20}
		drawTableHeader(pdf, headers, widths)

		// Table rows
		for i, book := range books {
			row := []string{
				book.ID.String(),
				book.Title,
				book.Author,
				book.Genre,
				fmt.Sprintf("%d", book.PublishedYear.Int32),
				fmt.Sprintf("%d", book.AvailableCopies.Int32),
				fmt.Sprintf("%d", book.TotalCopies),
			}
			drawTableRow(pdf, row, widths, i)
		}

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=books_page_%d.pdf", page))
		c.Header("Content-Type", "application/pdf")
		if err := pdf.Output(c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
			return
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

	case "pdf":
		pdf := setupPDF("Users Report")
		
		// Table headers
		headers := []string{"ID", "First Name", "Last Name", "Email", "Phone", "Role", "Created At"}
		widths := []float64{30, 30, 30, 40, 30, 20, 30}
		drawTableHeader(pdf, headers, widths)

		// Table rows
		for i, u := range users {
			row := []string{
				u.ID.String(),
				u.FirstName,
				u.LastName,
				u.Email,
				u.PhoneNumber,
				u.Role.String,
				u.CreatedAt.Time.Format("2006-01-02"),
			}
			drawTableRow(pdf, row, widths, i)
		}

		c.Header("Content-Disposition", "attachment; filename=users.pdf")
		c.Header("Content-Type", "application/pdf")
		if err := pdf.Output(c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
			return
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format, use ?format=csv or ?format=pdf"})
	}
}

func DownloadBorrowsHandler(c *gin.Context) {
	format := c.Query("format")

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

	case "pdf":
		pdf := setupPDF("Borrow Records Report")
		
		// Table headers
		headers := []string{"Borrow ID", "User ID", "Book ID", "Borrowed At", "Due Date", "Returned At"}
		widths := []float64{30, 30, 30, 30, 30, 30}
		drawTableHeader(pdf, headers, widths)

		// Table rows
		for i, b := range borrows {
			returned := "Not Returned"
			if b.ReturnedAt.Valid {
				returned = b.ReturnedAt.Time.Format("2006-01-02")
			}
			row := []string{
				b.ID.String(),
				b.UserID.String(),
				b.BookID.String(),
				b.BorrowedAt.Time.Format("2006-01-02"),
				b.DueDate.Time.Format("2006-01-02"),
				returned,
			}
			drawTableRow(pdf, row, widths, i)
		}

		c.Header("Content-Disposition", "attachment; filename=borrows.pdf")
		c.Header("Content-Type", "application/pdf")
		if err := pdf.Output(c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
			return
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid format, use ?format=csv or ?format=pdf"})
	}
}