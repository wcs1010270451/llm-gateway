package repository

import (
	"context"

	"gorm.io/gorm"

	"llm-gateway/backend/internal/entity"
)

type ModelRepository struct {
	db *gorm.DB
}

func NewModelRepository(db *gorm.DB) *ModelRepository {
	return &ModelRepository{db: db}
}

func (r *ModelRepository) List(ctx context.Context) ([]entity.Model, error) {
	var items []entity.Model
	err := r.db.WithContext(ctx).
		Preload("ActiveProviderModel.Provider").
		Order("id DESC").
		Find(&items).Error
	return items, err
}

func (r *ModelRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Model{}).
		Count(&count).Error
	return count, err
}

func (r *ModelRepository) ListEnabled(ctx context.Context) ([]entity.Model, error) {
	var items []entity.Model
	err := r.db.WithContext(ctx).
		Preload("ActiveProviderModel.Provider").
		Where("status = ? AND active_provider_model_id IS NOT NULL", "enabled").
		Order("name ASC").
		Find(&items).Error
	return items, err
}

func (r *ModelRepository) Get(ctx context.Context, id int64) (entity.Model, error) {
	var item entity.Model
	err := r.db.WithContext(ctx).
		Preload("ActiveProviderModel.Provider").
		First(&item, id).Error
	return item, err
}

func (r *ModelRepository) Create(ctx context.Context, item *entity.Model) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *ModelRepository) Update(ctx context.Context, item *entity.Model) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *ModelRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.Model{}, id).Error
}

func (r *ModelRepository) HasActiveProviderModel(ctx context.Context, modelID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Model{}).
		Where("id = ? AND active_provider_model_id IS NOT NULL", modelID).
		Count(&count).Error
	return count > 0, err
}

func (r *ModelRepository) CountProviderModels(ctx context.Context, modelID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.ProviderModel{}).
		Where("model_id = ?", modelID).
		Count(&count).Error
	return count, err
}

func (r *ModelRepository) ListProviderModels(ctx context.Context, modelID int64) ([]entity.ProviderModel, error) {
	var items []entity.ProviderModel
	err := r.db.WithContext(ctx).
		Preload("Provider").
		Where("model_id = ?", modelID).
		Order("id DESC").
		Find(&items).Error
	return items, err
}

func (r *ModelRepository) ListProviderModelsByProvider(ctx context.Context, providerID int64) ([]entity.ProviderModel, error) {
	var items []entity.ProviderModel
	err := r.db.WithContext(ctx).
		Preload("Provider").
		Preload("Model").
		Where("provider_id = ?", providerID).
		Order("id DESC").
		Find(&items).Error
	return items, err
}

func (r *ModelRepository) GetProviderModel(ctx context.Context, id int64) (entity.ProviderModel, error) {
	var item entity.ProviderModel
	err := r.db.WithContext(ctx).
		Preload("Provider").
		Preload("Model").
		First(&item, id).Error
	return item, err
}

func (r *ModelRepository) CreateProviderModel(ctx context.Context, item *entity.ProviderModel) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *ModelRepository) UpdateProviderModel(ctx context.Context, item *entity.ProviderModel) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *ModelRepository) ClearActiveProviderModel(ctx context.Context, providerModelID int64) error {
	return r.db.WithContext(ctx).
		Model(&entity.Model{}).
		Where("active_provider_model_id = ?", providerModelID).
		Updates(map[string]any{
			"active_provider_model_id": nil,
			"status":                   "disabled",
		}).Error
}

func (r *ModelRepository) DeleteProviderModel(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.Model{}).
			Where("active_provider_model_id = ?", id).
			Updates(map[string]any{
				"active_provider_model_id": nil,
				"status":                   "disabled",
			}).Error; err != nil {
			return err
		}

		return tx.Delete(&entity.ProviderModel{}, id).Error
	})
}

func (r *ModelRepository) SetActiveProviderModel(ctx context.Context, modelID int64, providerModelID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&entity.ProviderModel{}).
			Where("id = ? AND model_id = ? AND status = ?", providerModelID, modelID, "enabled").
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return gorm.ErrRecordNotFound
		}

		return tx.Model(&entity.Model{}).
			Where("id = ?", modelID).
			Updates(map[string]any{
				"active_provider_model_id": providerModelID,
				"status":                   "enabled",
			}).Error
	})
}
