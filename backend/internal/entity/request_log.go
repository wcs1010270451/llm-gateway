package entity

import (
	"time"

	"gorm.io/datatypes"
)

type RequestLog struct {
	ID               int64          `gorm:"primaryKey" json:"id"`
	RequestID        string         `json:"request_id"`
	TraceID          string         `json:"trace_id"`
	UserID           *int64         `json:"user_id,omitempty"`
	APIKeyID         *int64         `gorm:"column:api_key_id" json:"api_key_id,omitempty"`
	ModelID          *int64         `json:"model_id,omitempty"`
	PublicModelName  string         `json:"public_model_name"`
	ProviderID       *int64         `json:"provider_id,omitempty"`
	ProviderModelID  *int64         `json:"provider_model_id,omitempty"`
	AdapterType      string         `json:"adapter_type"`
	UpstreamModel    string         `json:"upstream_model"`
	RequestType      string         `json:"request_type"`
	Stream           bool           `json:"stream"`
	ClientIP         string         `json:"client_ip"`
	RequestMethod    string         `json:"request_method"`
	RequestPath      string         `json:"request_path"`
	HTTPStatus       int            `json:"http_status"`
	Success          bool           `json:"success"`
	LatencyMS        int            `json:"latency_ms"`
	PromptTokens     int            `json:"prompt_tokens"`
	CompletionTokens int            `json:"completion_tokens"`
	TotalTokens      int            `json:"total_tokens"`
	EstimatedCost    float64        `json:"estimated_cost"`
	ErrorType        string         `json:"error_type"`
	ErrorMessage     string         `json:"error_message"`
	RequestPreview   datatypes.JSON `gorm:"column:request_preview" json:"request_preview"`
	ResponsePreview  datatypes.JSON `gorm:"column:response_preview" json:"response_preview"`
	Metadata         datatypes.JSON `json:"metadata"`
	CreatedAt        time.Time      `json:"created_at"`
	User             *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type KeyModelUsageStat struct {
	PublicModelName  string  `json:"public_model_name"`
	RequestCount     int64   `json:"request_count"`
	PromptTokens     int64   `json:"prompt_tokens"`
	CompletionTokens int64   `json:"completion_tokens"`
	TotalTokens      int64   `json:"total_tokens"`
	EstimatedCost    float64 `json:"estimated_cost"`
}

type RequestUsageSummary struct {
	RequestCount     int64   `json:"request_count"`
	SuccessCount     int64   `json:"success_count"`
	ActiveUserCount  int64   `json:"active_user_count"`
	ActiveKeyCount   int64   `json:"active_key_count"`
	TotalTokens      int64   `json:"total_tokens"`
	AverageLatencyMS float64 `json:"average_latency_ms"`
	EstimatedCost    float64 `json:"estimated_cost"`
}

type ProviderUsagePoint struct {
	Period        time.Time `json:"period"`
	RequestCount  int64     `json:"request_count"`
	TotalTokens   int64     `json:"total_tokens"`
	EstimatedCost float64   `json:"estimated_cost"`
}

type ProviderModelUsageStat struct {
	ProviderModelID  *int64     `json:"provider_model_id,omitempty"`
	UpstreamModel    string     `json:"upstream_model"`
	PublicModelName  string     `json:"public_model_name"`
	RequestCount     int64      `json:"request_count"`
	SuccessCount     int64      `json:"success_count"`
	PromptTokens     int64      `json:"prompt_tokens"`
	CompletionTokens int64      `json:"completion_tokens"`
	TotalTokens      int64      `json:"total_tokens"`
	EstimatedCost    float64    `json:"estimated_cost"`
	AverageLatencyMS float64    `json:"average_latency_ms"`
	LastUsedAt       *time.Time `json:"last_used_at,omitempty"`
}
