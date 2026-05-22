package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"llm-gateway/backend/internal/entity"
)

type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) List(ctx context.Context) ([]entity.APIKey, error) {
	var items []entity.APIKey
	err := r.db.WithContext(ctx).
		Preload("User").
		Order("id DESC").
		Find(&items).Error
	return items, err
}

func (r *APIKeyRepository) ListByUser(ctx context.Context, userID int64) ([]entity.APIKey, error) {
	var items []entity.APIKey
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		Find(&items).Error
	return items, err
}

func (r *APIKeyRepository) Get(ctx context.Context, id int64) (entity.APIKey, error) {
	var item entity.APIKey
	err := r.db.WithContext(ctx).
		Preload("User").
		First(&item, id).Error
	return item, err
}

func (r *APIKeyRepository) GetByUser(ctx context.Context, userID int64, id int64) (entity.APIKey, error) {
	var item entity.APIKey
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).
		First(&item).Error
	return item, err
}

func (r *APIKeyRepository) Create(ctx context.Context, item *entity.APIKey) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *APIKeyRepository) Update(ctx context.Context, item *entity.APIKey) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *APIKeyRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.APIKey{}, id).Error
}

func (r *APIKeyRepository) FindByKeyHash(ctx context.Context, keyHash string) (entity.APIKey, error) {
	var item entity.APIKey
	err := r.db.WithContext(ctx).
		Preload("User").
		Where("key_hash = ?", keyHash).
		First(&item).Error
	return item, err
}

func (r *APIKeyRepository) MarkUsed(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entity.APIKey{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"last_used_at":       &now,
			"last_error_at":      nil,
			"last_error_message": "",
		}).Error
}
