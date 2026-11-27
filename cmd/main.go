package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	db "github.com/thoufeee/auth-system/DB"
	"github.com/thoufeee/auth-system/handlers"
	"github.com/thoufeee/auth-system/models"
	"github.com/thoufeee/auth-system/repository"
	"github.com/thoufeee/auth-system/routes"
	service "github.com/thoufeee/auth-system/services"
	"github.com/thoufeee/common-system/middleware"
)

func main() {

	//  Load environment variables
	if err := godotenv.Load("config/.env"); err != nil {
		slog.Warn("Environment file not found; using system environment variables")
	}

	// Initialize structured logger
	appLogger := middleware.NewLogger()
	slog.SetDefault(appLogger)

	slog.Info("Starting Auth Service")

	// JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")

	// Configure Gin
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Connect PostgreSQL
	psqlDB, err := db.ConnectPostgres()
	if err != nil {
		slog.Error("Database connection failed", "error", err)
		os.Exit(1)
	}

	slog.Info("PostgreSQL connected successfully")

	//  connect redis
	redisClient, err := db.ConnectRedis()
	if err != nil {
		slog.Error("Redis connection failed", "error", err)
		os.Exit(1)
	}

	defer redisClient.Close()

	slog.Info("Redis connected successfully")

	// Auto migrate models
	if err := psqlDB.AutoMigrate(
		&models.User{},
	); err != nil {
		slog.Error("Database migration failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Database migration completed")

	// Initialize repository
	userRepo := repository.NewUserRepository(psqlDB)

	refreshTokenRepo := repository.NewRefreshTokenRepository(
		redisClient,
	)

	tokenService := service.NewTokenService(jwtSecret)

	authService := service.NewAuthService(
		userRepo,
		refreshTokenRepo,
		tokenService,
	)

	// Initialize handler
	authHandler := handlers.NewAuthHandler(authService)

	// Register routes
	routes.SetupAuthRoutes(router, authHandler, jwtSecret)

	// 12. Start HTTP server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Auth Service is running", "port", port)

	if err := router.Run(":" + port); err != nil {
		slog.Error("Auth Service failed to start", "error", err)
		os.Exit(1)
	}
}
