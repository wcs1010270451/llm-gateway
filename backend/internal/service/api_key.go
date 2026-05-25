package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

type APIKeyService struct {
	keys   *repository.APIKeyRepository
	users  *repository.UserRepository
	cipher *ProviderKeyCipher
}

type APIKeyInput struct {
	UserID            int64      `json:"user_id"`
	Name              string     `json:"name"`
	Status            string     `json:"status"`
	RPMLimit          int        `json:"rpm_limit"`
	DailyRequestLimit int        `json:"daily_request_limit"`
	DailyTokenLimit   int        `json:"daily_token_limit"`
	ExpiresAt         *time.Time `json:"expires_at"`
}

type CreatedAPIKey struct {
	APIKey   entity.APIKey `json:"api_key"`
	PlainKey string        `json:"plain_key"`
}

type RevealedAPIKey struct {
	PlainKey string `json:"plain_key"`
}

type APIAuthContext struct {
	APIKey entity.APIKey
	User   entity.User
}

func NewAPIKeyService(keys *repository.APIKeyRepository, users *repository.UserRepository, cipher *ProviderKeyCipher) *APIKeyService {
	return &APIKeyService{keys: keys, users: users, cipher: cipher}
}

func (s *APIKeyService) List(ctx context.Context) ([]entity.APIKey, error) {
	return s.keys.List(ctx)
}

func (s *APIKeyService) ListForUser(ctx context.Context, user entity.User) ([]entity.APIKey, error) {
	if user.Role != "user" {
		return nil, validationError("admin users cannot own api keys")
	}
	return s.keys.ListByUser(ctx, user.ID)
}

func (s *APIKeyService) Get(ctx context.Context, id int64) (entity.APIKey, error) {
	return s.keys.Get(ctx, id)
}

func (s *APIKeyService) Create(ctx context.Context, input APIKeyInput) (CreatedAPIKey, error) {
	item, err := s.buildAPIKey(ctx, input)
	if err != nil {
		return CreatedAPIKey{}, err
	}

	rawKey, err := generateAPIKey()
	if err != nil {
		return CreatedAPIKey{}, err
	}
	item.KeyHash = HashAPIKey(rawKey)
	item.KeyEncrypted, err = s.cipher.Encrypt(rawKey)
	if err != nil {
		return CreatedAPIKey{}, err
	}
	item.MaskedKey = maskAPIKey(rawKey)

	if err := s.keys.Create(ctx, &item); err != nil {
		return CreatedAPIKey{}, err
	}
	created, err := s.keys.Get(ctx, item.ID)
	if err != nil {
		return CreatedAPIKey{}, err
	}

	return CreatedAPIKey{APIKey: created, PlainKey: rawKey}, nil
}

func (s *APIKeyService) CreateForUser(ctx context.Context, user entity.User, input APIKeyInput) (CreatedAPIKey, error) {
	if user.Role != "user" {
		return CreatedAPIKey{}, validationError("admin users cannot create api keys")
	}
	input.UserID = user.ID
	return s.Create(ctx, input)
}

func (s *APIKeyService) Update(ctx context.Context, id int64, input APIKeyInput) (entity.APIKey, error) {
	item, err := s.keys.Get(ctx, id)
	if err != nil {
		return entity.APIKey{}, err
	}
	next, err := s.buildAPIKey(ctx, input)
	if err != nil {
		return entity.APIKey{}, err
	}

	item.UserID = next.UserID
	item.Name = next.Name
	item.Status = next.Status
	item.RPMLimit = next.RPMLimit
	item.DailyRequestLimit = next.DailyRequestLimit
	item.DailyTokenLimit = next.DailyTokenLimit
	item.ExpiresAt = next.ExpiresAt

	if err := s.keys.Update(ctx, &item); err != nil {
		return entity.APIKey{}, err
	}
	return s.keys.Get(ctx, item.ID)
}

func (s *APIKeyService) UpdateForUser(ctx context.Context, user entity.User, id int64, input APIKeyInput) (entity.APIKey, error) {
	if user.Role != "user" {
		return entity.APIKey{}, validationError("admin users cannot update api keys")
	}
	if _, err := s.keys.GetByUser(ctx, user.ID, id); err != nil {
		return entity.APIKey{}, err
	}
	input.UserID = user.ID
	return s.Update(ctx, id, input)
}

