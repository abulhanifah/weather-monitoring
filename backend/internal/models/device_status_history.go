package models

import "time"

// DeviceStatusHistory mencatat setiap perubahan status device.
// ID gabungan device_id + timeseries (unix nano).
type DeviceStatusHistory struct {
	ID        string    `gorm:"type:varchar(100);primaryKey" json:"id"`
	DeviceID  string    `gorm:"type:varchar(50);index;not null" json:"device_id"`
	Status    string    `gorm:"type:varchar(20);not null" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
