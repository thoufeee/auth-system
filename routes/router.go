package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/thoufeee/auth-system/handlers"
)

func SetupAuthRoutes(
	router *gin.Engine,
	authHandler *handlers.AuthHandler, jwtSecret string,
) {
	// API version 1
	api := router.Group("/v1")

	// Authentication routes
	auth := api.Group("/user")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}
}
