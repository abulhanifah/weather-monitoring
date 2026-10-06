package constants

// Status Device Constants
const (
	StatusInstalled     = "installed"
	StatusActive        = "active"
	StatusDisconnected  = "disconnected"
	StatusMaintenance   = "maintenance"
	StatusSensorError   = "error"
	StatusDecomissioned = "decomissioned"
)

// Error Codes
const (
	ErrCodeSensorTimeout = "ERR_SENSOR_TIMEOUT"
	ErrCodeInvalidAPIKey = "ERR_INVALID_API_KEY"
	ErrCodeRevokedAPIKey = "ERR_REVOKED_API_KEY"
	ErrCodeUnauthorized  = "ERR_UNAUTHORIZED"
)
