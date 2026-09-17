package models

import (
	"time"

	"github.com/google/uuid"
)

// register user

type RegisterUser struct {
	Name   string `json:"name" validate:"required,min=2,max=50"`
	Email  string `json:"email" validate:"required,min=2,max=50"`
	Phone  string `json:"phone1" validate:"required,min=12,max=12"`
	Age    int    `json:"age" validate:"required"`
	Place  string `json:"place" validate:"required"`
	Gender string `json:"gender" validate:"required"`
}

// register organiser
type RegisterOrganizer struct {
	ID               uuid.UUID
	OrganizationName string `json:"org_name" validate:"required,min=2,max=50"`
	Description      string `json:"description" validate:"required,min=2,max=50"`
	Phone1           string `json:"phone1" validate:"required,min=12,max=12"`
	Phone2           string `json:"phone2"  validate:"required,min=12,max=12"`
	Address          string `json:"address"  validate:"required,min=2,max=150"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ModifiedBy       string    `json:"modified_by"`
	ModifiedAt       time.Time `json:"modified_At"`
}
