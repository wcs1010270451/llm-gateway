package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/datatypes"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

const defaultGeminiBaseURL = "https://generativelanguage.googleapis.com"
const defaultGeminiAPIVersion = "v1beta"
const defaultVertexAILocation = "us-central1"

type GeminiProxyService struct {
	routes          *repository.RoutingRepository
	logs            *repository.RequestLogRepository
	cipher          *ProviderKeyCipher
	bodies          *RequestLogBodyStore
	client          *http.Client
	vertexTokenSourceMu sync.RWMutex
	vertexTokenSource   oauth2.TokenSource
}

type GeminiGenerateContentInput struct {
	Body        []byte
	Headers     http.Header
	Method      string
	Path        string
	ClientIP    string
	UserID      int64
	APIKeyID    int64
	Stream      bool
	RequestType string
}

type GeminiGenerateContentResult struct {
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
	RequestType    string
}

func NewGeminiProxyService(routes *repository.RoutingRepository, logs *repository.RequestLogRepository, cipher *ProviderKeyCipher, bodies *RequestLogBodyStore) *GeminiProxyService {
	return &GeminiProxyService{
		routes:        routes,
		logs:          logs,
		cipher:        cipher,
		bodies:        bodies,
		client:        &http.Client{},
	}
}

func (s *GeminiProxyService) OpenGenerateContent(ctx context.Context, input GeminiGenerateContentInput) (*GeminiGenerateContentResult, error) {
	input.Stream = false
	if input.RequestType == "" {
		input.RequestType = "gemini_generate_content"
	}
	return s.open(ctx, input, "generateContent")
}

func (s *GeminiProxyService) OpenStreamGenerateContent(ctx context.Context, input GeminiGenerateContentInput) (*GeminiGenerateContentResult, error) {
	input.Stream = true
	if input.RequestType == "" {
		input.RequestType = "gemini_stream_generate_content"
	}
	return s.open(ctx, input, "streamGenerateContent")
}

func (s *GeminiProxyService) open(ctx context.Context, input GeminiGenerateContentInput, upstreamMethod string) (*GeminiGenerateContentResult, error) {
	startedAt := time.Now()
	payload, publicModel, err := parseGeminiPayload(input.Body)
	if err != nil {
		return nil, err
	}

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
	if route.Provider.AdapterType != "gemini" && route.Provider.AdapterType != "vertexai" {
		return nil, validationError("active provider adapter is %s, not gemini or vertexai", route.Provider.AdapterType)
	}
	if route.ProviderModel.Status != "enabled" {
		return nil, validationError("active provider model is disabled")
	}

	delete(payload, "model")
	result := &GeminiGenerateContentResult{
		Route:          route,
		PublicModel:    publicModel,
		Stream:         input.Stream,
		StartedAt:      startedAt,
		Method:         input.Method,
		Path:           input.Path,
		ClientIP:       input.ClientIP,
		UserID:         input.UserID,
		APIKeyID:       input.APIKeyID,
		RequestPreview: buildRequestPreview(publicModel, route.ProviderModel.UpstreamModel, input.Stream),
		RequestBody:    input.Body,
		RequestType:    input.RequestType,
	}

	if route.Provider.AdapterType == "vertexai" {
		resp, err := s.openVertexAI(ctx, route, payload, upstreamMethod, input.Stream)
		if err != nil {
			return nil, err
		}
		result.Response = resp
		return result, nil
	}

	resp, err := s.openGeminiREST(ctx, route, payload, upstreamMethod, input.Stream, input.Headers)
	if err != nil {
		_ = s.writeLog(context.Background(), buildRequestLog(RequestLogInput{
			Route:          route,
			PublicModel:    publicModel,
			Stream:         input.Stream,
			Method:         input.Method,
			Path:           input.Path,
			ClientIP:       input.ClientIP,
			UserID:         input.UserID,
			APIKeyID:       input.APIKeyID,
			StartedAt:      startedAt,
			StatusCode:     http.StatusBadGateway,
			ErrorType:      "upstream_request_error",
			ErrorMessage:   err.Error(),
			RequestPreview: result.RequestPreview,
			RequestType:    input.RequestType,
		}), input.Body, nil)
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	result.Response = resp
	return result, nil
}

func (s *GeminiProxyService) openGeminiREST(ctx context.Context, route repository.ResolvedRoute, payload map[string]any, method string, stream bool, headers http.Header) (*http.Response, error) {
	upstreamAPIKey := ""
	if route.Provider.AuthType == "api_key" {
		if strings.TrimSpace(route.Provider.APIKeyEncrypted) == "" {
			return nil, validationError("active provider api key is required")
		}
		var err error
		upstreamAPIKey, err = s.cipher.Decrypt(route.Provider.APIKeyEncrypted)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(upstreamAPIKey) == "" {
			return nil, validationError("active provider api key is required")
		}
	}

	upstreamBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	baseURL := strings.TrimSpace(route.Provider.BaseURL)
	if baseURL == "" {
		baseURL = defaultGeminiBaseURL
	}
	apiVersion := geminiAPIVersion(route.Provider.ConfigJSON)
	upstreamURL := buildGeminiUpstreamURL(baseURL, apiVersion, route.ProviderModel.UpstreamModel, method, stream)
	upstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, upstreamURL, bytes.NewReader(upstreamBody))
	if err != nil {
		return nil, err
	}
	upstreamReq.Header.Set("Content-Type", "application/json")
	upstreamReq.Header.Set("Accept", "application/json")
	if stream {
		upstreamReq.Header.Set("Accept", "text/event-stream")
	}
	if route.Provider.AuthType == "api_key" {
		upstreamReq.Header.Set("x-goog-api-key", upstreamAPIKey)
	}
	if value := strings.TrimSpace(headers.Get("x-goog-api-client")); value != "" {
		upstreamReq.Header.Set("x-goog-api-client", value)
	}

	return s.client.Do(upstreamReq)
}

