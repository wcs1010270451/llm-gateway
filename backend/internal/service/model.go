package service

import (
	"context"
	"strings"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

type ModelService struct {
	models   *repository.ModelRepository
	families *repository.ModelFamilyRepository
}

type ModelInput struct {
	Name        string         `json:"name"`
	DisplayName string         `json:"display_name"`
	Family      string         `json:"family"`
	Modality    string         `json:"modality"`
	Status      string         `json:"status"`
	Description string         `json:"description"`
	PricingJSON map[string]any `json:"pricing_json"`
	ConfigJSON  map[string]any `json:"config_json"`
}

type ProviderModelInput struct {
	ProviderID      int64          `json:"provider_id"`
	ModelID         *int64         `json:"model_id"`
	UpstreamModel   string         `json:"upstream_model"`
	Status          string         `json:"status"`
	MaxTokens       int            `json:"max_tokens"`
	TimeoutSeconds  int            `json:"timeout_seconds"`
	InputCostPer1M  float64        `json:"input_cost_per_1m"`
	OutputCostPer1M float64        `json:"output_cost_per_1m"`
	PricingJSON     map[string]any `json:"pricing_json"`
	ConfigJSON      map[string]any `json:"config_json"`
	SetActive       bool           `json:"set_active"`
}

func NewModelService(models *repository.ModelRepository, families *repository.ModelFamilyRepository) *ModelService {
	return &ModelService{models: models, families: families}
}

func (s *ModelService) List(ctx context.Context) ([]entity.Model, error) {
	return s.models.List(ctx)
}

func (s *ModelService) Get(ctx context.Context, id int64) (entity.Model, error) {
	return s.models.Get(ctx, id)
}

func (s *ModelService) Create(ctx context.Context, input ModelInput) (entity.Model, error) {
	item, err := s.buildModel(ctx, input)
	if err != nil {
		return entity.Model{}, err
	}
	item.Status = "disabled"
	err = s.models.Create(ctx, &item)
	return item, err
}

func (s *ModelService) Update(ctx context.Context, id int64, input ModelInput) (entity.Model, error) {
	item, err := s.models.Get(ctx, id)
	if err != nil {
		return entity.Model{}, err
	}

	next, err := s.buildModel(ctx, input)
	if err != nil {
		return entity.Model{}, err
	}
	if next.Status == "enabled" {
		hasActiveProviderModel, err := s.models.HasActiveProviderModel(ctx, id)
		if err != nil {
			return entity.Model{}, err
		}
		if !hasActiveProviderModel {
			return entity.Model{}, validationError("model must have an active provider model before it can be enabled")
		}
	}

	item.Name = next.Name
	item.DisplayName = next.DisplayName
	item.Family = next.Family
	item.Modality = next.Modality
	item.Status = next.Status
	item.Description = next.Description
	item.PricingJSON = next.PricingJSON
	item.ConfigJSON = next.ConfigJSON

	err = s.models.Update(ctx, &item)
	return item, err
}

func (s *ModelService) Delete(ctx context.Context, id int64) error {
	count, err := s.models.CountProviderModels(ctx, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return validationError("model has upstream model mappings and cannot be deleted")
	}
	return s.models.Delete(ctx, id)
}

func (s *ModelService) ListProviderModels(ctx context.Context, modelID int64) ([]entity.ProviderModel, error) {
	return s.models.ListProviderModels(ctx, modelID)
}

func (s *ModelService) ListProviderModelsByProvider(ctx context.Context, providerID int64) ([]entity.ProviderModel, error) {
	return s.models.ListProviderModelsByProvider(ctx, providerID)
}

func (s *ModelService) CreateProviderModel(ctx context.Context, modelID int64, input ProviderModelInput) (entity.ProviderModel, error) {
	if _, err := s.models.Get(ctx, modelID); err != nil {
		return entity.ProviderModel{}, err
	}

	item, err := buildProviderModel(&modelID, input)
	if err != nil {
		return entity.ProviderModel{}, err
	}
	if err := s.models.CreateProviderModel(ctx, &item); err != nil {
		return entity.ProviderModel{}, err
	}
	if input.SetActive {
		if err := s.models.SetActiveProviderModel(ctx, modelID, item.ID); err != nil {
			return entity.ProviderModel{}, err
		}
	}

	return s.models.GetProviderModel(ctx, item.ID)
}

func (s *ModelService) UpdateProviderModel(ctx context.Context, modelID int64, providerModelID int64, input ProviderModelInput) (entity.ProviderModel, error) {
	item, err := s.models.GetProviderModel(ctx, providerModelID)
	if err != nil {
		return entity.ProviderModel{}, err
	}
	if item.ModelID == nil || *item.ModelID != modelID {
		return entity.ProviderModel{}, validationError("provider model does not belong to model")
	}

	next, err := buildProviderModel(&modelID, input)
	if err != nil {
		return entity.ProviderModel{}, err
	}

	item.ProviderID = next.ProviderID
	item.ModelID = next.ModelID
	item.UpstreamModel = next.UpstreamModel
	item.Status = next.Status
	item.MaxTokens = next.MaxTokens
	item.TimeoutSeconds = next.TimeoutSeconds
	item.InputCostPer1M = next.InputCostPer1M
	item.OutputCostPer1M = next.OutputCostPer1M
	item.PricingJSON = next.PricingJSON
	item.ConfigJSON = next.ConfigJSON

	if err := s.models.UpdateProviderModel(ctx, &item); err != nil {
		return entity.ProviderModel{}, err
	}
	if input.SetActive {
		if err := s.models.SetActiveProviderModel(ctx, modelID, item.ID); err != nil {
			return entity.ProviderModel{}, err
		}
	}

	return s.models.GetProviderModel(ctx, item.ID)
}

func (s *ModelService) CreateProviderModelForProvider(ctx context.Context, providerID int64, input ProviderModelInput) (entity.ProviderModel, error) {
	input.ProviderID = providerID
	if input.ModelID != nil {
		if _, err := s.models.Get(ctx, *input.ModelID); err != nil {
			return entity.ProviderModel{}, err
		}
	}

	item, err := buildProviderModel(input.ModelID, input)
	if err != nil {
		return entity.ProviderModel{}, err
	}
	if err := s.models.CreateProviderModel(ctx, &item); err != nil {
		return entity.ProviderModel{}, err
	}
	if input.SetActive {
		if item.ModelID == nil {
			return entity.ProviderModel{}, validationError("provider model must be linked to a model before it can be active")
		}
		if err := s.models.SetActiveProviderModel(ctx, *item.ModelID, item.ID); err != nil {
			return entity.ProviderModel{}, err
		}
	}

	return s.models.GetProviderModel(ctx, item.ID)
}

func (s *ModelService) UpdateProviderModelForProvider(ctx context.Context, providerID int64, providerModelID int64, input ProviderModelInput) (entity.ProviderModel, error) {
	item, err := s.models.GetProviderModel(ctx, providerModelID)
	if err != nil {
		return entity.ProviderModel{}, err
	}
	if item.ProviderID != providerID {
		return entity.ProviderModel{}, validationError("provider model does not belong to provider")
	}

	input.ProviderID = providerID
	if input.ModelID != nil {
		if _, err := s.models.Get(ctx, *input.ModelID); err != nil {
			return entity.ProviderModel{}, err
		}
	}

	next, err := buildProviderModel(input.ModelID, input)
	if err != nil {
		return entity.ProviderModel{}, err
	}
	if modelLinkChanged(item.ModelID, next.ModelID) {
		if err := s.models.ClearActiveProviderModel(ctx, item.ID); err != nil {
			return entity.ProviderModel{}, err
		}
	}

	item.ProviderID = next.ProviderID
	item.ModelID = next.ModelID
	item.UpstreamModel = next.UpstreamModel
	item.Status = next.Status
	item.MaxTokens = next.MaxTokens
	item.TimeoutSeconds = next.TimeoutSeconds
	item.InputCostPer1M = next.InputCostPer1M
	item.OutputCostPer1M = next.OutputCostPer1M
	item.PricingJSON = next.PricingJSON
	item.ConfigJSON = next.ConfigJSON

	if err := s.models.UpdateProviderModel(ctx, &item); err != nil {
		return entity.ProviderModel{}, err
	}
	if input.SetActive {
		if item.ModelID == nil {
			return entity.ProviderModel{}, validationError("provider model must be linked to a model before it can be active")
		}
		if err := s.models.SetActiveProviderModel(ctx, *item.ModelID, item.ID); err != nil {
			return entity.ProviderModel{}, err
		}
	}

	return s.models.GetProviderModel(ctx, item.ID)
}

func (s *ModelService) DeleteProviderModelForProvider(ctx context.Context, providerID int64, providerModelID int64) error {
	item, err := s.models.GetProviderModel(ctx, providerModelID)
	if err != nil {
		return err
	}
	if item.ProviderID != providerID {
		return validationError("provider model does not belong to provider")
	}
	return s.models.DeleteProviderModel(ctx, providerModelID)
}

func (s *ModelService) DeleteProviderModel(ctx context.Context, modelID int64, providerModelID int64) error {
	item, err := s.models.GetProviderModel(ctx, providerModelID)
	if err != nil {
		return err
	}
	if item.ModelID == nil || *item.ModelID != modelID {
		return validationError("provider model does not belong to model")
	}
	return s.models.DeleteProviderModel(ctx, providerModelID)
}

func (s *ModelService) SetActiveProviderModel(ctx context.Context, modelID int64, providerModelID int64) (entity.Model, error) {
	if err := s.models.SetActiveProviderModel(ctx, modelID, providerModelID); err != nil {
		return entity.Model{}, err
	}
	return s.models.Get(ctx, modelID)
}

func (s *ModelService) buildModel(ctx context.Context, input ModelInput) (entity.Model, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Family = strings.TrimSpace(input.Family)
	input.Modality = strings.TrimSpace(input.Modality)
	input.Status = strings.TrimSpace(input.Status)

	if input.Name == "" {
		return entity.Model{}, validationError("model name is required")
	}
	if input.Family == "" {
		return entity.Model{}, validationError("model family is required")
	}
	if _, err := s.families.FindActiveByName(ctx, input.Family); err != nil {
		return entity.Model{}, err
	}
	if input.Modality == "" {
		input.Modality = "text"
	}
	if !allowed(input.Modality, "text", "vision", "embedding", "multimodal") {
		return entity.Model{}, validationError("invalid model modality")
	}
	if input.Status == "" {
		input.Status = "enabled"
	}
	if !allowed(input.Status, "enabled", "disabled") {
		return entity.Model{}, validationError("invalid model status")
	}

	configJSON, err := normalizeJSON(input.ConfigJSON)
	if err != nil {
		return entity.Model{}, err
	}
	pricingJSON, err := normalizeJSON(input.PricingJSON)
	if err != nil {
		return entity.Model{}, err
	}

	return entity.Model{
		Name:        input.Name,
		DisplayName: input.DisplayName,
		Family:      input.Family,
		Modality:    input.Modality,
		Status:      input.Status,
		Description: strings.TrimSpace(input.Description),
		PricingJSON: pricingJSON,
		ConfigJSON:  configJSON,
	}, nil
}

func buildProviderModel(modelID *int64, input ProviderModelInput) (entity.ProviderModel, error) {
	input.UpstreamModel = strings.TrimSpace(input.UpstreamModel)
	input.Status = strings.TrimSpace(input.Status)

	if input.ProviderID <= 0 {
		return entity.ProviderModel{}, validationError("provider is required")
	}
	if input.UpstreamModel == "" {
		return entity.ProviderModel{}, validationError("upstream model is required")
	}
	if input.Status == "" {
		input.Status = "enabled"
	}
	if !allowed(input.Status, "enabled", "disabled") {
		return entity.ProviderModel{}, validationError("invalid provider model status")
	}
	if input.MaxTokens < 0 {
		return entity.ProviderModel{}, validationError("max_tokens must be greater than or equal to 0")
	}
	if input.TimeoutSeconds <= 0 {
		input.TimeoutSeconds = 300
	}
	if input.InputCostPer1M < 0 || input.OutputCostPer1M < 0 {
		return entity.ProviderModel{}, validationError("cost must be greater than or equal to 0")
	}
	if input.SetActive && input.Status != "enabled" {
		return entity.ProviderModel{}, validationError("only enabled provider model can be active")
	}
	if input.SetActive && modelID == nil {
		return entity.ProviderModel{}, validationError("provider model must be linked to a model before it can be active")
	}

	configJSON, err := normalizeJSON(input.ConfigJSON)
	if err != nil {
		return entity.ProviderModel{}, err
	}
	pricingJSON, err := normalizeJSON(input.PricingJSON)
	if err != nil {
		return entity.ProviderModel{}, err
	}

	return entity.ProviderModel{
		ProviderID:      input.ProviderID,
		ModelID:         modelID,
		UpstreamModel:   input.UpstreamModel,
		Status:          input.Status,
		MaxTokens:       input.MaxTokens,
		TimeoutSeconds:  input.TimeoutSeconds,
		InputCostPer1M:  input.InputCostPer1M,
		OutputCostPer1M: input.OutputCostPer1M,
		PricingJSON:     pricingJSON,
		ConfigJSON:      configJSON,
	}, nil
}

func modelLinkChanged(left *int64, right *int64) bool {
	if left == nil || right == nil {
		return left != right
	}
	return *left != *right
}
