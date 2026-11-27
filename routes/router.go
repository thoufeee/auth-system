package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/thoufeee/auth-system/handlers"
	"github.com/thoufeee/common-system/middleware"
)

func SetupAuthRoutes(
	router *gin.Engine,
	authHandler *handlers.AuthHandler, jwtSecret string,
) {
	// API version 1
	api := router.Group("/v1")

	// Authentication routes (open routes)
	auth := api.Group("/user")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
	}

	// protected routes
	protected := api.Group("/user")
	protected.Use(middleware.Authentication(jwtSecret))

	{
		protected.POST("/logout", authHandler.Logout)
		protected.POST("/logout/all", authHandler.LogoutAll)
	}
}
