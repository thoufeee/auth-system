package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// user model

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string    `gorm:"not null" json:"name"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	Age          int       `gorm:"not null" json:"age"`
	Phone        string    `gorm:"uniqueIndex" json:"phone"`
	PasswordHash string    `gorm:"not null" json:"-"`

	Role   string `gorm:"not null;default:USER" json:"role"`
	Status string `gorm:"not null;default:ACTIVE" json:"status"`

	Trainer *Trainer `gorm:"foreignKey:UserID" json:"trainer,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
