package service

import (
	"context"
	"time"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

type AdminStatsService struct {
	providers *repository.ProviderRepository
	models    *repository.ModelRepository
	logs      *repository.RequestLogRepository
}

type AdminStats struct {
	RequestCount  int64                      `json:"request_count"`
	TotalTokens   int64                      `json:"total_tokens"`
	ProviderCount int64                      `json:"provider_count"`
	ModelCount    int64                      `json:"model_count"`
	RecentUsage   entity.RequestUsageSummary `json:"recent_usage"`
	TopModels     []entity.KeyModelUsageStat `json:"top_models"`
}

func NewAdminStatsService(providers *repository.ProviderRepository, models *repository.ModelRepository, logs *repository.RequestLogRepository) *AdminStatsService {
	return &AdminStatsService{
		providers: providers,
		models:    models,
		logs:      logs,
	}
}

func (s *AdminStatsService) Get(ctx context.Context) (AdminStats, error) {
	requestCount, totalTokens, err := s.logs.TotalUsage(ctx)
	if err != nil {
		return AdminStats{}, err
	}
	providerCount, err := s.providers.Count(ctx)
	if err != nil {
		return AdminStats{}, err
	}
	modelCount, err := s.models.Count(ctx)
	if err != nil {
		return AdminStats{}, err
	}
	since := time.Now().Add(-24 * time.Hour)
	recentUsage, err := s.logs.UsageSince(ctx, since)
	if err != nil {
		return AdminStats{}, err
	}
	topModels, err := s.logs.TopModelsSince(ctx, since, 5)
	if err != nil {
		return AdminStats{}, err
	}

	return AdminStats{
		RequestCount:  requestCount,
		TotalTokens:   totalTokens,
		ProviderCount: providerCount,
		ModelCount:    modelCount,
		RecentUsage:   recentUsage,
		TopModels:     topModels,
	}, nil
}
