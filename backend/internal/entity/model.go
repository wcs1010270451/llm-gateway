package entity

import (
	"time"

	"gorm.io/datatypes"
)

type Model struct {
	ID                    int64          `gorm:"primaryKey" json:"id"`
	Name                  string         `json:"name"`
	DisplayName           string         `json:"display_name"`
	Family                string         `json:"family"`
	Modality              string         `json:"modality"`
	Status                string         `json:"status"`
	ActiveProviderModelID *int64         `json:"active_provider_model_id,omitempty"`
	Description           string         `json:"description"`
	PricingJSON           datatypes.JSON `gorm:"column:pricing_json" json:"pricing_json"`
	ConfigJSON            datatypes.JSON `gorm:"column:config_json" json:"config_json"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`

	ActiveProviderModel *ProviderModel  `gorm:"foreignKey:ActiveProviderModelID" json:"active_provider_model,omitempty"`
	ProviderModels      []ProviderModel `gorm:"foreignKey:ModelID" json:"provider_models,omitempty"`
}
