package main

import (
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/indwar7/safaipay-backend/config"
	"github.com/indwar7/safaipay-backend/internal/auth"
	"github.com/indwar7/safaipay-backend/internal/badges"
	"github.com/indwar7/safaipay-backend/internal/booking"
	"github.com/indwar7/safaipay-backend/internal/collector"
	"github.com/indwar7/safaipay-backend/internal/leaderboard"
	"github.com/indwar7/safaipay-backend/internal/notification"
	"github.com/indwar7/safaipay-backend/internal/payment"
	"github.com/indwar7/safaipay-backend/internal/report"
	"github.com/indwar7/safaipay-backend/internal/user"
	"github.com/indwar7/safaipay-backend/pkg/database"
	"github.com/indwar7/safaipay-backend/pkg/middleware"
	"github.com/indwar7/safaipay-backend/pkg/sms"
	"github.com/indwar7/safaipay-backend/pkg/storage"
)

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load config
	cfg := config.Load()

	// Connect PostgreSQL
	db, err := database.NewPostgres(&cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}

	// Auto-migrate models
	if err := database.AutoMigrate(db,
		&user.User{},
		&collector.Collector{},
		&report.Report{},
		&booking.Booking{},
		&payment.Transaction{},
		&badges.Badge{},
		&badges.UserBadge{},
	); err != nil {
		log.Fatalf("failed to auto-migrate: %v", err)
	}

	// Connect Redis (non-fatal — app works without it, just no rate limiting/leaderboard)
	rdb, err := database.NewRedis(&cfg.Redis)
	if err != nil {
		slog.Warn("failed to connect to Redis, some features will be unavailable", "error", err)
	}

	// Initialize external services
	firebaseAuth := sms.NewFirebaseAuthService(cfg.FCM.ProjectID)

	r2Service, err := storage.NewR2Service(&cfg.R2)
	if err != nil {
		slog.Warn("failed to initialize R2 storage, image uploads will be unavailable", "error", err)
		r2Service = nil
	}

	notifService := notification.NewService(cfg.FCM.ProjectID, cfg.FCM.ServiceAccountJSON)

	// Initialize services
	userService := user.NewService(db)
	collectorService := collector.NewService(db)
	authService := auth.NewService(firebaseAuth, userService, collectorService, &cfg.JWT)
	reportService := report.NewService(db, userService, r2Service, notifService)
	bookingService := booking.NewService(db, userService, notifService)
	paymentService := payment.NewService(db, userService, notifService, &cfg.Razorpay)
	leaderboardService := leaderboard.NewService(rdb)
	badgesService := badges.NewService(db, userService, notifService)

	// Seed badges
	if err := badgesService.SeedBadges(nil); err != nil {
		slog.Error("failed to seed badges", "error", err)
	}

	// Initialize handlers
	authHandler := auth.NewHandler(authService)
	userHandler := user.NewHandler(userService)
	reportHandler := report.NewHandler(reportService, r2Service)
	bookingHandler := booking.NewHandler(bookingService)
	paymentHandler := payment.NewHandler(paymentService)
	leaderboardHandler := leaderboard.NewHandler(leaderboardService)
	badgesHandler := badges.NewHandler(badgesService)
	collectorHandler := collector.NewHandler(collectorService, bookingService, notifService, r2Service)

	// Setup Gin
	gin.SetMode(cfg.Server.Mode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORS())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "time": time.Now().Format(time.RFC3339)})
	})

	api := r.Group("/api/v1")

	// Auth routes (public, rate limited)
	authRoutes := api.Group("/auth")
	{
		authRoutes.POST("/send-otp",
			middleware.RateLimitOTP(rdb, 5, 15*time.Minute),
			authHandler.SendOTP,
		)
		authRoutes.POST("/verify-otp", authHandler.VerifyOTP)
	}

	// Protected user routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		// User
		userRoutes := protected.Group("/user")
		{
			userRoutes.GET("/profile", userHandler.GetProfile)
			userRoutes.PATCH("/profile", userHandler.UpdateProfile)
			userRoutes.POST("/checkin", userHandler.DailyCheckIn)
		}

		// Reports
		reportRoutes := protected.Group("/reports")
		{
			reportRoutes.POST("", reportHandler.CreateReport)
			reportRoutes.GET("", reportHandler.ListReports)
			reportRoutes.GET("/:id", reportHandler.GetReport)
			reportRoutes.PATCH("/:id/status", reportHandler.UpdateStatus)
		}

		// Bookings
		bookingRoutes := protected.Group("/bookings")
		{
			bookingRoutes.POST("", bookingHandler.CreateBooking)
			bookingRoutes.GET("", bookingHandler.ListBookings)
			bookingRoutes.GET("/:id", bookingHandler.GetBooking)
			bookingRoutes.PATCH("/:id/status", bookingHandler.UpdateStatus)
		}

		// Wallet & Payment
		walletRoutes := protected.Group("/wallet")
		{
			walletRoutes.GET("", paymentHandler.GetWallet)
			walletRoutes.POST("/redeem", paymentHandler.RedeemPoints)
			walletRoutes.POST("/withdraw", paymentHandler.Withdraw)
			walletRoutes.GET("/transactions", paymentHandler.GetTransactions)
		}
		api.POST("/payment/verify",
			middleware.AuthMiddleware(cfg.JWT.Secret),
			paymentHandler.VerifyPayment,
		)

		// Leaderboard
		lbRoutes := protected.Group("/leaderboard")
		{
			lbRoutes.GET("", leaderboardHandler.GetLeaderboard)
			lbRoutes.GET("/me", leaderboardHandler.GetMyRank)
		}

		// Badges
		badgeRoutes := protected.Group("/badges")
		{
			badgeRoutes.GET("", badgesHandler.GetAllBadges)
			badgeRoutes.GET("/user", badgesHandler.GetUserBadges)
		}
	}

	// Collector routes (separate auth)
	collectorRoutes := api.Group("/collector")
	{
		collectorRoutes.POST("/send-otp",
			middleware.RateLimitOTP(rdb, 5, 15*time.Minute),
			authHandler.SendCollectorOTP,
		)
		collectorRoutes.POST("/verify-otp", authHandler.VerifyCollectorOTP)

		// Protected collector routes
		collectorProtected := collectorRoutes.Group("")
		collectorProtected.Use(middleware.CollectorAuthMiddleware(cfg.JWT.Secret))
		{
			collectorProtected.GET("/profile", collectorHandler.GetProfile)
			collectorProtected.GET("/bookings", collectorHandler.GetBookings)
			collectorProtected.PATCH("/bookings/:id/complete", collectorHandler.CompleteBooking)
			collectorProtected.PATCH("/location", collectorHandler.UpdateLocation)
		}
	}

	// Start server
	port := cfg.Server.Port
	slog.Info("starting SafaiPay backend", "port", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