func (s *GeminiProxyService) openVertexAI(ctx context.Context, route repository.ResolvedRoute, payload map[string]any, upstreamMethod string, stream bool) (*http.Response, error) {
	cfg := parseVertexAIConfig(route.Provider.ConfigJSON)
	if cfg.ProjectID == "" {
		return geminiJSONResponse(http.StatusBadRequest, map[string]any{
			"error": map[string]any{"type": "invalid_config", "message": "provider config_json missing project_id for vertexai provider"},
		}), nil
	}

	modelName := strings.TrimSpace(route.ProviderModel.UpstreamModel)
	if modelName == "" {
		return geminiJSONResponse(http.StatusBadRequest, map[string]any{
			"error": map[string]any{"type": "invalid_config", "message": "upstream model is required for vertexai provider model"},
		}), nil
	}

	accessToken, err := s.getVertexAccessToken(ctx)
	if err != nil {
		return geminiJSONResponse(http.StatusInternalServerError, map[string]any{
			"error": map[string]any{"type": "auth_error", "message": "failed to obtain Vertex AI access token: " + err.Error()},
		}), nil
	}

	u := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:%s", cfg.Location, cfg.ProjectID, cfg.Location, modelName, upstreamMethod)
	if stream {
		u += "?alt=sse"
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return geminiJSONResponse(http.StatusInternalServerError, map[string]any{
			"error": map[string]any{"type": "internal_error", "message": "failed to encode request payload"},
		}), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(bodyBytes))
	if err != nil {
		return geminiJSONResponse(http.StatusInternalServerError, map[string]any{
			"error": map[string]any{"type": "internal_error", "message": "failed to build upstream request"},
		}), nil
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return geminiJSONResponse(http.StatusBadGateway, map[string]any{
			"error": map[string]any{"type": "upstream_error", "message": "vertex ai connection failed: " + err.Error()},
		}), nil
	}
	return resp, nil
}

func (s *GeminiProxyService) LogGenerateContentResult(ctx context.Context, result *GeminiGenerateContentResult, bodyPreview []byte, copyErr error) {
	if result == nil || result.Response == nil {
		return
	}

	statusCode := result.Response.StatusCode
	usage := extractGeminiUsage(bodyPreview)
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
	logMissingUsage(result.Route, result.PublicModel, result.Stream, result.RequestType, statusCode, usage.prompt, usage.completion, usage.total)

	logItem := buildRequestLog(RequestLogInput{
		Route:                result.Route,
		PublicModel:          result.PublicModel,
		Stream:               result.Stream,
		Method:               result.Method,
		Path:                 result.Path,
		ClientIP:             result.ClientIP,
		UserID:               result.UserID,
		APIKeyID:             result.APIKeyID,
		StartedAt:            result.StartedAt,
		StatusCode:           statusCode,
		PromptTokens:         usage.prompt,
		CompletionTokens:     usage.completion,
		CacheReadInputTokens: usage.cached,
		ReasoningTokens:      usage.thoughts,
		ToolTokens:           usage.tool,
		TotalTokens:          usage.total,
		ErrorType:            errorType,
		ErrorMessage:         errorMessage,
		RequestPreview:       result.RequestPreview,
		ResponsePreview:      buildBodyPlaceholder(bodyPreview),
		RequestType:          result.RequestType,
	})
	if err := s.writeLog(ctx, logItem, result.RequestBody, bodyPreview); err != nil {
		_ = s.writeLog(context.Background(), logItem, result.RequestBody, bodyPreview)
	}
}

func (s *GeminiProxyService) writeLog(ctx context.Context, item *entity.RequestLog, requestBody []byte, responseBody []byte) error {
	if err := s.logs.Create(ctx, item); err != nil {
		return err
	}
	s.bodies.Save(context.Background(), s.logs, item, requestBody, responseBody)
	return nil
}

