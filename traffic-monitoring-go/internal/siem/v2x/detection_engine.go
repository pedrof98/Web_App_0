package v2x

import (
	"context"
	"fmt"
	"math"
	"time"

	"traffic-monitoring-go/internal/domain"
)

// V2XMessage represents a parsed V2X communication message
type V2XMessage struct {
	ID           string
	VehicleID    string
	Timestamp    time.Time
	MessageType  V2XMessageType
	Position     Position
	Speed        float64 // km/h
	Heading      float64 // degrees
	Acceleration float64 // m/s²
	MessageCount uint32  // sequential message counter
	RawData      []byte
}

type V2XMessageType string

const (
	BSMMessage  V2XMessageType = "BSM"  // Basic Safety Message
	SPaTMessage V2XMessageType = "SPaT" // Signal Phase and Timing
	MAPMessage  V2XMessageType = "MAP"  // Map Data
	CAMMessage  V2XMessageType = "CAM"  // Cooperative Awareness Message
	DENMMessage V2XMessageType = "DENM" // Decentralized Environmental Notification Message
)

type Position struct {
	Latitude  float64
	Longitude float64
	Elevation float64
}

// VehicleState tracks the state of a vehicle over time
type VehicleState struct {
	VehicleID      string
	LastPosition   Position
	LastTimestamp  time.Time
	LastSpeed      float64
	LastHeading    float64
	MessageCount   uint32
	MessageHistory []V2XMessage
	MaxHistorySize int
}

// V2XDetectionEngine implements specific V2X threat detection rules
type V2XDetectionEngine struct {
	vehicleStates map[string]*VehicleState
	config        *DetectionConfig
}

type DetectionConfig struct {
	// Position Jump Detection
	MaxPositionJumpKm      float64       // Maximum allowed position change in km
	PositionJumpTimeWindow time.Duration // Time window for position jump detection

	// Speed Anomaly Detection
	MaxSpeedKmh          float64 // Maximum realistic speed in km/h
	MaxAcceleration      float64 // Maximum realistic acceleration in m/s²
	SpeedChangeThreshold float64 // Percentage change threshold for speed anomalies

	// Message Frequency Detection
	MinMessageInterval time.Duration // Minimum expected time between messages
	MaxMessageInterval time.Duration // Maximum expected time between messages

	// Geographic Boundaries
	AllowedBounds []GeographicBound

	// Replay Attack Detection
	MessageCountWindow uint32 // Window for detecting message count anomalies
}

type GeographicBound struct {
	Name   string
	MinLat float64
	MaxLat float64
	MinLon float64
	MaxLon float64
}

// DetectionResult represents the result of a threat detection analysis
type DetectionResult struct {
	ThreatDetected bool
	ThreatType     string
	Severity       domain.EventSeverity
	Description    string
	Evidence       map[string]interface{}
	Confidence     float64 // 0.0 to 1.0
}

func NewV2XDetectionEngine() *V2XDetectionEngine {
	return &V2XDetectionEngine{
		vehicleStates: make(map[string]*VehicleState),
		config: &DetectionConfig{
			MaxPositionJumpKm:      50.0, // 50km position jump
			PositionJumpTimeWindow: 10 * time.Second,
			MaxSpeedKmh:            300.0,                 // 300 km/h max speed
			MaxAcceleration:        15.0,                  // 15 m/s² max acceleration
			SpeedChangeThreshold:   200.0,                 // 200% speed change
			MinMessageInterval:     50 * time.Millisecond, // 20 Hz minimum
			MaxMessageInterval:     2 * time.Second,       // 0.5 Hz maximum
			MessageCountWindow:     100,
		},
	}
}

// ProcessMessage analyzes a V2X message for potential threats
func (engine *V2XDetectionEngine) ProcessMessage(ctx context.Context, msg *V2XMessage) []DetectionResult {
	var results []DetectionResult

	// Get or create vehicle state
	state := engine.getOrCreateVehicleState(msg.VehicleID)

	// Run all detection rules
	if result := engine.detectPositionJump(msg, state); result.ThreatDetected {
		results = append(results, result)
	}

	if result := engine.detectSpeedAnomaly(msg, state); result.ThreatDetected {
		results = append(results, result)
	}

	if result := engine.detectMessageFrequencyAnomaly(msg, state); result.ThreatDetected {
		results = append(results, result)
	}

	if result := engine.detectGeographicBoundaryViolation(msg); result.ThreatDetected {
		results = append(results, result)
	}

	if result := engine.detectReplayAttack(msg, state); result.ThreatDetected {
		results = append(results, result)
	}

	// Update vehicle state
	engine.updateVehicleState(msg, state)

	return results
}

