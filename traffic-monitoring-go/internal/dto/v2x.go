// File: internal/dto/v2x.go
package dto

import (
	"time"
	"traffic-monitoring-go/internal/siem/v2x"
)

// V2XMessageRequest represents a request to process a V2X message
type V2XMessageRequest struct {
	// Raw message data as hex string (for ASN.1 encoded messages)
	RawDataHex string `json:"raw_data_hex,omitempty"`

	// Structured message data (for JSON/testing format)
	MessageData interface{} `json:"message_data,omitempty"`

	// Message source information
	Source string `json:"source,omitempty"`
}

// V2XBatchRequest represents a batch of V2X messages to process
type V2XBatchRequest struct {
	Messages []V2XMessageRequest `json:"messages" binding:"required,min=1,max=1000"`
}

// V2XProcessingResponse represents the result of processing a V2X message
type V2XProcessingResponse struct {
	MessageID        string           `json:"message_id"`
	ProcessingTimeMs float64          `json:"processing_time_ms"`
	ThreatsDetected  int              `json:"threats_detected"`
	AlertsCreated    int              `json:"alerts_created"`
	SecurityEventID  uint             `json:"security_event_id,omitempty"`
	Threats          []ThreatResponse `json:"threats,omitempty"`
	Errors           []string         `json:"errors,omitempty"`
}

// V2XBatchResponse represents the result of processing a batch of V2X messages
type V2XBatchResponse struct {
	ProcessedCount int                     `json:"processed_count"`
	TotalThreats   int                     `json:"total_threats"`
	TotalAlerts    int                     `json:"total_alerts"`
	Results        []V2XProcessingResponse `json:"results"`
}

// ThreatResponse represents a detected threat
type ThreatResponse struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Evidence    map[string]interface{} `json:"evidence,omitempty"`
}

// V2XMetricsResponse represents V2X processing metrics
type V2XMetricsResponse struct {
	MessagesProcessed       uint64          `json:"messages_processed"`
	ThreatsDetected         uint64          `json:"threats_detected"`
	AlertsGenerated         uint64          `json:"alerts_generated"`
	AverageProcessingTimeMs float64         `json:"average_processing_time_ms"`
	LastProcessedTime       time.Time       `json:"last_processed_time"`
	ThreatBreakdown         ThreatBreakdown `json:"threat_breakdown"`
	ActiveVehicles          uint64          `json:"active_vehicles"`
}

// ThreatBreakdown shows counts by threat type
type ThreatBreakdown struct {
	PositionJumpAttacks uint64 `json:"position_jump_attacks"`
	SpeedAnomalies      uint64 `json:"speed_anomalies"`
	MessageAnomalies    uint64 `json:"message_anomalies"`
	BoundaryViolations  uint64 `json:"boundary_violations"`
	ReplayAttacks       uint64 `json:"replay_attacks"`
}

// V2XConfigResponse represents the current V2X detection configuration
type V2XConfigResponse struct {
	MaxPositionJumpKm      float64                   `json:"max_position_jump_km"`
	PositionJumpTimeWindow string                    `json:"position_jump_time_window"`
	MaxSpeedKmh            float64                   `json:"max_speed_kmh"`
	MaxAcceleration        float64                   `json:"max_acceleration"`
	SpeedChangeThreshold   float64                   `json:"speed_change_threshold"`
	MinMessageInterval     string                    `json:"min_message_interval"`
	MaxMessageInterval     string                    `json:"max_message_interval"`
	AllowedBounds          []GeographicBoundResponse `json:"allowed_bounds"`
	MessageCountWindow     uint32                    `json:"message_count_window"`
}

// V2XConfigRequest represents a request to update V2X detection configuration
type V2XConfigRequest struct {
	MaxPositionJumpKm      *float64                  `json:"max_position_jump_km,omitempty"`
	PositionJumpTimeWindow *string                   `json:"position_jump_time_window,omitempty"`
	MaxSpeedKmh            *float64                  `json:"max_speed_kmh,omitempty"`
	MaxAcceleration        *float64                  `json:"max_acceleration,omitempty"`
	SpeedChangeThreshold   *float64                  `json:"speed_change_threshold,omitempty"`
	MinMessageInterval     *string                   `json:"min_message_interval,omitempty"`
	MaxMessageInterval     *string                   `json:"max_message_interval,omitempty"`
	AllowedBounds          *[]GeographicBoundRequest `json:"allowed_bounds,omitempty"`
	MessageCountWindow     *uint32                   `json:"message_count_window,omitempty"`
}

// GeographicBoundResponse represents a geographic boundary in API responses
type GeographicBoundResponse struct {
	Name   string  `json:"name"`
	MinLat float64 `json:"min_lat"`
	MaxLat float64 `json:"max_lat"`
	MinLon float64 `json:"min_lon"`
	MaxLon float64 `json:"max_lon"`
}

// GeographicBoundRequest represents a geographic boundary in API requests
type GeographicBoundRequest struct {
	Name   string  `json:"name" binding:"required"`
	MinLat float64 `json:"min_lat" binding:"required"`
	MaxLat float64 `json:"max_lat" binding:"required"`
	MinLon float64 `json:"min_lon" binding:"required"`
	MaxLon float64 `json:"max_lon" binding:"required"`
}

// VehicleStateResponse represents the current state of a tracked vehicle
type VehicleStateResponse struct {
	VehicleID     string       `json:"vehicle_id"`
	LastPosition  v2x.Position `json:"last_position"`
	LastTimestamp time.Time    `json:"last_timestamp"`
	LastSpeed     float64      `json:"last_speed"`
	LastHeading   float64      `json:"last_heading"`
	MessageCount  uint32       `json:"message_count"`
	HistorySize   int          `json:"history_size"`
}
