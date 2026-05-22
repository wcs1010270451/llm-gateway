package service

import (
	"context"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

type RequestLogService struct {
	logs *repository.RequestLogRepository
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