// detectPositionJump detects unrealistic position changes (teleportation attacks)
func (engine *V2XDetectionEngine) detectPositionJump(msg *V2XMessage, state *VehicleState) DetectionResult {
	if state.LastTimestamp.IsZero() {
		return DetectionResult{ThreatDetected: false}
	}

	timeDiff := msg.Timestamp.Sub(state.LastTimestamp)
	if timeDiff > engine.config.PositionJumpTimeWindow {
		return DetectionResult{ThreatDetected: false} // Too much time passed
	}

	distance := calculateDistance(state.LastPosition, msg.Position)

	if distance > engine.config.MaxPositionJumpKm {
		return DetectionResult{
			ThreatDetected: true,
			ThreatType:     "position_jump_attack",
			Severity:       domain.SeverityHigh,
			Description: fmt.Sprintf("Vehicle %s jumped %.2f km in %.2f seconds",
				msg.VehicleID, distance, timeDiff.Seconds()),
			Evidence: map[string]interface{}{
				"distance_km":       distance,
				"time_seconds":      timeDiff.Seconds(),
				"previous_position": state.LastPosition,
				"current_position":  msg.Position,
				"threshold_km":      engine.config.MaxPositionJumpKm,
			},
			Confidence: 0.95,
		}
	}

	return DetectionResult{ThreatDetected: false}
}

// detectSpeedAnomaly detects unrealistic speed values or changes
func (engine *V2XDetectionEngine) detectSpeedAnomaly(msg *V2XMessage, state *VehicleState) DetectionResult {
	// Check for impossible speeds
	if msg.Speed > engine.config.MaxSpeedKmh {
		return DetectionResult{
			ThreatDetected: true,
			ThreatType:     "impossible_speed",
			Severity:       domain.SeverityHigh,
			Description: fmt.Sprintf("Vehicle %s reported impossible speed: %.2f km/h",
				msg.VehicleID, msg.Speed),
			Evidence: map[string]interface{}{
				"reported_speed": msg.Speed,
				"max_realistic":  engine.config.MaxSpeedKmh,
			},
			Confidence: 0.98,
		}
	}

	// Check for unrealistic speed changes
	if !state.LastTimestamp.IsZero() {
		timeDiff := msg.Timestamp.Sub(state.LastTimestamp).Seconds()
		if timeDiff > 0 {
			speedChange := math.Abs(msg.Speed - state.LastSpeed)
			maxSpeedChange := engine.config.SpeedChangeThreshold / 100.0 * state.LastSpeed

			if speedChange > maxSpeedChange && speedChange > 50 { // minimum 50 km/h change
				return DetectionResult{
					ThreatDetected: true,
					ThreatType:     "speed_anomaly",
					Severity:       domain.SeverityMedium,
					Description: fmt.Sprintf("Vehicle %s had suspicious speed change: %.2f to %.2f km/h",
						msg.VehicleID, state.LastSpeed, msg.Speed),
					Evidence: map[string]interface{}{
						"previous_speed": state.LastSpeed,
						"current_speed":  msg.Speed,
						"speed_change":   speedChange,
						"time_diff":      timeDiff,
					},
					Confidence: 0.75,
				}
			}
		}
	}

	return DetectionResult{ThreatDetected: false}
}

// detectMessageFrequencyAnomaly detects unusual message timing patterns
func (engine *V2XDetectionEngine) detectMessageFrequencyAnomaly(msg *V2XMessage, state *VehicleState) DetectionResult {
	if state.LastTimestamp.IsZero() {
		return DetectionResult{ThreatDetected: false}
	}

	interval := msg.Timestamp.Sub(state.LastTimestamp)

	if interval < engine.config.MinMessageInterval {
		return DetectionResult{
			ThreatDetected: true,
			ThreatType:     "message_flooding",
			Severity:       domain.SeverityMedium,
			Description: fmt.Sprintf("Vehicle %s sending messages too frequently: %.2f ms interval",
				msg.VehicleID, interval.Seconds()*1000),
			Evidence: map[string]interface{}{
				"message_interval_ms": interval.Seconds() * 1000,
				"min_allowed_ms":      engine.config.MinMessageInterval.Seconds() * 1000,
			},
			Confidence: 0.85,
		}
	}

	if interval > engine.config.MaxMessageInterval {
		return DetectionResult{
			ThreatDetected: true,
			ThreatType:     "message_dropout",
			Severity:       domain.SeverityLow,
			Description: fmt.Sprintf("Vehicle %s has suspicious message gap: %.2f seconds",
				msg.VehicleID, interval.Seconds()),
			Evidence: map[string]interface{}{
				"message_interval_s": interval.Seconds(),
				"max_allowed_s":      engine.config.MaxMessageInterval.Seconds(),
			},
			Confidence: 0.60,
		}
	}

	return DetectionResult{ThreatDetected: false}
}

