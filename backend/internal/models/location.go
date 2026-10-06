package models

import (
	"time"

	"gorm.io/gorm"
)

// Location untuk menyimpan data lokasi
type Location struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// LocationListResponse response paginated untuk GET /api/v1/locations
type LocationListResponse struct {
	Data      []Location `json:"data"`
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
	Total     int        `json:"total"`
	TotalPage int        `json:"total_page"`
}
