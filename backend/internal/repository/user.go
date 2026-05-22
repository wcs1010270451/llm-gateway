package repository

import (
	"context"

	"gorm.io/gorm"

	"llm-gateway/backend/internal/entity"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) List(ctx context.Context) ([]entity.User, error) {
	var items []entity.User
	err := r.db.WithContext(ctx).
		Order("id DESC").
		Find(&items).Error
	return items, err
}

func (r *UserRepository) Get(ctx context.Context, id int64) (entity.User, error) {
	var item entity.User
	err := r.db.WithContext(ctx).First(&item, id).Error
	return item, err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (entity.User, error) {
	var item entity.User
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&item).Error
	return item, err
}

func (r *UserRepository) Create(ctx context.Context, item *entity.User) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *UserRepository) Update(ctx context.Context, item *entity.User) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, id).Error
}
