package repository

import (
	"context"

	"gorm.io/gorm"

	"llm-gateway/backend/internal/entity"
)

type RoutingRepository struct {
	db *gorm.DB
}

type ResolvedRoute struct {
	Model         entity.Model
	ProviderModel entity.ProviderModel
	Provider      entity.Provider
}

func NewRoutingRepository(db *gorm.DB) *RoutingRepository {
	return &RoutingRepository{db: db}
}

func (r *RoutingRepository) ResolveByPublicModel(ctx context.Context, modelName string) (ResolvedRoute, error) {
	var model entity.Model
	err := r.db.WithContext(ctx).
		Preload("ActiveProviderModel.Provider").
		Where("name = ?", modelName).
		First(&model).Error
	if err != nil {
		return ResolvedRoute{}, err
	}

	if model.ActiveProviderModel == nil {
		return ResolvedRoute{Model: model}, nil
	}

	return ResolvedRoute{
		Model:         model,
		ProviderModel: *model.ActiveProviderModel,
		Provider:      model.ActiveProviderModel.Provider,
	}, nil
}
