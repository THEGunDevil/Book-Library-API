package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/models"
	"github.com/gin-gonic/gin"
)

func OverviewHandler(c *gin.Context) {
	now := time.Now()

	// -------------------- STATS --------------------
	row, err := db.Q.GetStats(c, gen.GetStatsParams{
		Column1: int32(now.Month()),
		Column2: int32(now.Year()),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// RevenueMonth may be string or float64 depending on DB type
	var rev float64
	switch v := row.RevenueMonth.(type) {
	case float64:
		rev = v
	case string:
		rev, _ = strconv.ParseFloat(v, 64)
	}

	res := models.OverviewResponse{
		Stats: models.Stats{
			TotalBooks:         int(row.TotalBooks),
			ActiveUsers:        int(row.ActiveUsers),
			TotalSubscriptions: int(row.TotalSubscriptions),
			RevenueMonth:       rev,
		},
	}

	// -------------------- BOOKS PER MONTH --------------------
	dbBpm, err := db.Q.GetBooksPerMonth(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, b := range dbBpm {
		res.BooksPerMonth = append(res.BooksPerMonth, models.BooksPerMonth{
			Month: b.Month,
			Books: int(b.Books),
		})
	}

	// -------------------- CATEGORY DATA --------------------
	dbCat, err := db.Q.GetCategoryData(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, cdata := range dbCat {
		res.CategoryData = append(res.CategoryData, models.CategoryData{
			Name: cdata.Name,
			Value:    int(cdata.Value),
		})
	}

	// -------------------- TOP BORROWED BOOKS --------------------
	dbTop, err := db.Q.GetTopBorrowedBooks(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, t := range dbTop {
		res.TopBorrowedBooks = append(res.TopBorrowedBooks, models.TopBorrowedBook{
			Title: t.Name,
			Count: int(t.Count),
		})
	}

	// -------------------- SUBSCRIPTION PLANS --------------------
	dbPlans, err := db.Q.GetSubscriptionPlans(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, p := range dbPlans {
		res.SubscriptionPlans = append(res.SubscriptionPlans, models.SubscriptionPlanCount{
			Name:  p.Plan,
			Count: int(p.Count),
		})
	}

	// -------------------- SUBSCRIPTION HISTORY --------------------
	dbHist, err := db.Q.GetSubscriptionHistory(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, h := range dbHist {
		res.SubscriptionHistory = append(res.SubscriptionHistory, models.SubscriptionHistory{
			Month: h.Month,
			Active:  int(h.Active),
			Cancelled: int(h.Cancelled),
		})
	}

	c.JSON(http.StatusOK, res)
}
