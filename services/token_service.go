package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenService struct {
	jwtSecret     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

type TokenClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`

	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken   string
	RefreshToken  string
	FamilyID      string
	AccessExpiry  time.Time
	RefreshExpiry time.Time
}

func NewTokenService(
	jwtSecret string,
) *TokenService {
	return &TokenService{
		jwtSecret:     []byte(jwtSecret),
		accessExpiry:  15 * time.Minute,
		refreshExpiry: 7 * 24 * time.Hour,
	}
}

func (s *TokenService) GenerateAccessToken(
	userID string,
	role string,
) (string, time.Time, error) {

	now := time.Now()
	expiresAt := now.Add(s.accessExpiry)

	claims := TokenClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expiresAt, nil
}

func generateRefreshToken() (string, error) {
	randomBytes := make([]byte, 32)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(randomBytes), nil
}

func hashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}

func (s *TokenService) GenerateTokenPair(
	userID string,
	role string,
	familyID string,
) (*TokenPair, error) {

	accessToken, accessExpiry, err := s.GenerateAccessToken(
		userID,
		role,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	if familyID == "" {
		familyID = uuid.NewString()
	}

	return &TokenPair{
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		FamilyID:      familyID,
		AccessExpiry:  accessExpiry,
		RefreshExpiry: time.Now().Add(s.refreshExpiry),
	}, nil
}

func (s *TokenService) HashRefreshToken(
	token string,
) string {
	return hashRefreshToken(token)
}