func (s *APIKeyService) Delete(ctx context.Context, id int64) error {
	return s.keys.Delete(ctx, id)
}

func (s *APIKeyService) DeleteForUser(ctx context.Context, user entity.User, id int64) error {
	if user.Role != "user" {
		return validationError("admin users cannot delete api keys")
	}
	if _, err := s.keys.GetByUser(ctx, user.ID, id); err != nil {
		return err
	}
	return s.keys.Delete(ctx, id)
}

func (s *APIKeyService) RevealForUser(ctx context.Context, user entity.User, id int64) (RevealedAPIKey, error) {
	if user.Role != "user" {
		return RevealedAPIKey{}, validationError("admin users cannot reveal api keys")
	}

	item, err := s.keys.GetByUser(ctx, user.ID, id)
	if err != nil {
		return RevealedAPIKey{}, err
	}
	if strings.TrimSpace(item.KeyEncrypted) == "" {
		return RevealedAPIKey{}, validationError("full key is unavailable; create a new key to enable copying")
	}

	plainKey, err := s.cipher.Decrypt(item.KeyEncrypted)
	if err != nil {
		return RevealedAPIKey{}, err
	}

	encryptedKey, err := s.cipher.Encrypt(item.KeyEncrypted)
	if err != nil {
		return RevealedAPIKey{}, err
	}
	if encryptedKey != item.KeyEncrypted {
		item.KeyEncrypted = encryptedKey
		if err := s.keys.Update(ctx, &item); err != nil {
			return RevealedAPIKey{}, err
		}
	}

	return RevealedAPIKey{PlainKey: plainKey}, nil
}

func (s *APIKeyService) Authenticate(ctx context.Context, rawKey string) (APIAuthContext, error) {
	rawKey = strings.TrimSpace(rawKey)
	if rawKey == "" {
		return APIAuthContext{}, validationError("api key is required")
	}

	item, err := s.keys.FindByKeyHash(ctx, HashAPIKey(rawKey))
	if err != nil {
		return APIAuthContext{}, err
	}
	if item.Status != "active" {
		return APIAuthContext{}, validationError("api key is disabled")
	}
	if item.User.Status != "active" {
		return APIAuthContext{}, validationError("user is disabled")
	}
	if item.ExpiresAt != nil && time.Now().After(*item.ExpiresAt) {
		return APIAuthContext{}, validationError("api key is expired")
	}
	if err := s.keys.MarkUsed(ctx, item.ID); err != nil {
		return APIAuthContext{}, err
	}

	return APIAuthContext{APIKey: item, User: item.User}, nil
}

func (s *APIKeyService) buildAPIKey(ctx context.Context, input APIKeyInput) (entity.APIKey, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Status = strings.TrimSpace(input.Status)

	if input.UserID <= 0 {
		return entity.APIKey{}, validationError("user is required")
	}
	if _, err := s.users.Get(ctx, input.UserID); err != nil {
		return entity.APIKey{}, err
	}
	if input.Name == "" {
		return entity.APIKey{}, validationError("api key name is required")
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if !allowed(input.Status, "active", "disabled") {
		return entity.APIKey{}, validationError("invalid api key status")
	}
	if input.RPMLimit < 0 || input.DailyRequestLimit < 0 || input.DailyTokenLimit < 0 {
		return entity.APIKey{}, validationError("limits must be greater than or equal to 0")
	}

	return entity.APIKey{
		UserID:            input.UserID,
		Name:              input.Name,
		Status:            input.Status,
		RPMLimit:          input.RPMLimit,
		DailyRequestLimit: input.DailyRequestLimit,
		DailyTokenLimit:   input.DailyTokenLimit,
		ExpiresAt:         input.ExpiresAt,
	}, nil
}

func HashAPIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

func generateAPIKey() (string, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return "lgw_" + hex.EncodeToString(random), nil
}

func maskAPIKey(rawKey string) string {
	if len(rawKey) <= 12 {
		return rawKey
	}
	return rawKey[:8] + "..." + rawKey[len(rawKey)-6:]
}
