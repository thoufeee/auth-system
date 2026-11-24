package repository

import (
	"auth_system/models"
	"errors"

	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

// create user account
func (r *UserRepository) Create(user *models.User) error {
	return r.DB.Create(user).Error
}

// check the mail \\\
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User

	err := r.DB.Where("email = ?", email).First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// check the phone ///
func (r *UserRepository) GetByPhone(phone string) (*models.User, error) {
	var user models.User

	err := r.DB.Where("phone = ?", phone).First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// check funcinality
func (r *UserRepository) Exists(email, phone string) (bool, error) {
	var count int64

	err := r.DB.model(&models.User{}).
		Where("email = ? OR phone = ?", email, phone).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// error
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
