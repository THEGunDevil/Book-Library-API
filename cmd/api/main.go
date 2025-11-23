package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/config"
	"github.com/THEGunDevil/GoForBackend/internal/db"
	"github.com/THEGunDevil/GoForBackend/internal/handlers"
	"github.com/THEGunDevil/GoForBackend/internal/middleware"
	"github.com/THEGunDevil/GoForBackend/internal/service"
	"github.com/gin-contrib/cors"
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
	// db.LocalConnect(cfg)
	defer db.Close()

	r := gin.New() // instead of gin.Default() if you want full control
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://himel-s-library.vercel.app", "http://localhost:3000", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.RateLimiter())

	// Health check
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/download/books", middleware.AuthMiddleware(), middleware.AdminOnly(), handlers.DownloadSearchBooksHandler)
	r.GET("/download/users", middleware.AuthMiddleware(), middleware.AdminOnly(), handlers.DownloadUsersHandler)
	r.GET("/download/borrows", middleware.AuthMiddleware(), middleware.AdminOnly(), handlers.DownloadBorrowsHandler)
	r.POST("/stripe/webhook", handlers.StripeWebhookHandler)
	r.GET("/stripe/success", handlers.StripeSuccessHandler)
	r.GET("/stripe/cancel", handlers.StripeCancelHandler)
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
		userGroup.GET("/",middleware.SkipRateLimit(), middleware.AdminOnly(), handlers.GetUsersHandler)
		userGroup.GET("user/email", middleware.AdminOnly(),middleware.SkipRateLimit(), handlers.SearchUsersPaginatedHandler)
		userGroup.PATCH("/user/ban/:id", middleware.AdminOnly(), handlers.BanUserByIDHandler)
	}
	bannedUserGroup := r.Group("/banned-users")
	bannedUserGroup.Use(middleware.AuthMiddleware())
	{
		bannedUserGroup.GET("/", middleware.AdminOnly(), handlers.GetUsersHandler)
	}
	// Book routes
	bookGroup := r.Group("/books")
	{
		// Public
		bookGroup.GET("/",middleware.SkipRateLimit(), handlers.GetBooksHandler)
		bookGroup.GET("/:id", handlers.GetBookByIDHandler)
		bookGroup.GET("/search",middleware.SkipRateLimit(), handlers.SearchBooksPaginatedHandler)
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
		reservationGroup.GET("/", handlers.GetReservationsHandler)
		reservationGroup.PATCH("/:id/status", middleware.AdminOnly(), handlers.UpdateReservationStatusHandler)
		reservationGroup.GET("/next/:id", middleware.AdminOnly(), handlers.GetNextReservationHandler)
	}

	// Borrow routes (protected, for any logged-in user)
	borrowGroup := r.Group("/borrows")
	borrowGroup.Use(middleware.AuthMiddleware())
	{
		borrowGroup.GET("/", middleware.AdminOnly(),middleware.SkipRateLimit(), handlers.GetAllBorrowsHandlers)
		borrowGroup.GET("/user/:id", handlers.GetBorrowsByUserIDHandler)
		borrowGroup.GET("/book/:id", handlers.GetBorrowsByBookIDHandler)
		borrowGroup.GET("/borrow/book/:id", handlers.GetBorrowByBookAndUserIDHandler)
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
	notificationGroup := r.Group("/notifications")
	notificationGroup.Use(middleware.AuthMiddleware())
	{
		notificationGroup.GET("/get", handlers.GetUserNotificationsByUserIDHandler)
		notificationGroup.PATCH("/mark-read", handlers.MarkAllNotificationsAsReadHandler)
	}

	subscriptionPlanGroup := r.Group("/subscription-plan")
	subscriptionPlanGroup.Use(middleware.AuthMiddleware())
	{
		subscriptionPlanGroup.GET("/", handlers.GetSubscriptionsPlanHandler)
		subscriptionPlanGroup.GET("/:id", handlers.GetSubscriptionPlanByIDHandler)
		subscriptionPlanGroup.POST("/", middleware.AdminOnly(), handlers.CreateSubscriptionPlanHandler)
		subscriptionPlanGroup.DELETE("/:id", middleware.AdminOnly(), handlers.DeleteSubscriptionPlanByID)
	}

	subscriptionGroup := r.Group("/subscription")
	subscriptionGroup.Use(middleware.AuthMiddleware())
	{
		subscriptionGroup.GET("/", middleware.AdminOnly(), handlers.ListSubscriptionsHandler)
		subscriptionGroup.GET("/:user_id", handlers.GetSubscriptionByUserIDHandler)
		subscriptionGroup.GET("/user/:user_id", middleware.AdminOnly(), handlers.ListSubscriptionsByUserHandler)
		subscriptionGroup.POST("/", handlers.CreateSubscriptionHandler)
		subscriptionGroup.DELETE("/:id", middleware.AdminOnly(), handlers.DeleteSubscriptionByIDHandler)
	}
	paymentGroup := r.Group("/payments")
	paymentGroup.Use(middleware.AuthMiddleware())
	{
		paymentGroup.POST("/payment", handlers.CreatePaymentHandler)                                           // create payment
		paymentGroup.GET("/:id", handlers.GetPaymentHandler)                                                   // get by ID
		paymentGroup.GET("/payments", middleware.AdminOnly(), handlers.ListAllPaymentsHandler) // list by user
		paymentGroup.PATCH("/payment/:id/status", handlers.UpdatePaymentStatusHandler)                         // update status
		paymentGroup.DELETE("/payment/:id", middleware.AdminOnly(), handlers.DeletePaymentByIDHandler)
	}
	refundGroup := r.Group("/refunds")
	refundGroup.Use(middleware.AuthMiddleware())
	{
		refundGroup.POST("/", handlers.CreateRefundHandler)                                                   // create refund
		refundGroup.GET("/:id", handlers.GetRefundHandler)                                                    // get by ID
		refundGroup.GET("/payment/:payment_id", middleware.AdminOnly(), handlers.ListRefundsByPaymentHandler) // list refunds by payment
		refundGroup.GET("/status", middleware.AdminOnly(), handlers.ListRefundsByStatusHandler)               // list refunds by status (query param: ?status=requested)
		refundGroup.PATCH("/:id/status", handlers.UpdateRefundStatusHandler)                                  // update status & processed_at
		refundGroup.DELETE("/:id", middleware.AdminOnly(), handlers.DeleteRefundHandler)                      // delete refund
	}
	listGroup := r.Group("/list")
	listGroup.Use(middleware.AuthMiddleware(), middleware.AdminOnly()) // ← Auth MUST come before Admin
	{
		listGroup.GET("/data-paginated", handlers.ListDataByStatusHandler)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on port %s", port)
	r.Run(":" + port)
}
