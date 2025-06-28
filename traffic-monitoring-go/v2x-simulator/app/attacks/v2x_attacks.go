package attacks

import (
	"fmt"
	"math/rand"
	"time"
)

// AttackType represents different types of V2X attacks
type AttackType string

const (
	PositionJumpAttack     AttackType = "position_jump"
	SpeedJumpAttack        AttackType = "speed_jump"
	MessageFloodingAttack  AttackType = "message_flooding"
	InvalidSignatureAttack AttackType = "invalid_signature"
	ReplayAttack           AttackType = "replay_attack"
	TrustLevelAttack       AttackType = "low_trust_level"
)

// AttackScenario represents a V2X attack scenario
type AttackScenario struct {
	Type          AttackType               `json:"type"`
	Description   string                   `json:"description"`
	Severity      string                   `json:"severity"`
	Anomalies     []map[string]interface{} `json:"anomalies"`
	Modifications map[string]interface{}   `json:"modifications"`
}

// GenerateAttackScenario creates a random attack scenario
func GenerateAttackScenario() *AttackScenario {
	attackTypes := []AttackType{
		PositionJumpAttack,
		SpeedJumpAttack,
		MessageFloodingAttack,
		InvalidSignatureAttack,
		ReplayAttack,
		TrustLevelAttack,
	}

	attackType := attackTypes[rand.Intn(len(attackTypes))]

	switch attackType {
	case PositionJumpAttack:
		return &AttackScenario{
			Type:        PositionJumpAttack,
			Description: "Vehicle position changed unrealistically between messages",
			Severity:    "high",
			Anomalies: []map[string]interface{}{
				{
					"type":       "position_jump",
					"confidence": 0.7 + rand.Float64()*0.3, // 0.7-1.0
					"distance":   100 + rand.Intn(900),     // 100-1000 meters
					"time_diff":  0.1 + rand.Float64()*0.4, // 0.1-0.5 seconds
				},
			},
			Modifications: map[string]interface{}{
				"position_offset": map[string]float64{
					"latitude":  (rand.Float64() - 0.5) * 0.01, // ±0.005 degrees
					"longitude": (rand.Float64() - 0.5) * 0.01,
				},
			},
		}

	case SpeedJumpAttack:
		return &AttackScenario{
			Type:        SpeedJumpAttack,
			Description: "Vehicle speed changed unrealistically between messages",
			Severity:    "medium",
			Anomalies: []map[string]interface{}{
				{
					"type":       "speed_jump",
					"confidence": 0.8 + rand.Float64()*0.2, // 0.8-1.0
					"speed_diff": 15 + rand.Intn(20),       // 15-35 m/s difference
					"time_diff":  0.1 + rand.Float64()*0.2, // 0.1-0.3 seconds
				},
			},
			Modifications: map[string]interface{}{
				"speed_offset": 20 + rand.Intn(30), // +20 to +50 m/s
			},
		}

	case MessageFloodingAttack:
		return &AttackScenario{
			Type:        MessageFloodingAttack,
			Description: "Abnormally high message frequency detected",
			Severity:    "critical",
			Anomalies: []map[string]interface{}{
				{
					"type":       "high_frequency",
					"confidence": 0.9 + rand.Float64()*0.1, // 0.9-1.0
					"frequency":  15 + rand.Intn(10),       // 15-25 msgs/sec
					"threshold":  10,                       // normal threshold
				},
			},
			Modifications: map[string]interface{}{
				"flood_count": 3 + rand.Intn(5), // Send 3-7 extra messages
			},
		}

	case InvalidSignatureAttack:
		return &AttackScenario{
			Type:        InvalidSignatureAttack,
			Description: "Message with invalid digital signature detected",
			Severity:    "critical",
			Anomalies:   []map[string]interface{}{}, // No anomalies array needed
			Modifications: map[string]interface{}{
				"signature_valid": false,
				"signature_error": "Invalid certificate chain",
			},
		}

	case ReplayAttack:
		return &AttackScenario{
			Type:        ReplayAttack,
			Description: "Potential replay attack detected (duplicate message)",
			Severity:    "high",
			Anomalies: []map[string]interface{}{
				{
					"type":       "timing_anomaly",
					"subtype":    "replay_attack",
					"confidence": 0.75 + rand.Float64()*0.25,                   // 0.75-1.0
					"delay":      time.Now().Unix() - int64(60+rand.Intn(300)), // 1-5 minutes old
				},
			},
			Modifications: map[string]interface{}{
				"is_replay":          true,
				"original_timestamp": time.Now().Add(-time.Duration(60+rand.Intn(300)) * time.Second),
			},
		}

	case TrustLevelAttack:
		return &AttackScenario{
			Type:        TrustLevelAttack,
			Description: "Message from vehicle with low trust level",
			Severity:    "medium",
			Anomalies:   []map[string]interface{}{}, // No anomalies array needed
			Modifications: map[string]interface{}{
				"trust_level":  rand.Intn(2), // 0 or 1 (very low trust)
				"trust_reason": "Expired certificate",
			},
		}

	default:
		// Fallback to position jump
		return GenerateAttackScenario()
	}
}

