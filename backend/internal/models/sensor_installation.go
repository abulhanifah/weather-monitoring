package models

import (
	"time"

	"gorm.io/gorm"
)

// SensorInstallation pemasangan sensor pada device.
// Status bool: true = terpasang aktif, false = dilepas/nonaktif.
type SensorInstallation struct {
	ID        uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID  string         `gorm:"type:varchar(50);index;not null" json:"device_id"`
	Device    *Device        `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
	SensorID  uint           `gorm:"index;not null" json:"sensor_id"`
	Sensor    *Sensor        `gorm:"foreignKey:SensorID" json:"sensor,omitempty"`
	Status    bool           `gorm:"default:true" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}
