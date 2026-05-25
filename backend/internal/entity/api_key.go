package entity

import "time"

type APIKey struct {
	ID                int64      `gorm:"primaryKey" json:"id"`
	UserID            int64      `json:"user_id"`
	Name              string     `json:"name"`
	KeyHash           string     `json:"-"`
	KeyEncrypted      string     `gorm:"column:plain_key" json:"-"`
	MaskedKey         string     `json:"masked_key"`
	Status            string     `json:"status"`
	RPMLimit          int        `json:"rpm_limit"`
	DailyRequestLimit int        `json:"daily_request_limit"`
	DailyTokenLimit   int        `json:"daily_token_limit"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	LastUsedAt        *time.Time `json:"last_used_at,omitempty"`
	LastErrorAt       *time.Time `json:"last_error_at,omitempty"`
	LastErrorMessage  string     `json:"last_error_message"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
