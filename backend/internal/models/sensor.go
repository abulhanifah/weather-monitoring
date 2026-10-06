package models

import (
	"time"

	"gorm.io/gorm"
)

// SensorType untuk menyimpan tipe sensor
type SensorType struct {
	ID              uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"name"`
	UnitMeasurement *string        `gorm:"type:varchar(20)" json:"unit_measurement"`
	MinValue        *float64       `json:"min_value"`
	MaxValue        *float64       `json:"max_value"`
	Precission      *int           `json:"precission"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// SensorTypeListResponse response paginated untuk GET /api/v1/sensor-types
type SensorTypeListResponse struct {
	Data      []SensorType `json:"data"`
	Page      int          `json:"page"`
	Limit     int          `json:"limit"`
	Total     int          `json:"total"`
	TotalPage int          `json:"total_page"`
}

// Sensor untuk menyimpan data sensor milik device.
// Status bool: true = aktif, false = nonaktif.
type Sensor struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string         `gorm:"type:varchar(100);not null" json:"name"`
	SensorTypeID uint           `gorm:"index;not null" json:"sensor_type_id"`
	SensorType   *SensorType    `gorm:"foreignKey:SensorTypeID" json:"sensor_type,omitempty"`
	Status       bool           `gorm:"default:true" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// SensorListResponse response paginated untuk GET /api/v1/sensors
type SensorListResponse struct {
	Data      []Sensor `json:"data"`
	Page      int      `json:"page"`
	Limit     int      `json:"limit"`
	Total     int      `json:"total"`
	TotalPage int      `json:"total_page"`
}