// detectGeographicBoundaryViolation detects vehicles outside allowed areas
func (engine *V2XDetectionEngine) detectGeographicBoundaryViolation(msg *V2XMessage) DetectionResult {
	if len(engine.config.AllowedBounds) == 0 {
		return DetectionResult{ThreatDetected: false} // No bounds configured
	}

	for _, bound := range engine.config.AllowedBounds {
		if isWithinBounds(msg.Position, bound) {
			return DetectionResult{ThreatDetected: false} // Within allowed area
		}
	}

	return DetectionResult{
		ThreatDetected: true,
		ThreatType:     "geographic_boundary_violation",
		Severity:       domain.SeverityMedium,
		Description:    fmt.Sprintf("Vehicle %s outside authorized geographic area", msg.VehicleID),
		Evidence: map[string]interface{}{
			"position":       msg.Position,
			"allowed_bounds": engine.config.AllowedBounds,
		},
		Confidence: 0.90,
	}
}

// detectReplayAttack detects message replay attacks through sequence analysis
func (engine *V2XDetectionEngine) detectReplayAttack(msg *V2XMessage, state *VehicleState) DetectionResult {
	if state.MessageCount == 0 {
		return DetectionResult{ThreatDetected: false}
	}

	// Check for message count going backwards (replay)
	if msg.MessageCount <= state.MessageCount {
		return DetectionResult{
			ThreatDetected: true,
			ThreatType:     "replay_attack",
			Severity:       domain.SeverityCritical,
			Description: fmt.Sprintf("Vehicle %s message replay detected: count %d <= previous %d",
				msg.VehicleID, msg.MessageCount, state.MessageCount),
			Evidence: map[string]interface{}{
				"current_count":  msg.MessageCount,
				"previous_count": state.MessageCount,
			},
			Confidence: 0.95,
		}
	}

	// Check for unrealistic message count jumps
	countJump := msg.MessageCount - state.MessageCount
	if countJump > engine.config.MessageCountWindow {
		return DetectionResult{
			ThreatDetected: true,
			ThreatType:     "message_count_anomaly",
			Severity:       domain.SeverityMedium,
			Description: fmt.Sprintf("Vehicle %s suspicious message count jump: %d",
				msg.VehicleID, countJump),
			Evidence: map[string]interface{}{
				"count_jump": countJump,
				"max_window": engine.config.MessageCountWindow,
			},
			Confidence: 0.70,
		}
	}

	return DetectionResult{ThreatDetected: false}
}

// Helper functions

func (engine *V2XDetectionEngine) getOrCreateVehicleState(vehicleID string) *VehicleState {
	if state, exists := engine.vehicleStates[vehicleID]; exists {
		return state
	}

	state := &VehicleState{
		VehicleID:      vehicleID,
		MessageHistory: make([]V2XMessage, 0),
		MaxHistorySize: 100,
	}
	engine.vehicleStates[vehicleID] = state
	return state
}

func (engine *V2XDetectionEngine) updateVehicleState(msg *V2XMessage, state *VehicleState) {
	state.LastPosition = msg.Position
	state.LastTimestamp = msg.Timestamp
	state.LastSpeed = msg.Speed
	state.LastHeading = msg.Heading
	state.MessageCount = msg.MessageCount

	// Add to history
	state.MessageHistory = append(state.MessageHistory, *msg)
	if len(state.MessageHistory) > state.MaxHistorySize {
		state.MessageHistory = state.MessageHistory[1:]
	}
}

// calculateDistance calculates the distance between two positions in kilometers
func calculateDistance(pos1, pos2 Position) float64 {
	const earthRadius = 6371.0 // Earth's radius in kilometers

	lat1Rad := pos1.Latitude * math.Pi / 180
	lat2Rad := pos2.Latitude * math.Pi / 180
	deltaLat := (pos2.Latitude - pos1.Latitude) * math.Pi / 180
	deltaLon := (pos2.Longitude - pos1.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func isWithinBounds(pos Position, bound GeographicBound) bool {
	return pos.Latitude >= bound.MinLat && pos.Latitude <= bound.MaxLat &&
		pos.Longitude >= bound.MinLon && pos.Longitude <= bound.MaxLon
}
