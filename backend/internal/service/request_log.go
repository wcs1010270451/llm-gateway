package service

import (
	"context"
	"time"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

type RequestLogService struct {
	logs *repository.RequestLogRepository
}

type ProviderUsageStats struct {
	Window      string                          `json:"window"`
	Granularity string                          `json:"granularity"`
	Since       time.Time                       `json:"since"`
	Summary     entity.RequestUsageSummary      `json:"summary"`
	Trend       []entity.ProviderUsagePoint     `json:"trend"`
	Models      []entity.ProviderModelUsageStat `json:"models"`
}

func NewRequestLogService(logs *repository.RequestLogRepository) *RequestLogService {
	return &RequestLogService{logs: logs}
}

func (s *RequestLogService) List(ctx context.Context, page int, pageSize int) (RequestLogPage, error) {
	page, pageSize = normalizePagination(page, pageSize)
	items, total, err := s.logs.List(ctx, page, pageSize)
	if err != nil {
		return RequestLogPage{}, err
	}
	return RequestLogPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *RequestLogService) Get(ctx context.Context, id int64) (entity.RequestLog, error) {
	return s.logs.Get(ctx, id)
}

func (s *RequestLogService) ProviderUsageStats(ctx context.Context, providerID int64, window string, granularity string) (ProviderUsageStats, error) {
	since, normalizedWindow, normalizedGranularity := resolveProviderStatsRange(window, granularity)

	summary, err := s.logs.ProviderUsageSummary(ctx, providerID, since)
	if err != nil {
		return ProviderUsageStats{}, err
	}
	trend, err := s.logs.ProviderUsageTrend(ctx, providerID, since, normalizedGranularity)
	if err != nil {
		return ProviderUsageStats{}, err
	}
	models, err := s.logs.ProviderModelUsageStats(ctx, providerID, since)
	if err != nil {
		return ProviderUsageStats{}, err
	}

	return ProviderUsageStats{
		Window:      normalizedWindow,
		Granularity: normalizedGranularity,
		Since:       since,
		Summary:     summary,
		Trend:       trend,
		Models:      models,
	}, nil
}

func resolveProviderStatsRange(window string, granularity string) (time.Time, string, string) {
	now := time.Now().UTC()
	duration := 24 * time.Hour
	normalizedWindow := "24h"

	switch window {
	case "7d":
		duration = 7 * 24 * time.Hour
		normalizedWindow = "7d"
	case "30d":
		duration = 30 * 24 * time.Hour
		normalizedWindow = "30d"
	}

	normalizedGranularity := granularity
	if normalizedGranularity != "hour" && normalizedGranularity != "day" {
		if normalizedWindow == "24h" {
			normalizedGranularity = "hour"
		} else {
			normalizedGranularity = "day"
		}
	}

	return now.Add(-duration), normalizedWindow, normalizedGranularity
}
