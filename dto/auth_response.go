package dto

import "time"

// for auth response >>> user for login respone

type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         UserResponse `json:"user"`
}

// refresh token

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
