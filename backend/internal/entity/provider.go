package entity

import (
	"time"

	"gorm.io/datatypes"
)

type Provider struct {
	ID              int64          `gorm:"primaryKey" json:"id"`
	Name            string         `json:"name"`
	Slug            string         `json:"slug"`
	Vendor          string         `json:"vendor"`
	AdapterType     string         `json:"adapter_type"`
	AuthType        string         `json:"auth_type"`
	BaseURL         string         `json:"base_url"`
	APIKeyEncrypted string         `json:"-"`
	ConfigJSON      datatypes.JSON `gorm:"column:config_json" json:"config_json"`
	Status          string         `json:"status"`
	Description     string         `json:"description"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}
