package service

import (
	"context"
	"strings"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

type ModelFamilyService struct {
	families *repository.ModelFamilyRepository
}

type ModelFamilyInput struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

func NewModelFamilyService(families *repository.ModelFamilyRepository) *ModelFamilyService {
	return &ModelFamilyService{families: families}
}

func (s *ModelFamilyService) List(ctx context.Context) ([]entity.ModelFamily, error) {
	return s.families.List(ctx)
}

func (s *ModelFamilyService) ListActive(ctx context.Context) ([]entity.ModelFamily, error) {
	return s.families.ListActive(ctx)
}

func (s *ModelFamilyService) Get(ctx context.Context, id int64) (entity.ModelFamily, error) {
	return s.families.Get(ctx, id)
}

func (s *ModelFamilyService) Create(ctx context.Context, input ModelFamilyInput) (entity.ModelFamily, error) {
	item, err := buildModelFamily(input)
	if err != nil {
		return entity.ModelFamily{}, err
	}
	err = s.families.Create(ctx, &item)
	return item, err
}

func (s *ModelFamilyService) Update(ctx context.Context, id int64, input ModelFamilyInput) (entity.ModelFamily, error) {
	item, err := s.families.Get(ctx, id)
	if err != nil {
		return entity.ModelFamily{}, err
	}
	next, err := buildModelFamily(input)
	if err != nil {
		return entity.ModelFamily{}, err
	}

	item.Name = next.Name
	item.DisplayName = next.DisplayName
	item.Status = next.Status
	item.Description = next.Description

	err = s.families.Update(ctx, &item)
	return item, err
}

func (s *ModelFamilyService) Delete(ctx context.Context, id int64) error {
	return s.families.Delete(ctx, id)
}

func buildModelFamily(input ModelFamilyInput) (entity.ModelFamily, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Status = strings.TrimSpace(input.Status)

	if input.Name == "" {
		return entity.ModelFamily{}, validationError("model family name is required")
	}
	if input.Status == "" {
		input.Status = "active"
	}
	if !allowed(input.Status, "active", "disabled") {
		return entity.ModelFamily{}, validationError("invalid model family status")
	}

	return entity.ModelFamily{
		Name:        input.Name,
		DisplayName: input.DisplayName,
		Status:      input.Status,
		Description: strings.TrimSpace(input.Description),
	}, nil
}
