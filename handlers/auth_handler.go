package handlers

import (
	"errors"
	"net/http"

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
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(
			"INVALID_REQUEST",
			"Invalid login request",
		))
		return
	}

	authResponse, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		// Keep your existing error-to-HTTP-status mapping here.
		c.JSON(http.StatusUnauthorized, response.Error(
			"LOGIN_FAILED",
			"Invalid credentials",
		))
		return
	}

	setAuthCookies(
		c,
		authResponse.AccessToken,
		authResponse.RefreshToken,
		authResponse.AccessExpiry,
		authResponse.RefreshExpiry,
	)

	// Do not return either token in the response body.
	c.JSON(http.StatusOK, response.Success(
		"LOGIN_SUCCESS",
		"Login successful",
		authResponse.User,
	))
}

// refersh handler
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshCookieName)
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, response.Error(
			"REFRESH_TOKEN_MISSING",
			"Refresh token is missing",
		))
		return
	}

	tokenPair, err := h.authService.Refresh(
		c.Request.Context(),
		refreshToken,
	)
	if err != nil {
		clearAuthCookies(c)

		c.JSON(http.StatusUnauthorized, response.Error(
			"REFRESH_FAILED",
			"Session expired. Please log in again",
		))
		return
	}

	setAuthCookies(
		c,
		tokenPair.AccessToken,
		tokenPair.RefreshToken,
		tokenPair.AccessExpiry,
		tokenPair.RefreshExpiry,
	)

	c.JSON(response.StatusOK, response.Error(
		response.MsgSuccess,
		"Tokens refreshed successfully",
	))

}

// logout handler

func (h *AuthHandler) Logout(c *gin.Context) {

	// Read refresh token from HttpOnly cookie.
	refreshToken, err := c.Cookie("refresh_token")

	if err != nil || refreshToken == "" {
		clearAuthCookies(c)

		c.JSON(http.StatusOK, response.Error(
			"LOGOUT_SUCCESS",
			"Logged out successfully",
		))
		return
	}

	// Revoke refresh token family.
	err = h.authService.Logout(
		c.Request.Context(),
		refreshToken,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(
			"LOGOUT_FAILED",
			"Unable to complete logout",
		))
		return
	}

	// Clear access and refresh cookies.
	clearAuthCookies(c)

	c.JSON(http.StatusOK, response.Error(
		"LOGOUT_SUCCESS",
		"Logged out successfully",
	))
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {

	// Get authenticated user ID from request context
	userID := middleware.GetUserID(c.Request.Context())

	if userID == "" {
		c.JSON(http.StatusUnauthorized, response.Error(
			"UNAUTHORIZED",
			"Authentication required",
		))
		return
	}

	// Revoke all refresh-token families
	err := h.authService.LogoutAll(
		c.Request.Context(),
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(
			"LOGOUT_FAILED",
			"Unable to logout from all devices",
		))
		return
	}

	// Clear cookies on the current browser
	clearAuthCookies(c)

	c.JSON(http.StatusOK, response.Error(
		"LOGOUT_ALL_SUCCESS",
		"Logged out from all devices successfully",
	))
}
