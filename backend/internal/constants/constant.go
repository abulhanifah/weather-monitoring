package constants

// Status Device Constants
const (
	StatusActive      = "active"
	StatusOffline     = "offline"
	StatusDegraded    = "degraded"
	StatusMaintenance = "maintenance"
	StatusFaulty      = "faulty"
	StatusDisabled    = "disabled"
)

// Error Codes
const (
	ErrCodeSensorTimeout = "ERR_SENSOR_TIMEOUT"
	ErrCodeInvalidAPIKey = "ERR_INVALID_API_KEY"
	ErrCodeRevokedAPIKey = "ERR_REVOKED_API_KEY"
	ErrCodeUnauthorized  = "ERR_UNAUTHORIZED"
)
