package models

import "time"

// DeviceCredentialsResponse response GET /api/v1/devices/{id}/credentials.
type DeviceCredentialsResponse struct {
	DeviceID string       `json:"device_id"`
	Total    int          `json:"total"`
	Data     []APIKeyMeta `json:"data"`
}

// Envelope response untuk dokumentasi swagger (bentuk aktual: {message, data}).
type DeviceCreateEnvelope struct {
	Message string               `json:"message" example:"Device created successfully"`
	Data    CreateDeviceResponse `json:"data"`
}

// Envelope response untuk dokumentasi swagger (bentuk aktual: {message, data}).
type DeviceEnvelope struct {
	Message string `json:"message" example:"Device updated successfully"`
	Data    Device `json:"data"`
}

// Envelope response untuk dokumentasi swagger (bentuk aktual: {message, data}).
type LocationEnvelope struct {
	Message  string   `json:"message" example:"Location created successfully"`
	Data     Location `json:"data"`
}

// Envelope response untuk dokumentasi swagger (bentuk aktual: {message, data}).
type RotateEnvelope struct {
	Message string                 `json:"message" example:"Device credentials rotated successfully"`
	Data    GenerateAPIKeyResponse `json:"data"`
}

// MessageEnvelope response {message} untuk dokumentasi swagger.
type MessageEnvelope struct {
	Message string `json:"message" example:"Device deleted successfully"`
}

// ErrorEnvelope response {error} untuk dokumentasi swagger.
type ErrorEnvelope struct {
	Error string `json:"error" example:"Device not found"`
}

// HealthEnvelope response health check untuk dokumentasi swagger.
type HealthEnvelope struct {
	Status string `json:"status" example:"ok"`
}

// LocationInput body create location untuk dokumentasi swagger.
type LocationInput struct {
	Name string `json:"name" example:"Bogor"`
}

// SensorTypeInput body create sensor type untuk dokumentasi swagger.
type SensorTypeInput struct {
	Name            string   `json:"name" example:"temperature"`
	UnitMeasurement *string  `json:"unit_measurement,omitempty" example:"°C"`
	MinValue        *float64 `json:"min_value,omitempty" example:"-50"`
	MaxValue        *float64 `json:"max_value,omitempty" example:"60"`
	Precission      *int     `json:"precission,omitempty" example:"1"`
}

// SensorInput body create sensor untuk dokumentasi swagger.
type SensorInput struct {
	Name         string `json:"name" example:"Sensor Suhu 1"`
	SensorTypeID uint   `json:"sensor_type_id" example:"1"`
	Status       *bool  `json:"status,omitempty" example:"true"`
}

// SensorPatch body PATCH sensor untuk dokumentasi swagger.
type SensorPatch struct {
	Name         *string `json:"name,omitempty" example:"Sensor Suhu 1"`
	SensorTypeID *uint   `json:"sensor_type_id,omitempty" example:"1"`
	Status       *bool   `json:"status,omitempty" example:"true"`
}

// SensorEnvelope response untuk dokumentasi swagger (bentuk aktual: {message, data}).
type SensorEnvelope struct {
	Message string `json:"message" example:"Sensor created successfully"`
	Data    Sensor `json:"data"`
}

// SensorCalibrationListResponse response GET /api/v1/sensors/{id}/calibrations.
type SensorCalibrationListResponse struct {
	SensorID uint                 `json:"sensor_id"`
	Total    int                  `json:"total"`
	Data     []SensorCalibration  `json:"data"`
}

// SensorCalibrationInput body create kalibrasi untuk dokumentasi swagger.
type SensorCalibrationInput struct {
	Offset   *float64   `json:"offset,omitempty" example:"0.5"`
	Scale    *float64   `json:"scale,omitempty" example:"1"`
	FromDate *time.Time `json:"from_date,omitempty"`
	ToDate   *time.Time `json:"to_date,omitempty"`
}

// SensorCalibrationEnvelope response untuk dokumentasi swagger.
type SensorCalibrationEnvelope struct {
	Message string            `json:"message" example:"Calibration created successfully"`
	Data    SensorCalibration `json:"data"`
}

// SensorInstallationEnvelope response untuk dokumentasi swagger.
type SensorInstallationEnvelope struct {
	Message string             `json:"message" example:"Sensor installed successfully"`
	Data    SensorInstallation `json:"data"`
}

// TelemetryReading satu baris readings telemetry (s = nama tipe, v = value).
type TelemetryReading struct {
	S string  `json:"s" example:"temp_air"`
	V float64 `json:"v" example:"27.4"`
}

// TelemetryBatchItem satu entri batch telemetry.
type TelemetryBatchItem struct {
	Ts        int64              `json:"ts" example:"1757308800"`
	Seq       *int64             `json:"seq,omitempty" example:"10432"`
	BatteryV  *float64           `json:"battery_v,omitempty" example:"3.92"`
	Rssi      *int               `json:"rssi,omitempty" example:"-71"`
	Readings  []TelemetryReading `json:"readings"`
}

// TelemetryBatchRequest payload ingest telemetry batch untuk dokumentasi swagger.
type TelemetryBatchRequest struct {
	DeviceID string               `json:"device_id" example:"WS-GRT-001"`
	Fw       *string              `json:"fw,omitempty" example:"1.4.2"`
	Batch    []TelemetryBatchItem `json:"batch"`
}

// TelemetryRequest payload ingest telemetry untuk dokumentasi swagger.
type TelemetryRequest struct {
	DeviceID  string             `json:"device_id" example:"WS-GRT-001"`
	Fw        *string            `json:"fw,omitempty" example:"1.4.2"`
	Ts        int64              `json:"ts" example:"1757308800"`
	Seq       *int64             `json:"seq,omitempty" example:"10432"`
	BatteryV  *float64           `json:"battery_v,omitempty" example:"3.92"`
	Rssi      *int               `json:"rssi,omitempty" example:"-71"`
	Readings  []TelemetryReading `json:"readings"`
}

// TelemetryResponse response ingest telemetry.
type TelemetryResponse struct {
	Message string `json:"message" example:"OK"`
	Saved   int    `json:"saved" example:"7"`
	Skipped int    `json:"skipped" example:"0"`
}

// SensorTypeEnvelope response untuk dokumentasi swagger (bentuk aktual: {message, data}).
type SensorTypeEnvelope struct {
	Message string     `json:"message" example:"Sensor type created successfully"`
	Data    SensorType `json:"data"`
}

// DevicePatch body PATCH device untuk dokumentasi swagger.
// location_id/longitude/latitude/altitude null untuk mengosongkan.
type DevicePatch struct {
	Name        *string  `json:"name,omitempty" example:"Stasiun Bogor"`
	Description *string  `json:"description,omitempty" example:"Stasiun pemantau cuaca Bogor"`
	Status      *string  `json:"status,omitempty" example:"active"`
	LocationID  *uint    `json:"location_id,omitempty" example:"4"`
	Longitude   *float64 `json:"longitude,omitempty" example:"106.8"`
	Latitude    *float64 `json:"latitude,omitempty" example:"-6.2"`
	Altitude    *float64 `json:"altitude,omitempty" example:"45"`
}

// HeartbeatRequest payload heartbeat.
// device_id dan ts wajib; field lain bebas (object JSON dinamis),
// seluruh payload disimpan ke latest_health.
type HeartbeatRequest struct {
	DeviceID string `json:"device_id" example:"WS-GRT-001"`
	Ts       int64  `json:"ts" example:"1757308920"`
}
