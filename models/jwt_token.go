package models

import "time"

// refresh token

type RefreshToken struct {
	TokenHash string
	UserID    string
	FamilyID  string
	ExpiresAt time.Time
}
