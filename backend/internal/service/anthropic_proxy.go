package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

type AnthropicProxyService struct {
	routes *repository.RoutingRepository
	logs   *repository.RequestLogRepository
	cipher *ProviderKeyCipher
	bodies *RequestLogBodyStore
	client *http.Client
}

type AnthropicMessagesInput struct {
	Body     []byte
	Headers  http.Header
	Method   string
	Path     string
	ClientIP string
	UserID   int64
	APIKeyID int64
}

type AnthropicMessagesResult struct {
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
	RequestBody    []byte
}

func NewAnthropicProxyService(routes *repository.RoutingRepository, logs *repository.RequestLogRepository, cipher *ProviderKeyCipher, bodies *RequestLogBodyStore) *AnthropicProxyService {
	return &AnthropicProxyService{
		routes: routes,
		logs:   logs,
		cipher: cipher,
		bodies: bodies,
		client: &http.Client{},
	}
}

func (s *AnthropicProxyService) OpenMessages(ctx context.Context, input AnthropicMessagesInput) (*AnthropicMessagesResult, error) {
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
	if route.Provider.AdapterType != "anthropic" {
		return nil, validationError("active provider adapter is %s, not anthropic", route.Provider.AdapterType)
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

	// 只重写 model 字段，其余 Anthropic Messages 参数透传，保留供应商兼容扩展能力。
	payload["model"] = route.ProviderModel.UpstreamModel

	var (
		upstreamBody []byte
		upstreamReq  *http.Request
	)
	if route.Provider.AuthType == "claude_oauth" {
		// owner 自有 Claude Max 订阅：注入 Claude Code 伪装（system 迁移 + cch + OAuth Bearer），直连 Anthropic。
		cfg := parseClaudeOAuthConfig(route.Provider.ConfigJSON)
		token, tokenErr := loadClaudeAccessToken(cfg.CredentialsFile)
		if tokenErr != nil {
			return nil, fmt.Errorf("claude_oauth credentials: %w", tokenErr)
		}
		upstreamBody, err = injectClaudeSystemAndCCH(payload, cfg.fullVersion())
		if err != nil {
			return nil, err
		}
		base := strings.TrimSpace(route.Provider.BaseURL)
		if base == "" {
			base = defaultAnthropicBaseURL
		}
		upstreamURL := buildUpstreamURL(base, "/v1/messages") + "?beta=true"
		upstreamReq, err = http.NewRequestWithContext(ctx, http.MethodPost, upstreamURL, bytes.NewReader(upstreamBody))
		if err != nil {
			return nil, err
		}
		applyClaudeOAuthHeaders(upstreamReq.Header, token, cfg, stream)
	} else {
		upstreamBody, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		upstreamURL := buildUpstreamURL(route.Provider.BaseURL, "/v1/messages")
		upstreamReq, err = http.NewRequestWithContext(ctx, http.MethodPost, upstreamURL, bytes.NewReader(upstreamBody))
		if err != nil {
			return nil, err
		}
		upstreamReq.Header.Set("Content-Type", "application/json")
		if route.Provider.AuthType == "api_key" {
			upstreamReq.Header.Set("x-api-key", upstreamAPIKey)
		}
		if version := strings.TrimSpace(input.Headers.Get("anthropic-version")); version != "" {
			upstreamReq.Header.Set("anthropic-version", version)
		} else {
			upstreamReq.Header.Set("anthropic-version", "2023-06-01")
		}
		if beta := strings.TrimSpace(input.Headers.Get("anthropic-beta")); beta != "" {
			upstreamReq.Header.Set("anthropic-beta", beta)
		}
	}

	resp, err := s.client.Do(upstreamReq)
	if err != nil {
		s.writeLog(context.Background(), buildRequestLog(RequestLogInput{
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
		}), input.Body, nil)
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}

	return &AnthropicMessagesResult{
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
		RequestBody:    input.Body,
	}, nil
}

func (s *AnthropicProxyService) LogMessagesResult(ctx context.Context, result *AnthropicMessagesResult, bodyPreview []byte, copyErr error) {
	if result == nil || result.Response == nil {
		return
	}

	// 非流式响应通常有完整 usage；流式日志只保存片段，未解析 stream usage 前 Token 可能为 0。
	statusCode := result.Response.StatusCode
	usage := extractAnthropicUsage(bodyPreview)
	totalTokens := usage.total()
	errorType := ""
	errorMessage := ""
	if statusCode < 200 || statusCode >= 300 {
		errorType = "upstream_error"
		errorMessage = truncateString(string(bodyPreview), 1000)
		// 打到 stderr，让上游错误（如 long context 需 credits）能进 Cloud Logging。
		log.Printf("anthropic upstream error: model=%s status=%d body=%s",
			result.PublicModel, statusCode, truncateString(string(bodyPreview), 500))
	}
	if copyErr != nil {
		errorType = "response_copy_error"
		errorMessage = copyErr.Error()
		log.Printf("anthropic response copy error: model=%s: %v", result.PublicModel, copyErr)
	}
	logMissingUsage(result.Route, result.PublicModel, result.Stream, "messages", statusCode, usage.input, usage.output, totalTokens)

	logItem := buildRequestLog(RequestLogInput{
		Route:                    result.Route,
		PublicModel:              result.PublicModel,
		Stream:                   result.Stream,
		Method:                   result.Method,
		Path:                     result.Path,
		ClientIP:                 result.ClientIP,
		UserID:                   result.UserID,
		APIKeyID:                 result.APIKeyID,
		StartedAt:                result.StartedAt,
		StatusCode:               statusCode,
		PromptTokens:             usage.input,
		CompletionTokens:         usage.output,
		CacheCreationInputTokens: usage.cacheCreation,
		CacheReadInputTokens:     usage.cacheRead,
		TotalTokens:              totalTokens,
		ErrorType:                errorType,
		ErrorMessage:             errorMessage,
		RequestPreview:           result.RequestPreview,
		ResponsePreview:          buildBodyPlaceholder(bodyPreview),
	})
	if err := s.writeLog(ctx, logItem, result.RequestBody, bodyPreview); err != nil {
		_ = s.writeLog(context.Background(), logItem, result.RequestBody, bodyPreview)
	}
}

type RequestLogInput struct {
	Route                    repository.ResolvedRoute
	PublicModel              string
	Stream                   bool
	Method                   string
	Path                     string
	ClientIP                 string
	UserID                   int64
	APIKeyID                 int64
	StartedAt                time.Time
	StatusCode               int
	PromptTokens             int
	CompletionTokens         int
	CacheCreationInputTokens int
	CacheReadInputTokens     int
	ReasoningTokens          int
	ToolTokens               int
	TotalTokens              int
	ErrorType                string
	ErrorMessage             string
	RequestPreview           datatypes.JSON
	ResponsePreview          datatypes.JSON
	RequestType              string
}

func buildRequestLog(input RequestLogInput) *entity.RequestLog {
	modelID := input.Route.Model.ID
	providerID := input.Route.Provider.ID
	providerModelID := input.Route.ProviderModel.ID
	userID := input.UserID
	apiKeyID := input.APIKeyID
	estimatedCost := estimateCost(input.PromptTokens, input.CompletionTokens, input.Route.ProviderModel)
	// Anthropic 和 OpenAI-compatible 共用日志构造器；非 /v1/messages 路径由调用方覆盖类型。
	requestType := input.RequestType
	if requestType == "" {
		requestType = "messages"
	}

	return &entity.RequestLog{
		RequestID:                uuid.NewString(),
		UserID:                   &userID,
		APIKeyID:                 &apiKeyID,
		ModelID:                  &modelID,
		PublicModelName:          input.PublicModel,
		ProviderID:               &providerID,
		ProviderModelID:          &providerModelID,
		AdapterType:              input.Route.Provider.AdapterType,
		UpstreamModel:            input.Route.ProviderModel.UpstreamModel,
		RequestType:              requestType,
		Stream:                   input.Stream,
		ClientIP:                 input.ClientIP,
		RequestMethod:            input.Method,
		RequestPath:              input.Path,
		HTTPStatus:               input.StatusCode,
		Success:                  input.StatusCode >= 200 && input.StatusCode < 300 && input.ErrorType == "",
		LatencyMS:                int(time.Since(input.StartedAt).Milliseconds()),
		PromptTokens:             input.PromptTokens,
		CompletionTokens:         input.CompletionTokens,
		CacheCreationInputTokens: input.CacheCreationInputTokens,
		CacheReadInputTokens:     input.CacheReadInputTokens,
		ReasoningTokens:          input.ReasoningTokens,
		ToolTokens:               input.ToolTokens,
		TotalTokens:              input.TotalTokens,
		EstimatedCost:            estimatedCost,
		ErrorType:                input.ErrorType,
		ErrorMessage:             truncateString(input.ErrorMessage, 1000),
		RequestPreview:           withDefaultJSON(input.RequestPreview),
		ResponsePreview:          withDefaultJSON(input.ResponsePreview),
		Metadata:                 datatypes.JSON([]byte("{}")),
	}
}

func (s *AnthropicProxyService) writeLog(ctx context.Context, item *entity.RequestLog, requestBody []byte, responseBody []byte) error {
	if err := s.logs.Create(ctx, item); err != nil {
		return err
	}
	s.bodies.Save(context.Background(), s.logs, item, requestBody, responseBody)
	return nil
}

func buildRequestPreview(publicModel string, upstreamModel string, stream bool) datatypes.JSON {
	return mustJSON(map[string]any{
		"model":          publicModel,
		"upstream_model": upstreamModel,
		"stream":         stream,
	})
}

func buildBodyPlaceholder(body []byte) datatypes.JSON {
	if len(body) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	return mustJSON(map[string]any{
		"stored": true,
		"bytes":  len(body),
	})
}

func withDefaultJSON(value datatypes.JSON) datatypes.JSON {
	if len(value) == 0 {
		return datatypes.JSON([]byte("{}"))
	}
	return value
}

func mustJSON(value any) datatypes.JSON {
	data, err := json.Marshal(value)
	if err != nil {
		return datatypes.JSON([]byte("{}"))
	}
	return datatypes.JSON(data)
}

func extractAnthropicUsage(body []byte) anthropicUsage {
	usage := extractAnthropicUsageFromJSON(body)
	if usage.input > 0 || usage.output > 0 {
		return usage
	}

	for _, event := range bytes.Split(body, []byte("\n\n")) {
		for _, line := range bytes.Split(event, []byte("\n")) {
			line = bytes.TrimSpace(line)
			if !bytes.HasPrefix(line, []byte("data:")) {
				continue
			}
			raw := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
			if len(raw) == 0 || bytes.Equal(raw, []byte("[DONE]")) {
				continue
			}
			next := extractAnthropicUsageFromJSON(raw)
			if next.input > 0 {
				usage.input = next.input
			}
			if next.output > 0 {
				usage.output = next.output
			}
			if next.cacheCreation > 0 {
				usage.cacheCreation = next.cacheCreation
			}
			if next.cacheRead > 0 {
				usage.cacheRead = next.cacheRead
			}
		}
	}
	return usage
}

type anthropicUsage struct {
	input         int
	output        int
	cacheCreation int
	cacheRead     int
}

func (u anthropicUsage) total() int {
	return u.input + u.output + u.cacheCreation + u.cacheRead
}

func extractAnthropicUsageFromJSON(body []byte) anthropicUsage {
	var payload struct {
		Usage *struct {
			InputTokens              int `json:"input_tokens"`
			OutputTokens             int `json:"output_tokens"`
			CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
			CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		} `json:"usage"`
		Message *struct {
			Usage *struct {
				InputTokens              int `json:"input_tokens"`
				OutputTokens             int `json:"output_tokens"`
				CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
				CacheReadInputTokens     int `json:"cache_read_input_tokens"`
			} `json:"usage"`
		} `json:"message"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || payload.Usage == nil {
		if payload.Message == nil || payload.Message.Usage == nil {
			return anthropicUsage{}
		}
		return anthropicUsage{
			input:         payload.Message.Usage.InputTokens,
			output:        payload.Message.Usage.OutputTokens,
			cacheCreation: payload.Message.Usage.CacheCreationInputTokens,
			cacheRead:     payload.Message.Usage.CacheReadInputTokens,
		}
	}
	return anthropicUsage{
		input:         payload.Usage.InputTokens,
		output:        payload.Usage.OutputTokens,
		cacheCreation: payload.Usage.CacheCreationInputTokens,
		cacheRead:     payload.Usage.CacheReadInputTokens,
	}
}

func estimateCost(promptTokens int, completionTokens int, providerModel entity.ProviderModel) float64 {
	inputCost := float64(promptTokens) / 1_000_000 * providerModel.InputCostPer1M
	outputCost := float64(completionTokens) / 1_000_000 * providerModel.OutputCostPer1M
	return inputCost + outputCost
}

func truncateString(value string, maxLength int) string {
	if len(value) <= maxLength {
		return value
	}
	return value[:maxLength]
}
