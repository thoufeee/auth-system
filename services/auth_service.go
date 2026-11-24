package service

import (
	"auth_system/dto"
	"auth_system/models"
	"auth_system/repository"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserExists         = errors.New("email or phone already registered")
	ErrInvalidCredentials = errors.New("invalid email/phone or password")
	ErrAccountBlocked     = errors.New("account is blocked")
	ErrInvalidLogin       = errors.New("provide either email or phone")
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
}

func NewAuthService(
	userRepo *repository.UserRepository,
	jwtSecret string,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

// USER SIGNUP >>>

func (s *AuthService) RegisterUser(
	req dto.RegisterRequest,
) (*dto.UserResponse, error) {

	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)

	exists, err := s.userRepo.Exists(email, phone)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return nil, err
	}

	user := models.User{
		Name:         strings.TrimSpace(req.Name),
		Email:        email,
		Phone:        phone,
		PasswordHash: string(hashedPassword),
		Role:         model.RoleUser,
		Status:       model.StatusActive,
	}

	if err := s.userRepo.Create(&user); err != nil {
		return nil, err
	}

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

// USER SIGNIN

func (s *AuthService) Login(
	req dto.LoginRequest,
) (*dto.AuthResponse, error) {

	// Exactly one login identifier is required.
	if (req.Email == "") == (req.Phone == "") {
		return nil, ErrInvalidLogin
	}

	var user *model.User
	var err error

	if req.Email != "" {
		email := strings.ToLower(strings.TrimSpace(req.Email))
		user, err = s.userRepo.GetByEmail(email)
	} else {
		phone := strings.TrimSpace(req.Phone)
		user, err = s.userRepo.GetByPhone(phone)
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password.
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(req.Password),
	)

	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check account status.
	if user.Status != model.StatusActive {
		return nil, ErrAccountBlocked
	}

	// Generate JWT.
	expiresAt := time.Now().Add(15 * time.Minute)

	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"role":    string(user.Role),
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

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
