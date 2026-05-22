package entity

import (
	"time"

	"gorm.io/datatypes"
)

type ProviderModel struct {
	ID              int64          `gorm:"primaryKey" json:"id"`
	ProviderID      int64          `json:"provider_id"`
	ModelID         *int64         `json:"model_id,omitempty"`
	UpstreamModel   string         `json:"upstream_model"`
	Status          string         `json:"status"`
	MaxTokens       int            `json:"max_tokens"`
	TimeoutSeconds  int            `json:"timeout_seconds"`
	InputCostPer1M  float64        `gorm:"column:input_cost_per_1m" json:"input_cost_per_1m"`
	OutputCostPer1M float64        `gorm:"column:output_cost_per_1m" json:"output_cost_per_1m"`
	PricingJSON     datatypes.JSON `gorm:"column:pricing_json" json:"pricing_json"`
	ConfigJSON      datatypes.JSON `gorm:"column:config_json" json:"config_json"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	Provider Provider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	Model    *Model   `gorm:"foreignKey:ModelID" json:"model,omitempty"`
}