func parseGeminiPayload(body []byte) (map[string]any, string, error) {
	payload := make(map[string]any)
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, "", validationError("request body must be valid JSON")
	}
	modelName, ok := payload["model"].(string)
	if !ok || strings.TrimSpace(modelName) == "" {
		return nil, "", validationError("model is required")
	}
	if _, ok := payload["contents"]; !ok {
		return nil, "", validationError("contents is required")
	}
	return payload, strings.TrimSpace(modelName), nil
}

func geminiAPIVersion(configJSON datatypes.JSON) string {
	version := defaultGeminiAPIVersion
	var payload map[string]any
	if err := json.Unmarshal(configJSON, &payload); err != nil {
		return version
	}
	if configuredVersion, ok := payload["gemini_api_version"].(string); ok && strings.TrimSpace(configuredVersion) != "" {
		return strings.Trim(strings.TrimSpace(configuredVersion), "/")
	}
	return version
}

func buildGeminiUpstreamURL(baseURL string, apiVersion string, model string, method string, stream bool) string {
	upstreamURL := buildUpstreamURL(baseURL, "/"+strings.Trim(apiVersion, "/")+"/models/"+strings.TrimSpace(model)+":"+strings.TrimSpace(method))
	if !stream {
		return upstreamURL
	}
	parsed, err := url.Parse(upstreamURL)
	if err != nil {
		if strings.Contains(upstreamURL, "?") {
			return upstreamURL + "&alt=sse"
		}
		return upstreamURL + "?alt=sse"
	}
	query := parsed.Query()
	query.Set("alt", "sse")
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

type vertexAIConfig struct {
	ProjectID string
	Location  string
}

func parseVertexAIConfig(configJSON datatypes.JSON) vertexAIConfig {
	cfg := vertexAIConfig{Location: defaultVertexAILocation}
	var payload map[string]any
	if err := json.Unmarshal(configJSON, &payload); err != nil {
		return cfg
	}
	if value, ok := payload["project_id"].(string); ok && strings.TrimSpace(value) != "" {
		cfg.ProjectID = strings.TrimSpace(value)
	}
	if value, ok := payload["location"].(string); ok && strings.TrimSpace(value) != "" {
		cfg.Location = strings.TrimSpace(value)
	}
	return cfg
}

func (s *GeminiProxyService) getVertexAccessToken(ctx context.Context) (string, error) {
	s.vertexTokenSourceMu.RLock()
	ts := s.vertexTokenSource
	s.vertexTokenSourceMu.RUnlock()

	if ts == nil {
		s.vertexTokenSourceMu.Lock()
		if s.vertexTokenSource == nil {
			tsLocal, err := google.DefaultTokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
			if err != nil {
				s.vertexTokenSourceMu.Unlock()
				return "", err
			}
			s.vertexTokenSource = tsLocal
		}
		ts = s.vertexTokenSource
		s.vertexTokenSourceMu.Unlock()
	}

	tok, err := ts.Token()
	if err != nil {
		return "", err
	}
	return tok.AccessToken, nil
}



func geminiJSONResponse(statusCode int, payload map[string]any) *http.Response {
	body, _ := json.Marshal(payload)
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
		Body:       io.NopCloser(bytes.NewReader(body)),
	}
}

func writeSSEJSON(writer io.Writer, payload map[string]any) error {
	body, _ := json.Marshal(payload)
	_, err := writer.Write(append(append([]byte("data: "), body...), '\n', '\n'))
	return err
}

func extractGeminiUsage(body []byte) geminiUsage {
	usage := extractGeminiUsageFromJSON(body)
	if usage.total > 0 {
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
			next := extractGeminiUsageFromJSON(raw)
			if next.total > 0 {
				usage = next
			}
		}
	}
	return usage
}

type geminiUsage struct {
	prompt     int
	completion int
	cached     int
	thoughts   int
	tool       int
	total      int
}

func extractGeminiUsageFromJSON(body []byte) geminiUsage {
	var payload struct {
		UsageMetadata struct {
			PromptTokenCount        int `json:"promptTokenCount"`
			CandidatesTokenCount    int `json:"candidatesTokenCount"`
			CachedContentTokenCount int `json:"cachedContentTokenCount"`
			ThoughtsTokenCount      int `json:"thoughtsTokenCount"`
			ToolUsePromptTokenCount int `json:"toolUsePromptTokenCount"`
			TotalTokenCount         int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return geminiUsage{}
	}
	total := payload.UsageMetadata.TotalTokenCount
	if total == 0 {
		total = payload.UsageMetadata.PromptTokenCount + payload.UsageMetadata.CandidatesTokenCount + payload.UsageMetadata.ThoughtsTokenCount + payload.UsageMetadata.ToolUsePromptTokenCount
	}
	return geminiUsage{
		prompt:     payload.UsageMetadata.PromptTokenCount,
		completion: payload.UsageMetadata.CandidatesTokenCount,
		cached:     payload.UsageMetadata.CachedContentTokenCount,
		thoughts:   payload.UsageMetadata.ThoughtsTokenCount,
		tool:       payload.UsageMetadata.ToolUsePromptTokenCount,
		total:      total,
	}
}
