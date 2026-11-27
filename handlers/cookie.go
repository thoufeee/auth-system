package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	accessCookieName  = "access_token"
	refreshCookieName = "refresh_token"
)

// to set the cookie >>>
func setAuthCookies(
	c *gin.Context,
	accessToken string,
	refreshToken string,
	accessExpiry time.Time,
	refreshExpiry time.Time,
) {
	secure := true // Set false only for local HTTP development.

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessCookieName,
		Value:    accessToken,
		Path:     "/",
		Expires:  accessExpiry,
		MaxAge:   int(time.Until(accessExpiry).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Path:     "/v1/user",
		Expires:  refreshExpiry,
		MaxAge:   int(time.Until(refreshExpiry).Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// for clearing cookies
func clearAuthCookies(c *gin.Context) {
	secure := true

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/v1/user",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
