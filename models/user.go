package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RECRUITER model

type Recruiter struct {
	BaseModel
	Name         string  `gorm:"type:varchar(150);not null" json:"name"`
	Email        string  `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PhoneNumber  string  `gorm:"type:varchar(13);not null" json:"phone_number" validate:"required,e164"`
	PasswordHash *string `gorm:"type:varchar(255)" json:"-"`

	Role       string `gorm:"type:varchar(30);not null" json:"role"`
	IsVerified bool   `gorm:"default:false" json:"is_verified"`
	Status     string `gorm:"type:varchar(20);not null;default:ACTIVE" json:"status"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// candidate model
type Candidate struct {
	BaseModel
	Name        string  `gorm:"type:varchar(150);not null" json:"name"`
	Email       string  `gorm:"type:varchar(255);uniqueIndex;not null" json:"email" validate:"required,email"`
	PhoneNumber *string `gorm:"type:varchar(13)" json:"phone_number,omitempty" validate:"omitempty,e164"`

	IsVerified bool   `gorm:"default:false" json:"is_verified"`
	Status     string `gorm:"type:varchar(20);not null;default:ACTIVE" json:"status"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// for unique id
type BaseModel struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}

	return nil
}
