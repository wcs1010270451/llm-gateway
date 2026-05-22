package service

import (
	"context"
	"strings"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

type ProviderService struct {
	providers *repository.ProviderRepository
	cipher    *ProviderKeyCipher
}

type ProviderInput struct {
	Name            string         `json:"name"`
	Slug            string         `json:"slug"`
	Vendor          string         `json:"vendor"`
	AdapterType     string         `json:"adapter_type"`
	AuthType        string         `json:"auth_type"`
	BaseURL         string         `json:"base_url"`
	APIKeyEncrypted string         `json:"api_key_encrypted"`
	ConfigJSON      map[string]any `json:"config_json"`
	Status          string         `json:"status"`
	Description     string         `json:"description"`
}

func NewProviderService(providers *repository.ProviderRepository, cipher *ProviderKeyCipher) *ProviderService {
	return &ProviderService{providers: providers, cipher: cipher}
}

func (s *ProviderService) List(ctx context.Context) ([]entity.Provider, error) {
	return s.providers.List(ctx)
}

func (s *ProviderService) Get(ctx context.Context, id int64) (entity.Provider, error) {
	return s.providers.Get(ctx, id)
}

func (s *ProviderService) Create(ctx context.Context, input ProviderInput) (entity.Provider, error) {
	item, err := s.buildProvider(input)
	if err != nil {
		return entity.Provider{}, err
	}
	err = s.providers.Create(ctx, &item)
	return item, err
}

func (s *ProviderService) Update(ctx context.Context, id int64, input ProviderInput) (entity.Provider, error) {
	item, err := s.providers.Get(ctx, id)
	if err != nil {
		return entity.Provider{}, err
	}

	next, err := s.buildProvider(input)
	if err != nil {
		return entity.Provider{}, err
	}

	item.Name = next.Name
	item.Slug = next.Slug
	item.Vendor = next.Vendor
	item.AdapterType = next.AdapterType
	item.AuthType = next.AuthType
	item.BaseURL = next.BaseURL
	if next.APIKeyEncrypted != "" {
		item.APIKeyEncrypted = next.APIKeyEncrypted
	}
	item.ConfigJSON = next.ConfigJSON
	item.Status = next.Status
	item.Description = next.Description

	err = s.providers.Update(ctx, &item)
	return item, err
}

func (s *ProviderService) Delete(ctx context.Context, id int64) error {
	count, err := s.providers.CountProviderModels(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return validationError("provider has upstream models and cannot be deleted")
	}
	return s.providers.Delete(ctx, id)
}

func (s *ProviderService) buildProvider(input ProviderInput) (entity.Provider, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(input.Slug)
	input.Vendor = strings.TrimSpace(input.Vendor)
	input.AdapterType = strings.TrimSpace(input.AdapterType)
	input.AuthType = strings.TrimSpace(input.AuthType)
	input.BaseURL = strings.TrimSpace(input.BaseURL)
	input.Status = strings.TrimSpace(input.Status)

	if input.Name == "" {
		return entity.Provider{}, validationError("provider name is required")
	}
	if input.Slug == "" {
		return entity.Provider{}, validationError("provider slug is required")
	}
	if !allowed(input.Vendor, "openai", "anthropic", "google", "custom") {
		return entity.Provider{}, validationError("invalid provider vendor")
	}
	if !allowed(input.AdapterType, "openai_compatible", "anthropic", "claude_code", "gemini", "vertexai") {
		return entity.Provider{}, validationError("invalid provider adapter_type")
	}
	if input.AuthType == "" {
		input.AuthType = "api_key"
	}
	if !allowed(input.AuthType, "api_key", "local_oauth", "adc", "none") {
		return entity.Provider{}, validationError("invalid provider auth_type")
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if !allowed(input.Status, "active", "disabled") {
		return entity.Provider{}, validationError("invalid provider status")
	}

	configJSON, err := normalizeJSON(input.ConfigJSON)
	if err != nil {
		return entity.Provider{}, err
	}
	encryptedKey, err := s.cipher.Encrypt(input.APIKeyEncrypted)
	if err != nil {
		return entity.Provider{}, err
	}

	return entity.Provider{
		Name:            input.Name,
		Slug:            input.Slug,
		Vendor:          input.Vendor,
		AdapterType:     input.AdapterType,
		AuthType:        input.AuthType,
		BaseURL:         input.BaseURL,
		APIKeyEncrypted: encryptedKey,
		ConfigJSON:      configJSON,
		Status:          input.Status,
		Description:     strings.TrimSpace(input.Description),
	}, nil
}
