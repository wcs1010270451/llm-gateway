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

	"google.golang.org/genai"
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
	client          *http.Client
	vertexClientsMu sync.RWMutex
	vertexClients   map[string]*genai.Client
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
	RequestType    string
}

func NewGeminiProxyService(routes *repository.RoutingRepository, logs *repository.RequestLogRepository, cipher *ProviderKeyCipher) *GeminiProxyService {
	return &GeminiProxyService{
		routes:        routes,
		logs:          logs,
		cipher:        cipher,
		client:        &http.Client{},
		vertexClients: make(map[string]*genai.Client),
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
		RequestType:    input.RequestType,
	}

	if route.Provider.AdapterType == "vertexai" {
		resp, err := s.openVertexAI(ctx, route, payload, input.Stream)
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
		}))
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

func (s *GeminiProxyService) openVertexAI(ctx context.Context, route repository.ResolvedRoute, payload map[string]any, stream bool) (*http.Response, error) {
	cfg := parseVertexAIConfig(route.Provider.ConfigJSON)
	if cfg.ProjectID == "" {
		return geminiJSONResponse(http.StatusBadRequest, map[string]any{
			"error": map[string]any{"type": "invalid_config", "message": "provider config_json missing project_id for vertexai provider"},
		}), nil
	}

	client, err := s.getVertexAIClient(ctx, cfg.ProjectID, cfg.Location)
	if err != nil {
		return geminiJSONResponse(http.StatusInternalServerError, map[string]any{
			"error": map[string]any{"type": "auth_error", "message": "failed to initialize Vertex AI client: " + err.Error()},
		}), nil
	}

	contentsJSON, _ := json.Marshal(payload["contents"])
	var contents []*genai.Content
	if err := json.Unmarshal(contentsJSON, &contents); err != nil {
		return geminiJSONResponse(http.StatusBadRequest, map[string]any{
			"error": map[string]any{"type": "invalid_request", "message": "invalid contents format"},
		}), nil
	}
	config := buildVertexAIGenConfig(payload)
	modelName := strings.TrimSpace(route.ProviderModel.UpstreamModel)
	if modelName == "" {
		return geminiJSONResponse(http.StatusBadRequest, map[string]any{
			"error": map[string]any{"type": "invalid_config", "message": "upstream model is required for vertexai provider model"},
		}), nil
	}

	if stream {
		reader, writer := io.Pipe()
		go func() {
			defer writer.Close()
			for chunk, err := range client.Models.GenerateContentStream(ctx, modelName, contents, config) {
				if err != nil {
					_ = writeSSEJSON(writer, map[string]any{"error": map[string]any{"type": "stream_error", "message": err.Error()}})
					return
				}
				chunkJSON, err := json.Marshal(chunk)
				if err != nil {
					continue
				}
				if _, err := writer.Write(append(append([]byte("data: "), chunkJSON...), '\n', '\n')); err != nil {
					return
				}
			}
		}()
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       reader,
		}, nil
	}

	resp, err := client.Models.GenerateContent(ctx, modelName, contents, config)
	if err != nil {
		return geminiJSONResponse(vertexAIErrorCode(err), map[string]any{
			"error": map[string]any{"type": "upstream_error", "message": err.Error()},
		}), nil
	}
	body, err := json.Marshal(resp)
	if err != nil {
		return geminiJSONResponse(http.StatusInternalServerError, map[string]any{
			"error": map[string]any{"type": "internal_error", "message": "failed to encode response"},
		}), nil
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json; charset=utf-8"}},
		Body:       io.NopCloser(bytes.NewReader(body)),
	}, nil
}

func (s *GeminiProxyService) LogGenerateContentResult(ctx context.Context, result *GeminiGenerateContentResult, bodyPreview []byte, copyErr error) {
	if result == nil || result.Response == nil {
		return
	}

	statusCode := result.Response.StatusCode
	promptTokens, completionTokens, totalTokens := extractGeminiUsage(bodyPreview)
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
	logMissingUsage(result.Route, result.PublicModel, result.Stream, result.RequestType, statusCode, promptTokens, completionTokens, totalTokens)

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
		RequestType:      result.RequestType,
	})
	if err := s.writeLog(ctx, logItem); err != nil {
		_ = s.writeLog(context.Background(), logItem)
	}
}

func (s *GeminiProxyService) writeLog(ctx context.Context, item *entity.RequestLog) error {
	return s.logs.Create(ctx, item)
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

func (s *GeminiProxyService) getVertexAIClient(ctx context.Context, projectID string, location string) (*genai.Client, error) {
	key := projectID + ":" + location
	s.vertexClientsMu.RLock()
	if client, ok := s.vertexClients[key]; ok {
		s.vertexClientsMu.RUnlock()
		return client, nil
	}
	s.vertexClientsMu.RUnlock()

	s.vertexClientsMu.Lock()
	defer s.vertexClientsMu.Unlock()
	if client, ok := s.vertexClients[key]; ok {
		return client, nil
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  projectID,
		Location: location,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, err
	}
	s.vertexClients[key] = client
	return client, nil
}

func buildVertexAIGenConfig(payload map[string]any) *genai.GenerateContentConfig {
	config := &genai.GenerateContentConfig{}
	if raw, ok := payload["generationConfig"]; ok {
		if body, err := json.Marshal(raw); err == nil {
			_ = json.Unmarshal(body, config)
		}
	}
	if raw, ok := payload["systemInstruction"]; ok {
		if body, err := json.Marshal(raw); err == nil {
			var content genai.Content
			if err := json.Unmarshal(body, &content); err == nil && len(content.Parts) > 0 {
				config.SystemInstruction = &content
			}
		}
	}
	if raw, ok := payload["tools"]; ok {
		if body, err := json.Marshal(raw); err == nil {
			var tools []*genai.Tool
			if err := json.Unmarshal(body, &tools); err == nil {
				config.Tools = tools
			}
		}
	}
	if raw, ok := payload["toolConfig"]; ok {
		if body, err := json.Marshal(raw); err == nil {
			var toolConfig genai.ToolConfig
			if err := json.Unmarshal(body, &toolConfig); err == nil {
				config.ToolConfig = &toolConfig
			}
		}
	}
	return config
}

func vertexAIErrorCode(err error) int {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "quota") || strings.Contains(message, "rate") || strings.Contains(message, "429"):
		return http.StatusTooManyRequests
	case strings.Contains(message, "permission") || strings.Contains(message, "403") || strings.Contains(message, "unauthorized"):
		return http.StatusForbidden
	case strings.Contains(message, "not found") || strings.Contains(message, "404"):
		return http.StatusNotFound
	case strings.Contains(message, "invalid") || strings.Contains(message, "400"):
		return http.StatusBadRequest
	default:
		return http.StatusBadGateway
	}
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

func extractGeminiUsage(body []byte) (int, int, int) {
	usage := extractGeminiUsageFromJSON(body)
	if usage.total > 0 {
		return usage.prompt, usage.completion, usage.total
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
	return usage.prompt, usage.completion, usage.total
}

type geminiUsage struct {
	prompt     int
	completion int
	total      int
}

func extractGeminiUsageFromJSON(body []byte) geminiUsage {
	var payload struct {
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return geminiUsage{}
	}
	total := payload.UsageMetadata.TotalTokenCount
	if total == 0 {
		total = payload.UsageMetadata.PromptTokenCount + payload.UsageMetadata.CandidatesTokenCount
	}
	return geminiUsage{
		prompt:     payload.UsageMetadata.PromptTokenCount,
		completion: payload.UsageMetadata.CandidatesTokenCount,
		total:      total,
	}
}
