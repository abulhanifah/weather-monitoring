package models

import (
	"time"

	"gorm.io/gorm"
)

// Device untuk menyimpan data device
type Device struct {
	ID          string         `gorm:"type:varchar(50);primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"type:varchar(20);default:'active'" json:"status"`
	LocationID  *uint          `gorm:"index" json:"location_id"`
	Location    *Location      `gorm:"foreignKey:LocationID" json:"location,omitempty"`
	Longitude   *float64       `json:"longitude"`
	Latitude    *float64       `json:"latitude"`
	Altitude    *float64       `json:"altitude"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

type CreateDeviceResponse struct {
	Device *Device    `json:"device"`
	RawKey string     `json:"raw_key"` // Hanya dikembalikan 1x ke client
	Meta   APIKeyMeta `json:"meta"`
}

// APIKeyMeta untuk menyimpan historical api key dari device
type APIKeyMeta struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	DeviceID  string         `gorm:"type:varchar(50)" json:"device_id"`
	Masking   string         `gorm:"type:varchar(20)" json:"masking"`
	Hash      string         `json:"hash"`
	CreatedAt time.Time      `json:"created_at"`
	IsRevoked bool           `json:"is_revoked"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// GenerateAPIKeyResponse response yang dikembalikan ke caller (Device / Admin)
type GenerateAPIKeyResponse struct {
	RawKey string     `json:"raw_key"` // Hanya dikembalikan 1x ke client
	Meta   APIKeyMeta `json:"meta"`
}

// DeviceListResponse response paginated untuk GET /api/v1/devices
type DeviceListResponse struct {
	Data      []Device `json:"data"`
	Page      int      `json:"page"`
	Limit     int      `json:"limit"`
	Total     int      `json:"total"`
	TotalPage int      `json:"total_page"`
}
