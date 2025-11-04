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

// setupPDF (unchanged)
func setupPDF(title string) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Set fonts
	pdf.SetFont("Helvetica", "B", 16)
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

// truncateText ensures text fits within a given width by truncating with ellipsis
func truncateText(pdf *gofpdf.Fpdf, text string, maxWidth float64) string {
	const maxIterations = 100 // Prevent infinite loops
	origText := text
	for pdf.GetStringWidth(text) > maxWidth-2 && len(text) > 0 && maxIterations > 0 {
		text = text[:len(text)-1]
	}
	if len(text) < len(origText) {
		return text + "..."
	}
	return text
}

// drawTableHeader (unchanged from previous fix)
func drawTableHeader(pdf *gofpdf.Fpdf, headers []string, widths []float64) {
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(200, 200, 200)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(100, 100, 100)

	startY := pdf.GetY()
	for i, header := range headers {
		pdf.SetXY(15+sumWidths(widths, i), startY) // 15mm left margin
		pdf.MultiCell(widths[i], 4, header, "1", "C", true)
	}
	pdf.SetY(startY + 8)
}

// sumWidths (unchanged)
func sumWidths(widths []float64, i int) float64 {
	sum := 0.0
	for j := 0; j < i; j++ {
		sum += widths[j]
	}
	return sum
}

// drawTableRow (modified to use MultiCell and truncate text)
func drawTableRow(pdf *gofpdf.Fpdf, row []string, widths []float64, rowIndex int) {
	pdf.SetFont("Helvetica", "", 9)
	if rowIndex%2 == 0 {
		pdf.SetFillColor(240, 240, 240) // Light gray for even rows
	} else {
		pdf.SetFillColor(255, 255, 255) // White for odd rows
	}
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(100, 100, 100)

	startY := pdf.GetY()
	for i, cell := range row {
		// Truncate text to fit within column width
		truncatedCell := truncateText(pdf, cell, widths[i])
		pdf.SetXY(15+sumWidths(widths, i), startY)
		pdf.MultiCell(widths[i], 4, truncatedCell, "1", "L", true)
	}
	pdf.SetY(startY + 8) // Ensure consistent row height
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
		widths := []float64{25, 45, 35, 25, 15, 15, 15} // Total: 175mm
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