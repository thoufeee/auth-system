package dto

import "time"

// for auth response >>> user for login respone

type AuthResponse struct {
	AccessToken   string        `json:"-"`
	RefreshToken  string        `json:"-"`
	TokenType     string        `json:"token_type"`
	AccessExpiry  time.Time     `json:"-"`
	RefreshExpiry time.Time     `json:"-"`
	User          *UserResponse `json:"user"`
}

// refresh token

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
