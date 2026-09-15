package models

import (
	"time"

	"github.com/google/uuid"
)

// user model

type User struct {
	ID        uuid.UUID
	Name      string    `json:"name"`
	Password  string    `json:"password"`
	Age       int       `json:"age"`
	Place     string    `json:"place"`
	Email     string    `json:"email"`
	Gender    string    `json:"gender"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
