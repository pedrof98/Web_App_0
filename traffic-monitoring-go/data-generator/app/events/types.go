package events

import "time"

// Event represents a security event
type Event struct {
	SourceName string                 `json:"source_name"`
	SourceType string                 `json:"source_type"`
	Timestamp  time.Time              `json:"timestamp"`
	Severity   string                 `json:"severity"`
	Category   string                 `json:"category"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details"`
}

// Event severity levels
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
)

// Event categories
const (
	CategoryAuthentication = "authentication"
	CategoryAuthorization  = "authorization"
	CategoryNetwork        = "network"
	CategoryMalware        = "malware"
	CategorySystem         = "system"
	CategoryVehicle        = "vehicle"
	CategoryV2X            = "v2x"
)
