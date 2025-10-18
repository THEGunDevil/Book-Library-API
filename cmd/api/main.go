package main

import (
	"fmt"
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
	godotenv.Load()
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
	r.Use(middleware.CORSMiddleware())

	// Health check
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth routes (public)
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", handlers.RegisterHandler)
		authGroup.POST("/login", handlers.LoginHandler)
	}

	// User routes (protected)
	userGroup := r.Group("/users")
	userGroup.Use(middleware.AuthMiddleware())
	{
		userGroup.GET("/", middleware.AdminOnly(), handlers.GetAllUsersHandler)
		userGroup.GET("/user", handlers.GetUserHandler)
		// only admin can update user info
		userGroup.PATCH("/:id", handlers.UpdateUserByIDHandler)
	}

	// Book routes
	bookGroup := r.Group("/books")
	{
		// Public
		bookGroup.GET("/", handlers.GetBooksHandler)
		bookGroup.GET("/:id", handlers.GetBookByIDHandler)

		// Admin-only
		bookGroup.POST("/", middleware.AuthMiddleware(), middleware.AdminOnly(), handlers.CreateBookHandler)
		bookGroup.PATCH("/:id", middleware.AuthMiddleware(), middleware.AdminOnly(), handlers.UpdateBookByIDHandler)
		bookGroup.DELETE("/:id", middleware.AuthMiddleware(), middleware.AdminOnly(), handlers.DeleteBookHandler)
	}

	// Borrow routes (protected, for any logged-in user)
	borrowGroup := r.Group("/borrows")
	borrowGroup.Use(middleware.AuthMiddleware())
	{
		borrowGroup.GET("/:id", handlers.GetBorrowsByIDHandler)
		borrowGroup.GET("/", handlers.GetAllBorrowsHandlers)
		borrowGroup.POST("/borrow", handlers.BorrowBookHandler)
		borrowGroup.PUT("/return", handlers.ReturnBookHandler)
	}
	reviewGroup := r.Group("/reviews")
	reviewGroup.Use(middleware.AuthMiddleware())
	{
		reviewGroup.POST("/review", handlers.CreateReviewHandler)
		reviewGroup.PATCH("/review/:id", handlers.UpdateReviewByIDHandler)
		reviewGroup.GET("/review/:id", handlers.GetReviewsByBookIDHandler)
	}

	r.Run(":8080")
}
