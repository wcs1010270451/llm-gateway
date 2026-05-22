package service

import (
	"context"
	"net/http"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

type UserConsoleService struct {
	keys           *repository.APIKeyRepository
	models         *repository.ModelRepository
	logs           *repository.RequestLogRepository
	anthropicProxy *AnthropicProxyService
	openAIProxy    *OpenAIProxyService
	geminiProxy    *GeminiProxyService
}

type RequestLogPage struct {
	Items    []entity.RequestLog `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

func NewUserConsoleService(keys *repository.APIKeyRepository, models *repository.ModelRepository, logs *repository.RequestLogRepository, anthropicProxy *AnthropicProxyService, openAIProxy *OpenAIProxyService, geminiProxy *GeminiProxyService) *UserConsoleService {
	return &UserConsoleService{
		keys:           keys,
		models:         models,
		logs:           logs,
		anthropicProxy: anthropicProxy,
		openAIProxy:    openAIProxy,
		geminiProxy:    geminiProxy,
	}
}

func (s *UserConsoleService) ListModels(ctx context.Context, user entity.User) ([]entity.Model, error) {
	if err := requireNormalUser(user); err != nil {
		return nil, err
	}
	return s.models.ListEnabled(ctx)
}

func (s *UserConsoleService) GetAPIKey(ctx context.Context, user entity.User, id int64) (entity.APIKey, error) {
	if err := requireNormalUser(user); err != nil {
		return entity.APIKey{}, err
	}
	return s.keys.GetByUser(ctx, user.ID, id)
}

func (s *UserConsoleService) ListAPIKeyModelStats(ctx context.Context, user entity.User, apiKeyID int64) ([]entity.KeyModelUsageStat, error) {
	if _, err := s.GetAPIKey(ctx, user, apiKeyID); err != nil {
		return nil, err
	}
	return s.logs.ModelStatsByUserAPIKey(ctx, user.ID, apiKeyID)
}

func (s *UserConsoleService) ListAPIKeyLogs(ctx context.Context, user entity.User, apiKeyID int64, page int, pageSize int) (RequestLogPage, error) {
	if _, err := s.GetAPIKey(ctx, user, apiKeyID); err != nil {
		return RequestLogPage{}, err
	}
	page, pageSize = normalizePagination(page, pageSize)
	items, total, err := s.logs.ListByUserAPIKey(ctx, user.ID, apiKeyID, page, pageSize)
	if err != nil {
		return RequestLogPage{}, err
	}
	return RequestLogPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *UserConsoleService) GetAPIKeyLog(ctx context.Context, user entity.User, logID int64) (entity.RequestLog, error) {
	if err := requireNormalUser(user); err != nil {
		return entity.RequestLog{}, err
	}
	return s.logs.GetByUser(ctx, user.ID, logID)
}

func (s *UserConsoleService) OpenDebugAnthropicMessages(ctx context.Context, user entity.User, apiKeyID int64, input AnthropicMessagesInput) (*AnthropicMessagesResult, error) {
	key, err := s.GetAPIKey(ctx, user, apiKeyID)
	if err != nil {
		return nil, err
	}
	if key.Status != "active" {
		return nil, validationError("api key is disabled")
	}
	if err := s.keys.MarkUsed(ctx, key.ID); err != nil {
		return nil, err
	}

	input.Method = http.MethodPost
	input.UserID = user.ID
	input.APIKeyID = key.ID
	return s.anthropicProxy.OpenMessages(ctx, input)
}

func (s *UserConsoleService) OpenDebugOpenAIChatCompletions(ctx context.Context, user entity.User, apiKeyID int64, input OpenAIChatCompletionsInput) (*OpenAIChatCompletionsResult, error) {
	key, err := s.GetAPIKey(ctx, user, apiKeyID)
	if err != nil {
		return nil, err
	}
	if key.Status != "active" {
		return nil, validationError("api key is disabled")
	}
	if err := s.keys.MarkUsed(ctx, key.ID); err != nil {
		return nil, err
	}

	input.Method = http.MethodPost
	input.UserID = user.ID
	input.APIKeyID = key.ID
	return s.openAIProxy.OpenChatCompletions(ctx, input)
}

func (s *UserConsoleService) OpenDebugGeminiGenerateContent(ctx context.Context, user entity.User, apiKeyID int64, input GeminiGenerateContentInput) (*GeminiGenerateContentResult, error) {
	key, err := s.GetAPIKey(ctx, user, apiKeyID)
	if err != nil {
		return nil, err
	}
	if key.Status != "active" {
		return nil, validationError("api key is disabled")
	}
	if err := s.keys.MarkUsed(ctx, key.ID); err != nil {
		return nil, err
	}

	input.Method = http.MethodPost
	input.UserID = user.ID
	input.APIKeyID = key.ID
	return s.geminiProxy.OpenGenerateContent(ctx, input)
}

func (s *UserConsoleService) OpenDebugGeminiStreamGenerateContent(ctx context.Context, user entity.User, apiKeyID int64, input GeminiGenerateContentInput) (*GeminiGenerateContentResult, error) {
	key, err := s.GetAPIKey(ctx, user, apiKeyID)
	if err != nil {
		return nil, err
	}
	if key.Status != "active" {
		return nil, validationError("api key is disabled")
	}
	if err := s.keys.MarkUsed(ctx, key.ID); err != nil {
		return nil, err
	}

	input.Method = http.MethodPost
	input.UserID = user.ID
	input.APIKeyID = key.ID
	return s.geminiProxy.OpenStreamGenerateContent(ctx, input)
}

func (s *UserConsoleService) LogMessagesResult(ctx context.Context, result *AnthropicMessagesResult, bodyPreview []byte, copyErr error) {
	s.anthropicProxy.LogMessagesResult(ctx, result, bodyPreview, copyErr)
}

func (s *UserConsoleService) LogChatCompletionsResult(ctx context.Context, result *OpenAIChatCompletionsResult, bodyPreview []byte, copyErr error) {
	s.openAIProxy.LogChatCompletionsResult(ctx, result, bodyPreview, copyErr)
}

func (s *UserConsoleService) LogGeminiGenerateContentResult(ctx context.Context, result *GeminiGenerateContentResult, bodyPreview []byte, copyErr error) {
	s.geminiProxy.LogGenerateContentResult(ctx, result, bodyPreview, copyErr)
}

func requireNormalUser(user entity.User) error {
	if user.Role != "user" {
		return validationError("admin users cannot use user console api keys")
	}
	if user.Status != "active" {
		return validationError("user is disabled")
	}
	return nil
}

func normalizePagination(page int, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