// ApplyAttackToVehicle modifies vehicle data based on attack scenario
func (a *AttackScenario) ApplyAttackToVehicle(vehicle *Vehicle) {
	switch a.Type {
	case PositionJumpAttack:
		if offset, ok := a.Modifications["position_offset"].(map[string]float64); ok {
			vehicle.Latitude += offset["latitude"]
			vehicle.Longitude += offset["longitude"]
		}

	case SpeedJumpAttack:
		if speedOffset, ok := a.Modifications["speed_offset"].(int); ok {
			vehicle.Speed += float32(speedOffset)
			if vehicle.Speed > 80 { // Cap at 80 m/s
				vehicle.Speed = 80
			}
		}

	case MessageFloodingAttack:
		// Flooding is handled at the message sending level

	case InvalidSignatureAttack:
		// Signature modification is handled at the message level

	case ReplayAttack:
		// Replay handling is done at the message level

	case TrustLevelAttack:
		// Trust level modification is handled at the message level
	}
}

// Vehicle represents a simulated vehicle (copied from main.go structure)
type Vehicle struct {
	ID        uint32  `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Speed     float32 `json:"speed"`
	Heading   float32 `json:"heading"`
}

// GetAttackDescription returns a human-readable description of the attack
func (a *AttackScenario) GetAttackDescription() string {
	return fmt.Sprintf("[%s] %s (Severity: %s)", string(a.Type), a.Description, a.Severity)
}

// GenerateSpecificAttackScenario creates a specific attack scenario
func GenerateSpecificAttackScenario(attackType AttackType) *AttackScenario {
	switch attackType {
	case PositionJumpAttack:
		return &AttackScenario{
			Type:        PositionJumpAttack,
			Description: "Vehicle position changed unrealistically between messages",
			Severity:    "high",
			Anomalies: []map[string]interface{}{
				{
					"type":       "position_jump",
					"confidence": 0.7 + rand.Float64()*0.3,
					"distance":   100 + rand.Intn(900),
					"time_diff":  0.1 + rand.Float64()*0.4,
				},
			},
			Modifications: map[string]interface{}{
				"position_offset": map[string]float64{
					"latitude":  (rand.Float64() - 0.5) * 0.01,
					"longitude": (rand.Float64() - 0.5) * 0.01,
				},
			},
		}

	case SpeedJumpAttack:
		return &AttackScenario{
			Type:        SpeedJumpAttack,
			Description: "Vehicle speed changed unrealistically between messages",
			Severity:    "medium",
			Anomalies: []map[string]interface{}{
				{
					"type":       "speed_jump",
					"confidence": 0.8 + rand.Float64()*0.2,
					"speed_diff": 15 + rand.Intn(20),
					"time_diff":  0.1 + rand.Float64()*0.2,
				},
			},
			Modifications: map[string]interface{}{
				"speed_offset": 20 + rand.Intn(30),
			},
		}

	case InvalidSignatureAttack:
		return &AttackScenario{
			Type:        InvalidSignatureAttack,
			Description: "Message with invalid digital signature detected",
			Severity:    "critical",
			Anomalies:   []map[string]interface{}{},
			Modifications: map[string]interface{}{
				"signature_valid": false,
				"signature_error": "Invalid certificate chain",
			},
		}

	case MessageFloodingAttack:
		return &AttackScenario{
			Type:        MessageFloodingAttack,
			Description: "Abnormally high message frequency detected",
			Severity:    "critical",
			Anomalies: []map[string]interface{}{
				{
					"type":       "high_frequency",
					"confidence": 0.9 + rand.Float64()*0.1,
					"frequency":  15 + rand.Intn(10),
					"threshold":  10,
				},
			},
			Modifications: map[string]interface{}{
				"flood_count": 3 + rand.Intn(5),
			},
		}

	case ReplayAttack:
		return &AttackScenario{
			Type:        ReplayAttack,
			Description: "Potential replay attack detected (duplicate message)",
			Severity:    "high",
			Anomalies: []map[string]interface{}{
				{
					"type":       "timing_anomaly",
					"subtype":    "replay_attack",
					"confidence": 0.75 + rand.Float64()*0.25,
					"delay":      time.Now().Unix() - int64(60+rand.Intn(300)),
				},
			},
			Modifications: map[string]interface{}{
				"is_replay":          true,
				"original_timestamp": time.Now().Add(-time.Duration(60+rand.Intn(300)) * time.Second),
			},
		}

	case TrustLevelAttack:
		return &AttackScenario{
			Type:        TrustLevelAttack,
			Description: "Message from vehicle with low trust level",
			Severity:    "medium",
			Anomalies:   []map[string]interface{}{},
			Modifications: map[string]interface{}{
				"trust_level":  rand.Intn(2),
				"trust_reason": "Expired certificate",
			},
		}

	default:
		return GenerateAttackScenario()
	}
}

// GetAttackMetadata returns metadata about the attack for logging/analysis
func (a *AttackScenario) GetAttackMetadata() map[string]interface{} {
	return map[string]interface{}{
		"attack_type":       string(a.Type),
		"severity":          a.Severity,
		"description":       a.Description,
		"anomaly_count":     len(a.Anomalies),
		"has_modifications": len(a.Modifications) > 0,
	}
}
