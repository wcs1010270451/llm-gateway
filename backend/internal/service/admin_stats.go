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
	users     *repository.UserRepository
	apiKeys   *repository.APIKeyRepository
}

type AdminStats struct {
	ProviderCount int64                      `json:"provider_count"`
	ModelCount    int64                      `json:"model_count"`
	UserCount     int64                      `json:"user_count"`
	ActiveUsers   int64                      `json:"active_user_count"`
	APIKeyCount   int64                      `json:"api_key_count"`
	RecentUsage   entity.RequestUsageSummary `json:"recent_usage"`
	TopModels     []entity.KeyModelUsageStat `json:"top_models"`
}

func NewAdminStatsService(providers *repository.ProviderRepository, models *repository.ModelRepository, logs *repository.RequestLogRepository, users *repository.UserRepository, apiKeys *repository.APIKeyRepository) *AdminStatsService {
	return &AdminStatsService{
		providers: providers,
		models:    models,
		logs:      logs,
		users:     users,
		apiKeys:   apiKeys,
	}
}

func (s *AdminStatsService) Get(ctx context.Context) (AdminStats, error) {
	providerCount, err := s.providers.Count(ctx)
	if err != nil {
		return AdminStats{}, err
	}
	modelCount, err := s.models.Count(ctx)
	if err != nil {
		return AdminStats{}, err
	}
	userCount, err := s.users.Count(ctx)
	if err != nil {
		return AdminStats{}, err
	}
	activeUsers, err := s.users.CountByStatus(ctx, "active")
	if err != nil {
		return AdminStats{}, err
	}
	apiKeyCount, err := s.apiKeys.Count(ctx)
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
		ProviderCount: providerCount,
		ModelCount:    modelCount,
		UserCount:     userCount,
		ActiveUsers:   activeUsers,
		APIKeyCount:   apiKeyCount,
		RecentUsage:   recentUsage,
		TopModels:     topModels,
	}, nil
}
