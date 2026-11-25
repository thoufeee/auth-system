package repository

import (
	"context"
	"errors"

	"github.com/thoufeee/auth-system/models"
	"github.com/thoufeee/common-system/middleware"
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
func (r *UserRepository) Create(
	ctx context.Context,
	user *models.User,
) error {
	logger := middleware.GetLogger(ctx)

	if err := r.DB.WithContext(ctx).Create(user).Error; err != nil {
		logger.Error(
			"failed to create user",
			"error", err,
		)
		return err
	}

	logger.Info(
		"user created in database",
		"user_id", user.ID,
	)

	return nil
}

// check the mail \\\
func (r *UserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*models.User, error) {
	logger := middleware.GetLogger(ctx)

	var user models.User

	err := r.DB.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not found by email")
			return nil, err
		}

		logger.Error(
			"failed to fetch user by email",
			"error", err,
		)
		return nil, err
	}

	return &user, nil
}

// check the phone ///
func (r *UserRepository) GetByPhone(
	ctx context.Context,
	phone string,
) (*models.User, error) {
	logger := middleware.GetLogger(ctx)

	var user models.User

	err := r.DB.WithContext(ctx).
		Where("phone = ?", phone).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("user not found by phone")
			return nil, err
		}

		logger.Error(
			"failed to fetch user by phone",
			"error", err,
		)
		return nil, err
	}

	return &user, nil
}

// check funcinality
func (r *UserRepository) Exists(
	ctx context.Context,
	email string,
	phone string,
) (bool, error) {
	logger := middleware.GetLogger(ctx)

	var count int64

	err := r.DB.WithContext(ctx).
		Model(&models.User{}).
		Where("email = ? OR phone = ?", email, phone).
		Count(&count).Error

	if err != nil {
		logger.Error(
			"failed to check existing user",
			"error", err,
		)
		return false, err
	}

	return count > 0, nil
}
