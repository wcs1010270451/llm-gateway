package repository

import (
	"context"

	"gorm.io/gorm"

	"llm-gateway/backend/internal/entity"
)

type ProviderRepository struct {
	db *gorm.DB
}

func NewProviderRepository(db *gorm.DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

func (r *ProviderRepository) List(ctx context.Context) ([]entity.Provider, error) {
	var items []entity.Provider
	err := r.db.WithContext(ctx).
		Order("id DESC").
		Find(&items).Error
	return items, err
}

func (r *ProviderRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Provider{}).
		Count(&count).Error
	return count, err
}

func (r *ProviderRepository) Get(ctx context.Context, id int64) (entity.Provider, error) {
	var item entity.Provider
	err := r.db.WithContext(ctx).First(&item, id).Error
	return item, err
}

func (r *ProviderRepository) Create(ctx context.Context, item *entity.Provider) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *ProviderRepository) Update(ctx context.Context, item *entity.Provider) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *ProviderRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.Provider{}, id).Error
}

func (r *ProviderRepository) CountProviderModels(ctx context.Context, providerID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.ProviderModel{}).
		Where("provider_id = ?", providerID).
		Count(&count).Error
	return count, err
}
