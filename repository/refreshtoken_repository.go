package repository

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/thoufeee/auth-system/models"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenReused   = errors.New("refresh token already used")
	ErrRefreshFamilyRevoked = errors.New("refresh token family revoked")
)

// to loading the lua script >>>>>

//go:embed script/rotate_rtoken.lua
var rotateRefreshTokenScript string

//go:embed script/revoke_rfamily.lua
var revokeRefreshFamilyScript string

var rotateRefreshToken = redis.NewScript(
	rotateRefreshTokenScript,
)

var revokeRefreshFamily = redis.NewScript(revokeRefreshFamilyScript)

// r_token repo
type RefreshTokenRepository struct {
	client *redis.Client
}

func NewRefreshTokenRepository(
	client *redis.Client,
) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		client: client,
	}
}

type refreshTokenRecord struct {
	UserID   string `json:"user_id"`
	FamilyID string `json:"family_id"`
	Status   string `json:"status"`
}

// to store toekn
func (r *RefreshTokenRepository) Store(
	ctx context.Context,
	token models.RefreshToken,
) error {

	key := "refresh:" + token.TokenHash

	record := refreshTokenRecord{
		UserID:   token.UserID,
		FamilyID: token.FamilyID,
		Status:   "active",
	}

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal refresh token: %w", err)
	}

	ttl := time.Until(token.ExpiresAt)
	if ttl <= 0 {
		return errors.New("refresh token already expired")
	}

	err = r.client.SetNX(
		ctx,
		key,
		data,
		ttl,
	).Err()

	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}

	return nil
}

// get toekn >>
func (r *RefreshTokenRepository) Get(
	ctx context.Context,
	tokenHash string,
) (*refreshTokenRecord, error) {

	key := "refresh:" + tokenHash

	data, err := r.client.Get(ctx, key).Bytes()

	if errors.Is(err, redis.Nil) {
		return nil, ErrRefreshTokenNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get refresh token: %w", err)
	}

	var record refreshTokenRecord

	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("unmarshal refresh token: %w", err)
	}

	return &record, nil
}

// delete toekn //
func (r *RefreshTokenRepository) Delete(
	ctx context.Context,
	tokenHash string,
) error {

	key := "refresh:" + tokenHash

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return nil
}

// to rotate the refresh token >>>>>.
func (r *RefreshTokenRepository) Rotate(
	ctx context.Context,
	oldTokenHash string,
	newToken models.RefreshToken,
) error {

	oldKey := "refresh:" + oldTokenHash
	familyKey := familyKey(newToken.FamilyID)
	newKey := "refresh:" + newToken.TokenHash

	ttl := time.Until(newToken.ExpiresAt)

	if ttl <= 0 {
		return errors.New("new refresh token already expired")
	}

	newRecord := refreshTokenRecord{
		UserID:   newToken.UserID,
		FamilyID: newToken.FamilyID,
		Status:   "active",
	}

	data, err := json.Marshal(newRecord)
	if err != nil {
		return fmt.Errorf("marshal refresh token: %w", err)
	}

	result, err := rotateRefreshToken.Run(
		ctx,
		r.client,
		[]string{
			oldKey,
			familyKey,
			newKey,
		},
		string(data),
		int64(ttl.Milliseconds()),
	).Int()

	if err != nil {
		return fmt.Errorf("rotate refresh token: %w", err)
	}

	switch result {
	case 1:
		return nil
	case -1:
		return ErrRefreshTokenNotFound
	case -2:
		return ErrRefreshFamilyRevoked
	case -3:
		return ErrRefreshTokenReused
	case -4:
		return errors.New("new refresh token already exists")
	case -5:
		return errors.New("invalid refresh token TTL")
	default:
		return errors.New("unexpected refresh rotation result")
	}
}

// creating token family
func (r *RefreshTokenRepository) CreateFamily(
	ctx context.Context,
	familyID string,
	userID string,
	expiresAt time.Time,
) error {

	key := familyKey(familyID)
	data := map[string]string{
		"family_id":  familyID,
		"user_id":    userID,
		"status":     "active",
		"expires_at": expiresAt.UTC().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return errors.New("refresh family already expired")
	}

	// Store family information.
	if err := r.client.Set(ctx, key, jsonData, ttl).Err(); err != nil {
		return err
	}

	// Track this family under the user.
	if err := r.client.SAdd(
		ctx,
		userFamilyKey(userID),
		familyID,
	).Err(); err != nil {
		// Roll back the family key if indexing fails.
		_ = r.client.Del(ctx, key).Err()
		return err
	}

	// Keep the index until the session expires.
	return r.client.Expire(
		ctx,
		userFamilyKey(userID),
		ttl,
	).Err()
}

// to retrive the original family expiry
func (r *RefreshTokenRepository) GetFamilyExpiry(
	ctx context.Context,
	familyID string,
) (time.Time, error) {

	data, err := r.client.Get(
		ctx,
		familyKey(familyID),
	).Bytes()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return time.Time{}, ErrRefreshFamilyRevoked
		}
		return time.Time{}, err
	}

	var family struct {
		ExpiresAt string `json:"expires_at"`
	}

	if err := json.Unmarshal(data, &family); err != nil {
		return time.Time{}, err
	}

	expiresAt, err := time.Parse(time.RFC3339, family.ExpiresAt)
	if err != nil {
		return time.Time{}, err
	}

	if !expiresAt.After(time.Now()) {
		return time.Time{}, ErrRefreshFamilyRevoked
	}

	return expiresAt, nil
}

// to revoke  the toekn family
func (r *RefreshTokenRepository) RevokeFamily(
	ctx context.Context,
	familyID string,
) error {
	result, err := revokeRefreshFamily.Run(
		ctx,
		r.client,
		[]string{familyKey(familyID)},
	).Int()

	if err != nil {
		return fmt.Errorf("failed to revoke refresh token family: %w", err)
	}

	if result == 0 {
		return ErrRefreshFamilyRevoked
	}

	return nil
}

// to revoke all family
func (r *RefreshTokenRepository) RevokeAllFamilies(
	ctx context.Context,
	userID string,
) error {

	familyIDs, err := r.client.SMembers(
		ctx,
		userFamilyKey(userID),
	).Result()

	if err != nil {
		return fmt.Errorf("failed to retrieve user sessions: %w", err)
	}

	for _, familyID := range familyIDs {
		if err := r.RevokeFamily(ctx, familyID); err != nil &&
			!errors.Is(err, ErrRefreshFamilyRevoked) {
			return err
		}
	}

	return r.client.Del(ctx, userFamilyKey(userID)).Err()
}

// family key helper
func familyKey(familyID string) string {
	return "refresh_family:" + familyID
}

// user family key helper
func userFamilyKey(userID string) string {
	return "user_refresh_families:" + userID
}
