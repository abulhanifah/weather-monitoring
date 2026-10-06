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

// SensorReadingListResponse response paginated untuk GET /api/v1/readings (raw).
type SensorReadingListResponse struct {
	Data      []SensorReading `json:"data"`
	Page      int             `json:"page"`
	Limit     int             `json:"limit"`
	Total     int             `json:"total"`
	TotalPage int             `json:"total_page"`
}

// AggregatedReading satu bucket agregasi readings.
type AggregatedReading struct {
	Time     time.Time `json:"time"`
	DeviceID string    `json:"device_id"`
	SensorID uint      `json:"sensor_id"`
	Avg      float64   `json:"avg"`
	Min      float64   `json:"min"`
	Max      float64   `json:"max"`
	Count    int       `json:"count"`
}

// AggregatedReadingListResponse response GET /api/v1/readings (interval agregasi).
type AggregatedReadingListResponse struct {
	Data      []AggregatedReading `json:"data"`
	Page      int                 `json:"page"`
	Limit     int                 `json:"limit"`
	Total     int                 `json:"total"`
	TotalPage int                 `json:"total_page"`
	Interval  string              `json:"interval"`
}
