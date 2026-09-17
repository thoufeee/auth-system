package models

import (
	"time"

	"github.com/google/uuid"
)

// user model

type User struct {
	ID         uuid.UUID
	Name       string    `json:"name"`
	Age        int       `json:"age"`
	Place      string    `json:"place"`
	Email      string    `json:"email"`
	Gender     string    `json:"gender"`
	ModifiedBy string    `json:"modified_by"`
	ModifiedAt string    `json:"modified_At"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// organizer model

type Organizer struct {
	ID               uuid.UUID
	OrganizationName string `json:"org_name"`
	Description      string `json:"description"`
	Phone1           string `json:"phone1"`
	Phone2           string `json:"phone2"`
	Address          string `json:"address"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ModifiedBy       string    `json:"modified_by"`
	ModifiedAt       time.Time `json:"modified_At"`
}
