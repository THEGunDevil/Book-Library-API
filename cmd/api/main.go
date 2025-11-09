package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/THEGunDevil/GoForBackend/internal/config"
	"github.com/THEGunDevil/GoForBackend/internal/db"
	"github.com/THEGunDevil/GoForBackend/internal/handlers"
	"github.com/THEGunDevil/GoForBackend/internal/middleware"
	"github.com/THEGunDevil/GoForBackend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load env
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	// Init Cloudinary
	cldURL := fmt.Sprintf("cloudinary://%s:%s@%s", apiKey, apiSecret, cloudName)
	service.InitCloudinary(cldURL)

	// Load config & connect DB
	cfg := config.LoadConfig()
	db.Connect(cfg)
	defer db.Close()

	r := gin.Default()
	r.Use(middleware.CORSMiddleware(
		"https://himel-s-library.vercel.app",
		"http://localhost:3000", // dev frontend
	))

	// Health check
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/download/books", middleware.AuthMiddleware(), middleware.AdminOnly(), handlers.DownloadBooksHandler)
	r.GET("/download/users", middleware.AuthMiddleware(), middleware.AdminOnly(), handlers.DownloadUsersHandler)
	r.GET("/download/borrows", middleware.AuthMiddleware(), middleware.AdminOnly(), handlers.DownloadBorrowsHandler)

	// Auth routes (public)
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", handlers.RegisterHandler)
		authGroup.POST("/login", handlers.LoginHandler)
		authGroup.POST("/refresh", handlers.RefreshHandler) // ✅ add this
		authGroup.POST("/logout", handlers.LogoutHandler)
	}

	// User routes (protected)
	userGroup := r.Group("/users")
	userGroup.Use(middleware.AuthMiddleware())
	{
		userGroup.GET("/user/:id", handlers.GetUserByIDHandler)
		userGroup.GET("/user/profile/:id", handlers.GetProfileDataByIDHandler)
		userGroup.PATCH("/user/:id", handlers.UpdateUserByIDHandler)
		// only admin can update user info
		userGroup.GET("/", middleware.AdminOnly(), handlers.GetUsersHandler)
		userGroup.PATCH("/user/ban/:id", middleware.AdminOnly(), handlers.BanUserByIDHandler)
	}

	// Book routes
	bookGroup := r.Group("/books")
	{
		// Public
		bookGroup.GET("/", handlers.GetBooksHandler)
		bookGroup.GET("/:id", handlers.GetBookByIDHandler)
		bookGroup.GET("/search", handlers.SearchBooksHandler)
		bookGroup.GET("/genres", handlers.ListGenresHandler)
		bookGroup.GET("/genre/:genre", handlers.ListBooksByGenreHandler)
		// Admin-only
		bookGroup.POST("/", middleware.AuthMiddleware(), middleware.AdminOnly(), handlers.CreateBookHandler)
		bookGroup.PATCH("/:id", middleware.AuthMiddleware(), middleware.AdminOnly(), handlers.UpdateBookByIDHandler)
		bookGroup.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminOnly(), handlers.DeleteBookHandler)
	}

	reservationGroup := r.Group("/reservations")
	reservationGroup.Use(middleware.AuthMiddleware())
	{
		reservationGroup.POST("/", handlers.CreateReservationHandler)
		reservationGroup.GET("/book/:id", handlers.GetReservationsByBookIDHandler)
		reservationGroup.GET("/book/:id/user", handlers.GetReservationsByBookIDAndUserIDHandler)
		reservationGroup.GET("/reservation/:id", handlers.GetReservationsByReservationID)
		reservationGroup.GET("/", middleware.AdminOnly(), handlers.GetReservationsHandler)
		reservationGroup.PATCH("/:id/status", middleware.AdminOnly(), handlers.UpdateReservationStatusHandler)
		reservationGroup.GET("/next/:id", middleware.AdminOnly(), handlers.GetNextReservationHandler)
	}

	// Borrow routes (protected, for any logged-in user)
	borrowGroup := r.Group("/borrows")
	borrowGroup.Use(middleware.AuthMiddleware())
	{
		borrowGroup.GET("/", middleware.AdminOnly(), handlers.GetAllBorrowsHandlers)
		borrowGroup.GET("/:id", handlers.GetBorrowsByIDHandler)
		borrowGroup.POST("/borrow", handlers.BorrowBookHandler)
		borrowGroup.PATCH("/borrow/:id/return", handlers.ReturnBookHandler)
	}
	reviewGroup := r.Group("/reviews")
	reviewGroup.Use(middleware.AuthMiddleware())
	{
		reviewGroup.POST("/review", handlers.CreateReviewHandler)
		reviewGroup.PATCH("/review/:id", handlers.UpdateReviewByIDHandler)
		reviewGroup.GET("/book/:id", handlers.GetReviewsByBookIDHandler)
		reviewGroup.GET("/user/:id", handlers.GetReviewsByUserIDHandler)
		reviewGroup.GET("/review/:id", handlers.GetReviewsByReviewIDHandler)
		reviewGroup.DELETE("/review/:id", handlers.DeleteReviewsByIDHandler)
	}
	contactGroup := r.Group("/contact")
	contactGroup.Use(middleware.AuthMiddleware())
	{
		contactGroup.POST("/send", handlers.ContactHandler)
	}
	notificationGroup := r.Group("/notification")
	notificationGroup.Use(middleware.AuthMiddleware())
	{
		notificationGroup.GET("/get", handlers.GetUserNotificationByUserIDHandler)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s", port)
	r.Run(":" + port)
}
