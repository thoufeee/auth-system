package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/thoufeee/auth-system/dto"
	"github.com/thoufeee/auth-system/models"
	"github.com/thoufeee/auth-system/repository"
	"github.com/thoufeee/common-system/constants"
	"github.com/thoufeee/common-system/middleware"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserInactive       = errors.New("user account is inactive")
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(
	userRepo *repository.UserRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) RegisterUser(
	ctx context.Context,
	req dto.RegisterRequest,
) (*dto.UserResponse, error) {
	logger := middleware.GetLogger(ctx)

	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)

	// Check whether user already exists.
	exists, err := s.userRepo.Exists(ctx, email, phone)
	if err != nil {
		logger.Error("registration failed during existence check", "error", err)
		return nil, err
	}

	if exists {
		logger.Warn("registration rejected: user already exists")
		return nil, ErrUserExists
	}

	// Hash password.
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		logger.Error("password hashing failed", "error", err)
		return nil, err
	}

	user := &models.User{
		Name:         strings.TrimSpace(req.Name),
		Email:        email,
		Phone:        phone,
		PasswordHash: string(hashedPassword),
		Role:         constants.RoleUser,
		Status:       string(constants.UserActive),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		logger.Error("user registration failed", "error", err)
		return nil, err
	}

	logger.Info(
		"user registration successful",
		"user_id", user.ID,
	)

	return &dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      string(user.Role),
		Status:    string(user.Status),
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *AuthService) Login(
	ctx context.Context,
	req dto.LoginRequest,
) (*dto.AuthResponse, error) {
	logger := middleware.GetLogger(ctx)

	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)

	// Exactly one identifier must be provided.
	if (email == "") == (phone == "") {
		logger.Warn("login rejected: invalid login identifier")
		return nil, ErrInvalidCredentials
	}

	var (
		user *models.User
		err  error
	)

	if email != "" {
		user, err = s.userRepo.GetByEmail(ctx, email)
	} else {
		user, err = s.userRepo.GetByPhone(ctx, phone)
	}

	if err != nil {
		logger.Warn("login failed: user lookup unsuccessful")
		return nil, ErrInvalidCredentials
	}

	// Verify password.
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	); err != nil {
		logger.Warn(
			"login failed: password verification unsuccessful",
			"user_id", user.ID,
		)
		return nil, ErrInvalidCredentials
	}

	if user.Status != string(constants.UserActive) {
		logger.Warn(
			"login rejected: inactive account",
			"user_id", user.ID,
		)
		return nil, ErrUserInactive
	}

	// Generate JWT.
	expiresAt := time.Now().Add(15 * time.Minute)

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		logger.Error("JWT generation failed", "error", err)
		return nil, err
	}

	logger.Info(
		"user login successful",
		"user_id", user.ID,
	)

	return &dto.AuthResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User: dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Phone:     user.Phone,
			Role:      string(user.Role),
			Status:    string(user.Status),
			CreatedAt: user.CreatedAt,
		},
	}, nil
}
