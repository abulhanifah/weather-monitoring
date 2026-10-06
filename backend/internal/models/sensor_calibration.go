package models

import (
	"time"

	"gorm.io/gorm"
)

// SensorCalibration kalibrasi sensor: nilai terkoreksi = (raw + offset) * scale.
// FromDate default now (autoCreateTime), ToDate default null (masih berlaku).
type SensorCalibration struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	SensorID  uint           `gorm:"index;not null" json:"sensor_id"`
	Offset    *float64       `json:"offset"`
	Scale     *float64       `json:"scale"`
	FromDate  time.Time      `gorm:"autoCreateTime" json:"from_date"`
	ToDate    *time.Time     `json:"to_date"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
