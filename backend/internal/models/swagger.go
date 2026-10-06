package models

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
// location_id null untuk melepas lokasi.
type DevicePatch struct {
	Name        *string `json:"name,omitempty" example:"Stasiun Bogor"`
	Description *string `json:"description,omitempty" example:"Stasiun pemantau cuaca Bogor"`
	Status      *string `json:"status,omitempty" example:"active"`
	LocationID  *uint   `json:"location_id,omitempty" example:"4"`
}
