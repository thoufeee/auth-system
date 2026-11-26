package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/thoufeee/auth-system/db"
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
	db, err := db.ConnectPostgres()
	if err != nil {
		slog.Error("Database connection failed", "error", err)
		os.Exit(1)
	}

	slog.Info("PostgreSQL connected successfully")

	// Auto migrate models
	if err := db.AutoMigrate(
		&models.User{},
	); err != nil {
		slog.Error("Database migration failed", "error", err)
		os.Exit(1)
	}

	slog.Info("Database migration completed")

	// Initialize repository
	userRepo := repository.NewUserRepository(db)

	// Initialize service
	authService := service.NewAuthService(
		userRepo,
		jwtSecret,
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
