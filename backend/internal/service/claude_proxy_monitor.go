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
	"time"

	"llm-gateway/backend/internal/entity"
)

type ClaudeProxyMonitorService struct {
	providers *ProviderService
	client    *http.Client
}

type ClaudeProxyStatus struct {
	ProviderID  int64    `json:"provider_id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Status      string   `json:"status"`
	BaseURL     string   `json:"base_url"`
	Reachable   bool     `json:"reachable"`
	ProxyStatus string   `json:"proxy_status"`
	TokenHours  *float64 `json:"token_hours,omitempty"`
	CCVersion   string   `json:"cc_version,omitempty"`
	HTTPStatus  int      `json:"http_status,omitempty"`
	Error       string   `json:"error,omitempty"`
	CheckedAt   string   `json:"checked_at"`
}

type ClaudeProxyProbeResult struct {
	ClaudeProxyStatus
	ProbeOK bool `json:"probe_ok"`
}

type claudeProxyHealthResponse struct {
	Status     string   `json:"status"`
	TokenHours *float64 `json:"token_hours"`
	CCVersion  string   `json:"cc_version"`
	Error      string   `json:"error"`
}

func NewClaudeProxyMonitorService(providers *ProviderService) *ClaudeProxyMonitorService {
	return &ClaudeProxyMonitorService{
		providers: providers,
		client:    &http.Client{Timeout: 8 * time.Second},
	}
}

func (s *ClaudeProxyMonitorService) ListStatus(ctx context.Context) ([]ClaudeProxyStatus, error) {
	providers, err := s.claudeCodeProviders(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]ClaudeProxyStatus, 0, len(providers))
	for _, provider := range providers {
		items = append(items, s.checkHealth(ctx, provider))
	}
	return items, nil
}

func (s *ClaudeProxyMonitorService) Probe(ctx context.Context, providerID int64) (ClaudeProxyProbeResult, error) {
	provider, err := s.providers.Get(ctx, providerID)
	if err != nil {
		return ClaudeProxyProbeResult{}, err
	}
	if !isClaudeProxyProvider(provider) {
		return ClaudeProxyProbeResult{}, validationError("provider is not a claude proxy")
	}

	status := s.checkHealth(ctx, provider)
	result := ClaudeProxyProbeResult{ClaudeProxyStatus: status}
	if provider.BaseURL == "" {
		result.Error = "base_url is required"
		return result, nil
	}

	body := []byte(`{"model":"claude-sonnet-4-6","max_tokens":8,"messages":[{"role":"user","content":"ping"}]}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(provider.BaseURL, "/v1/messages"), bytes.NewReader(body))
	if err != nil {
		result.Error = err.Error()
		return result, nil
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", "monitor")

	resp, err := s.client.Do(req)
	if err != nil {
		result.Error = err.Error()
		return result, nil
	}
	defer resp.Body.Close()

	result.HTTPStatus = resp.StatusCode
	result.ProbeOK = resp.StatusCode >= 200 && resp.StatusCode < 300
	result.Reachable = true
	result.CheckedAt = time.Now().UTC().Format(time.RFC3339)
	if !result.ProbeOK {
		result.Error = readResponsePreview(resp.Body)
	}
	return result, nil
}

func (s *ClaudeProxyMonitorService) claudeCodeProviders(ctx context.Context) ([]entity.Provider, error) {
	items, err := s.providers.List(ctx)
	if err != nil {
		return nil, err
	}
	providers := make([]entity.Provider, 0)
	for _, item := range items {
		if isClaudeProxyProvider(item) {
			providers = append(providers, item)
		}
	}
	return providers, nil
}

func isClaudeProxyProvider(provider entity.Provider) bool {
	return provider.Slug == "claude_max_proxy"
}

func (s *ClaudeProxyMonitorService) checkHealth(ctx context.Context, provider entity.Provider) ClaudeProxyStatus {
	status := ClaudeProxyStatus{
		ProviderID: provider.ID,
		Name:       provider.Name,
		Slug:       provider.Slug,
		Status:     provider.Status,
		BaseURL:    provider.BaseURL,
		CheckedAt:  time.Now().UTC().Format(time.RFC3339),
	}
	if strings.TrimSpace(provider.BaseURL) == "" {
		status.ProxyStatus = "missing_base_url"
		status.Error = "base_url is required"
		return status
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, joinURL(provider.BaseURL, "/health"), nil)
	if err != nil {
		status.ProxyStatus = "error"
		status.Error = err.Error()
		return status
	}
	resp, err := s.client.Do(req)
	if err != nil {
		status.ProxyStatus = "unreachable"
		status.Error = err.Error()
		return status
	}
	defer resp.Body.Close()

	status.Reachable = true
	status.HTTPStatus = resp.StatusCode
	var health claudeProxyHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		status.ProxyStatus = "error"
		status.Error = err.Error()
		return status
	}
	status.ProxyStatus = health.Status
	status.TokenHours = health.TokenHours
	status.CCVersion = health.CCVersion
	status.Error = health.Error
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		status.Reachable = false
		if status.Error == "" {
			status.Error = fmt.Sprintf("health returned status %d", resp.StatusCode)
		}
	}
	return status
}

func joinURL(base string, path string) string {
	parsed, err := url.Parse(strings.TrimSpace(base))
	if err != nil {
		return strings.TrimRight(base, "/") + path
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + path
	parsed.RawQuery = ""
	return parsed.String()
}

func readResponsePreview(body io.Reader) string {
	data, err := io.ReadAll(io.LimitReader(body, 1024))
	if err != nil {
		return err.Error()
	}
	return string(data)
}
