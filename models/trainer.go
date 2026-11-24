package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// trainer model

type Trainer struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`

	Bio             string   `gorm:"type:text" json:"bio"`
	Specializations []string `gorm:"type:text[]" json:"specializations"`
	ExperienceYears int      `json:"experience_years"`

	Certifications []string `gorm:"type:text[]" json:"certifications"`

	HourlyRate   float64 `gorm:"type:decimal(10,2)" json:"hourly_rate"`
	ProfileImage string  `json:"profile_image"`

	Status string `gorm:"not null;default:PENDING" json:"status"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}
