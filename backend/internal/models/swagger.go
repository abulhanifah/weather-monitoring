package models

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
