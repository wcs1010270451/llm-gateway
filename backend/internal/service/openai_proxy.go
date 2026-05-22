package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/datatypes"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

type OpenAIProxyService struct {
	routes *repository.RoutingRepository
	logs   *repository.RequestLogRepository
	models *repository.ModelRepository
	cipher *ProviderKeyCipher
	client *http.Client
}

type OpenAIChatCompletionsInput struct {
	Body     []byte
	Headers  http.Header
	Method   string
	Path     string
	ClientIP string
	UserID   int64
	APIKeyID int64
}

type OpenAIChatCompletionsResult struct {
	Response       *http.Response
	Route          repository.ResolvedRoute
	PublicModel    string
	Stream         bool
	StartedAt      time.Time
	Method         string
	Path           string
	ClientIP       string
	UserID         int64
	APIKeyID       int64
	RequestPreview datatypes.JSON
}

func NewOpenAIProxyService(routes *repository.RoutingRepository, logs *repository.RequestLogRepository, models *repository.ModelRepository, cipher *ProviderKeyCipher) *OpenAIProxyService {
	return &OpenAIProxyService{
		routes: routes,
		logs:   logs,
		models: models,
		cipher: cipher,
		client: &http.Client{},
	}
}

func (s *OpenAIProxyService) ListModels(ctx context.Context) ([]entity.Model, error) {
	return s.models.ListEnabled(ctx)
}

func (s *OpenAIProxyService) OpenChatCompletions(ctx context.Context, input OpenAIChatCompletionsInput) (*OpenAIChatCompletionsResult, error) {
	startedAt := time.Now()

	var payload map[string]any
	if err := json.Unmarshal(input.Body, &payload); err != nil {
		return nil, validationError("invalid JSON request body")
	}

	publicModel, ok := payload["model"].(string)
	if !ok || strings.TrimSpace(publicModel) == "" {
		return nil, validationError("request model is required")
	}
	publicModel = strings.TrimSpace(publicModel)
	stream, _ := payload["stream"].(bool)

	// 对外模型名由网关维护，当前激活的供应商路由决定实际使用哪个上游模型和协议。
	route, err := s.routes.ResolveByPublicModel(ctx, publicModel)
	if err != nil {
		return nil, err
	}
	if route.Model.Status != "enabled" {
		return nil, validationError("model is disabled")
	}
	if route.Model.ActiveProviderModelID == nil || route.ProviderModel.ID == 0 {
		return nil, validationError("model has no active provider route")
	}
	if route.Provider.Status != "active" {
		return nil, validationError("active provider is disabled")
	}
	if route.Provider.AdapterType != "openai_compatible" {
		return nil, validationError("active provider adapter is %s, not openai_compatible", route.Provider.AdapterType)
	}
	if route.ProviderModel.Status != "enabled" {
		return nil, validationError("active provider model is disabled")
	}
	if strings.TrimSpace(route.Provider.BaseURL) == "" {
		return nil, validationError("active provider base_url is required")
	}

	// 上游 Key 在库里加密保存，只在转发路径中解密，避免管理接口暴露供应商凭证。
	upstreamAPIKey := ""
	if route.Provider.AuthType == "api_key" {
		if strings.TrimSpace(route.Provider.APIKeyEncrypted) == "" {
			return nil, validationError("active provider api key is required")
		}
		upstreamAPIKey, err = s.cipher.Decrypt(route.Provider.APIKeyEncrypted)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(upstreamAPIKey) == "" {
			return nil, validationError("active provider api key is required")
		}
	}

	// 只重写 model 字段，其余 OpenAI-compatible 参数透传，保留供应商兼容扩展能力。
	payload["model"] = route.ProviderModel.UpstreamModel
	upstreamBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	upstreamURL := buildUpstreamURL(route.Provider.BaseURL, "/v1/chat/completions")
	upstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, upstreamURL, bytes.NewReader(upstreamBody))
	if err != nil {
		return nil, err
	}
	upstreamReq.Header.Set("Content-Type", "application/json")
	if route.Provider.AuthType == "api_key" {
		upstreamReq.Header.Set("Authorization", "Bearer "+upstreamAPIKey)
	}

	resp, err := s.client.Do(upstreamReq)
	if err != nil {
		_ = s.writeLog(context.Background(), buildRequestLog(RequestLogInput{
			Route:          route,
			PublicModel:    publicModel,
			Stream:         stream,
			Method:         input.Method,
			Path:           input.Path,
			ClientIP:       input.ClientIP,
			UserID:         input.UserID,
			APIKeyID:       input.APIKeyID,
			StartedAt:      startedAt,
			StatusCode:     http.StatusBadGateway,
			ErrorType:      "upstream_request_error",
			ErrorMessage:   err.Error(),
			RequestPreview: buildRequestPreview(publicModel, route.ProviderModel.UpstreamModel, stream),
			RequestType:    "chat_completions",
		}))
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}

	return &OpenAIChatCompletionsResult{
		Response:       resp,
		Route:          route,
		PublicModel:    publicModel,
		Stream:         stream,
		StartedAt:      startedAt,
		Method:         input.Method,
		Path:           input.Path,
		ClientIP:       input.ClientIP,
		UserID:         input.UserID,
		APIKeyID:       input.APIKeyID,
		RequestPreview: buildRequestPreview(publicModel, route.ProviderModel.UpstreamModel, stream),
	}, nil
}

func (s *OpenAIProxyService) LogChatCompletionsResult(ctx context.Context, result *OpenAIChatCompletionsResult, bodyPreview []byte, copyErr error) {
	if result == nil || result.Response == nil {
		return
	}

	// 非流式响应通常有完整 usage；流式日志只保存片段，未解析 stream usage 前 Token 可能为 0。
	statusCode := result.Response.StatusCode
	promptTokens, completionTokens, totalTokens := extractOpenAIUsage(bodyPreview)
	errorType := ""
	errorMessage := ""
	if statusCode < 200 || statusCode >= 300 {
		errorType = "upstream_error"
		errorMessage = truncateString(string(bodyPreview), 1000)
	}
	if copyErr != nil {
		errorType = "response_copy_error"
		errorMessage = copyErr.Error()
	}

	logItem := buildRequestLog(RequestLogInput{
		Route:            result.Route,
		PublicModel:      result.PublicModel,
		Stream:           result.Stream,
		Method:           result.Method,
		Path:             result.Path,
		ClientIP:         result.ClientIP,
		UserID:           result.UserID,
		APIKeyID:         result.APIKeyID,
		StartedAt:        result.StartedAt,
		StatusCode:       statusCode,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		ErrorType:        errorType,
		ErrorMessage:     errorMessage,
		RequestPreview:   result.RequestPreview,
		ResponsePreview:  buildBodyPreview(bodyPreview),
		RequestType:      "chat_completions",
	})
	if err := s.writeLog(ctx, logItem); err != nil {
		_ = s.writeLog(context.Background(), logItem)
	}
}

func (s *OpenAIProxyService) writeLog(ctx context.Context, item *entity.RequestLog) error {
	return s.logs.Create(ctx, item)
}

func extractOpenAIUsage(body []byte) (int, int, int) {
	var payload struct {
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, 0, 0
	}
	totalTokens := payload.Usage.TotalTokens
	if totalTokens == 0 {
		totalTokens = payload.Usage.PromptTokens + payload.Usage.CompletionTokens
	}
	return payload.Usage.PromptTokens, payload.Usage.CompletionTokens, totalTokens
}
