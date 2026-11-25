package handlers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/thoufeee/auth-system/dto"
	service "github.com/thoufeee/auth-system/services"
	"github.com/thoufeee/common-system/middleware"
	"github.com/thoufeee/common-system/response"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register user
func (h *AuthHandler) Register(c *gin.Context) {
	ctx := c.Request.Context()
	logger := middleware.GetLogger(ctx)

	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn(
			"invalid registration request",
			"error", err,
		)

		c.JSON(response.StatusBadRequest, response.Error(
			response.MsgBadRequest,
			"Please provide valid registration details",
		))
		return
	}

	user, err := h.authService.RegisterUser(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserExists):
			c.JSON(response.StatusConflict, response.Error(
				response.MsgConflict,
				"Email or phone number already registered",
			))

		default:
			logger.Error(
				"registration handler failed",
				"error", err,
			)

			c.JSON(response.StatusInternalServerError, response.Error(
				response.CodeInternal,
				"Unable to register user",
			))
		}
		return
	}

	c.JSON(response.StatusOK, response.Success(
		response.CodeSuccess,
		"User registered successfully",
		*user,
	))
}

// Login user
func (h *AuthHandler) Login(c *gin.Context) {
	ctx := c.Request.Context()
	logger := middleware.GetLogger(ctx)

	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn(
			"invalid login request",
			"error", err,
		)

		c.JSON(response.StatusBadRequest, response.Error(
			response.CodeBadRequest,
			"Please provide valid login details",
		))
		return
	}

	auth, err := h.authService.Login(ctx, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(response.StatusUnauthorized, response.Error(
				response.CodeUnauthorized,
				"Invalid email/phone or password",
			))

		case errors.Is(err, service.ErrUserInactive):
			c.JSON(response.StatusForbidden, response.Error(
				response.MsgForbidden,
				"Your account is not active",
			))

		default:
			logger.Error(
				"login handler failed",
				"error", err,
			)

			c.JSON(response.StatusInternalServerError, response.Error(
				response.MsgInternal,
				"Unable to process login",
			))
		}
		return
	}

	c.JSON(response.StatusOK, response.Success(
		response.CodeSuccess,
		"Login successful",
		*auth,
	))
}
