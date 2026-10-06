package models

import "time"

// SensorReading data time series hasil pembacaan sensor.
// Disimpan sebagai hypertable TimescaleDB terpartisi reading_time_origin.
// Index composite (sensor_id, reading_time_origin) untuk dedup & lookup.
type SensorReading struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	DeviceID          string    `gorm:"type:varchar(50);index;not null" json:"device_id"`
	SensorID          uint      `gorm:"index;index:idx_sensor_time,priority:1;not null" json:"sensor_id"`
	ReadingTimeOrigin time.Time `gorm:"index:idx_sensor_time,priority:2;not null" json:"reading_time_origin"`
	Value             float64   `gorm:"not null" json:"value"`
	CreatedAt         time.Time `json:"created_at"`
}
