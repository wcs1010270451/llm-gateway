package repository

import (
	"context"

	"gorm.io/gorm"

	"llm-gateway/backend/internal/entity"
)

type ModelFamilyRepository struct {
	db *gorm.DB
}

func NewModelFamilyRepository(db *gorm.DB) *ModelFamilyRepository {
	return &ModelFamilyRepository{db: db}
}

func (r *ModelFamilyRepository) List(ctx context.Context) ([]entity.ModelFamily, error) {
	var items []entity.ModelFamily
	err := r.db.WithContext(ctx).
		Order("id DESC").
		Find(&items).Error
	return items, err
}

func (r *ModelFamilyRepository) ListActive(ctx context.Context) ([]entity.ModelFamily, error) {
	var items []entity.ModelFamily
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("name ASC").
		Find(&items).Error
	return items, err
}

func (r *ModelFamilyRepository) Get(ctx context.Context, id int64) (entity.ModelFamily, error) {
	var item entity.ModelFamily
	err := r.db.WithContext(ctx).First(&item, id).Error
	return item, err
}

func (r *ModelFamilyRepository) FindActiveByName(ctx context.Context, name string) (entity.ModelFamily, error) {
	var item entity.ModelFamily
	err := r.db.WithContext(ctx).
		Where("name = ? AND status = ?", name, "active").
		First(&item).Error
	return item, err
}

func (r *ModelFamilyRepository) Create(ctx context.Context, item *entity.ModelFamily) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *ModelFamilyRepository) Update(ctx context.Context, item *entity.ModelFamily) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *ModelFamilyRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.ModelFamily{}, id).Error
}
