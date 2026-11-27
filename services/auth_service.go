package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

// structure of auth service
type AuthService struct {
	userRepo         *repository.UserRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	tokenService     *TokenService
}

func NewAuthService(
	userRepo *repository.UserRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	tokenService *TokenService,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenService:     tokenService,
	}
}

// register service
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

// login service
func (s *AuthService) Login(
	ctx context.Context,
	req dto.LoginRequest,
) (*dto.AuthResponse, error) {

	logger := middleware.GetLogger(ctx)

	email := strings.ToLower(strings.TrimSpace(req.Email))
	phone := strings.TrimSpace(req.Phone)

	// Exactly one identifier must be provided
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

	// Verify password
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

	// Verify account status
	if user.Status != string(constants.UserActive) {
		logger.Warn(
			"login rejected: inactive account",
			"user_id", user.ID,
		)
		return nil, ErrUserInactive
	}

	// Generate access and refresh tokens
	tokenPair, err := s.tokenService.GenerateTokenPair(
		user.ID.String(),
		string(user.Role),
		"",
	)
	if err != nil {
		logger.Error("token generation failed", "error", err)
		return nil, err
	}

	// Hash refresh token before storing it
	refreshTokenHash := s.tokenService.HashRefreshToken(
		tokenPair.RefreshToken,
	)

	// Create refresh token family
	if err := s.refreshTokenRepo.CreateFamily(
		ctx,
		tokenPair.FamilyID,
		user.ID.String(),
		tokenPair.RefreshExpiry,
	); err != nil {
		logger.Error("refresh token family creation failed", "error", err)
		return nil, err
	}

	// Create refresh token record
	refreshToken := models.RefreshToken{
		TokenHash: refreshTokenHash,
		UserID:    user.ID.String(),
		FamilyID:  tokenPair.FamilyID,
		ExpiresAt: tokenPair.RefreshExpiry,
	}

	if err := s.refreshTokenRepo.Store(
		ctx,
		refreshToken,
	); err != nil {
		logger.Error("refresh token storage failed", "error", err)
		return nil, err
	}

	logger.Info(
		"user login successful",
		"user_id", user.ID,
	)

	return &dto.AuthResponse{
		AccessToken:   tokenPair.AccessToken,
		RefreshToken:  tokenPair.RefreshToken,
		TokenType:     "Bearer",
		AccessExpiry:  tokenPair.AccessExpiry,
		RefreshExpiry: tokenPair.RefreshExpiry,
		User: &dto.UserResponse{
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

// refresh token handlingv(rotating with newone)  >>>>
func (s *AuthService) Refresh(
	ctx context.Context,
	rawRefreshToken string,
) (*TokenPair, error) {

	logger := middleware.GetLogger(ctx)

	if rawRefreshToken == "" {
		return nil, ErrInvalidCredentials
	}

	//  Hash the incoming refresh token
	oldTokenHash := s.tokenService.HashRefreshToken(
		rawRefreshToken,
	)

	// Find the existing refresh token in Redis
	oldRecord, err := s.refreshTokenRepo.Get(
		ctx,
		oldTokenHash,
	)
	if err != nil {
		logger.Warn("refresh token lookup failed")
		return nil, err
	}

	// Retrieve the original session expiry.
	familyExpiry, err := s.refreshTokenRepo.GetFamilyExpiry(
		ctx,
		oldRecord.FamilyID,
	)
	if err != nil {
		logger.Warn("failed to retrieve refresh family expiry", "error", err)
		return nil, err
	}

	//  Retrieve the user
	user, err := s.userRepo.GetByID(
		ctx,
		oldRecord.UserID,
	)
	if err != nil {
		logger.Warn("refresh failed: user not found")
		return nil, ErrInvalidCredentials
	}

	// Check account status
	if user.Status != string(constants.UserActive) {
		logger.Warn(
			"refresh rejected: inactive account",
			"user_id", user.ID,
		)
		return nil, ErrUserInactive
	}

	//  Generate a new access and refresh token
	newTokenPair, err := s.tokenService.GenerateTokenPair(
		user.ID.String(),
		string(user.Role),
		oldRecord.FamilyID,
	)
	if err != nil {
		logger.Error("token generation failed", "error", err)
		return nil, err
	}

	// Preserve the original session expiry.
	newTokenPair.RefreshExpiry = familyExpiry
	//  Hash the new refresh token
	newTokenHash := s.tokenService.HashRefreshToken(
		newTokenPair.RefreshToken,
	)

	// Atomically rotate the old refresh token
	newRefreshToken := models.RefreshToken{
		TokenHash: newTokenHash,
		UserID:    user.ID.String(),
		FamilyID:  oldRecord.FamilyID,
		ExpiresAt: newTokenPair.RefreshExpiry,
	}

	if err := s.refreshTokenRepo.Rotate(
		ctx,
		oldTokenHash,
		newRefreshToken,
	); err != nil {

		if errors.Is(err, repository.ErrRefreshTokenReused) {
			logger.Warn(
				"refresh token reuse detected",
				"user_id", user.ID,
				"family_id", oldRecord.FamilyID,
			)
		}

		return nil, err
	}

	logger.Info(
		"access and refresh tokens rotated successfully",
		"user_id", user.ID,
	)

	return newTokenPair, nil
}

// logout >>>
func (s *AuthService) Logout(
	ctx context.Context,
	rawRefreshToken string,
) error {
	if rawRefreshToken == "" {
		return repository.ErrRefreshTokenNotFound
	}

	// Hash the raw refresh token
	tokenHash := s.tokenService.HashRefreshToken(rawRefreshToken)

	// Find the refresh token record
	tokenRecord, err := s.refreshTokenRepo.Get(ctx, tokenHash)
	if err != nil {
		return err
	}

	// Revoke the entire token family
	if err := s.refreshTokenRepo.RevokeFamily(
		ctx,
		tokenRecord.FamilyID,
	); err != nil {
		return err
	}

	return nil
}

// to logout from all devices
func (s *AuthService) LogoutAll(
	ctx context.Context,
	userID string,
) error {

	if userID == "" {
		return errors.New("user ID is required")
	}

	err := s.refreshTokenRepo.RevokeAllFamilies(
		ctx,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to logout from all devices: %w", err)
	}

	return nil
}
