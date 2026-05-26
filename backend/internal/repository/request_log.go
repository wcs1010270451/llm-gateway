package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"llm-gateway/backend/internal/entity"
)

type RequestLogRepository struct {
	db *gorm.DB
}

func NewRequestLogRepository(db *gorm.DB) *RequestLogRepository {
	return &RequestLogRepository{db: db}
}

func (r *RequestLogRepository) Create(ctx context.Context, item *entity.RequestLog) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *RequestLogRepository) List(ctx context.Context, page int, pageSize int) ([]entity.RequestLog, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&entity.RequestLog{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []entity.RequestLog
	err := query.
		Preload("User").
		Order("id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&items).Error
	return items, total, err
}

func (r *RequestLogRepository) Get(ctx context.Context, id int64) (entity.RequestLog, error) {
	var item entity.RequestLog
	err := r.db.WithContext(ctx).Preload("User").First(&item, id).Error
	return item, err
}

func (r *RequestLogRepository) TotalUsage(ctx context.Context) (int64, int64, error) {
	var row struct {
		RequestCount int64
		TotalTokens  int64
	}
	err := r.db.WithContext(ctx).
		Model(&entity.RequestLog{}).
		Select("COUNT(*) AS request_count, COALESCE(SUM(total_tokens), 0) AS total_tokens").
		Scan(&row).Error
	return row.RequestCount, row.TotalTokens, err
}

func (r *RequestLogRepository) UsageSince(ctx context.Context, since time.Time) (entity.RequestUsageSummary, error) {
	var summary entity.RequestUsageSummary
	err := r.db.WithContext(ctx).
		Model(&entity.RequestLog{}).
		Select(`
			COUNT(*) AS request_count,
			COALESCE(SUM(CASE WHEN success THEN 1 ELSE 0 END), 0) AS success_count,
			COUNT(DISTINCT user_id) AS active_user_count,
			COUNT(DISTINCT api_key_id) AS active_key_count,
			COALESCE(SUM(total_tokens), 0) AS total_tokens,
			COALESCE(AVG(latency_ms), 0) AS average_latency_ms,
			COALESCE(SUM(estimated_cost), 0) AS estimated_cost
		`).
		Where("created_at >= ?", since).
		Scan(&summary).Error
	return summary, err
}

func (r *RequestLogRepository) TopModelsSince(ctx context.Context, since time.Time, limit int) ([]entity.KeyModelUsageStat, error) {
	var items []entity.KeyModelUsageStat
	err := r.db.WithContext(ctx).
		Model(&entity.RequestLog{}).
		Select(`
			public_model_name,
			COUNT(*) AS request_count,
			COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) AS completion_tokens,
			COALESCE(SUM(total_tokens), 0) AS total_tokens,
			COALESCE(SUM(estimated_cost), 0) AS estimated_cost
		`).
		Where("created_at >= ? AND public_model_name <> ''", since).
		Group("public_model_name").
		Order("request_count DESC").
		Limit(limit).
		Scan(&items).Error
	return items, err
}

func (r *RequestLogRepository) ListByUserAPIKey(ctx context.Context, userID int64, apiKeyID int64, page int, pageSize int) ([]entity.RequestLog, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).
		Model(&entity.RequestLog{}).
		Where("user_id = ? AND api_key_id = ?", userID, apiKeyID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []entity.RequestLog
	err := query.
		Order("id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&items).Error
	return items, total, err
}

func (r *RequestLogRepository) GetByUser(ctx context.Context, userID int64, id int64) (entity.RequestLog, error) {
	var item entity.RequestLog
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).
		First(&item).Error
	return item, err
}

func (r *RequestLogRepository) ModelStatsByUserAPIKey(ctx context.Context, userID int64, apiKeyID int64) ([]entity.KeyModelUsageStat, error) {
	var items []entity.KeyModelUsageStat
	err := r.db.WithContext(ctx).
		Model(&entity.RequestLog{}).
		Select(`
			public_model_name,
			COUNT(*) AS request_count,
			COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens,
			COALESCE(SUM(completion_tokens), 0) AS completion_tokens,
			COALESCE(SUM(total_tokens), 0) AS total_tokens,
			COALESCE(SUM(estimated_cost), 0) AS estimated_cost
		`).
		Where("user_id = ? AND api_key_id = ?", userID, apiKeyID).
		Group("public_model_name").
		Order("request_count DESC").
		Scan(&items).Error
	return items, err
}
