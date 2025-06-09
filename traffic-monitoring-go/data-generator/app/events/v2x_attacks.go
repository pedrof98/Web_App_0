package events

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

// GenerateV2XPositionJumpAttack creates position jump anomaly events that trigger our SIEM rules
func GenerateV2XPositionJumpAttack(eventCount int, sendFunc func(Event)) {
	vehicleIDs := []string{"VEH-ANOM1", "VEH-ANOM2", "VEH-ANOM3"}

	log.Printf("Generating V2X position jump attack with %d events", eventCount)

	for i := 0; i < eventCount; i++ {
		vehicleID := vehicleIDs[i%len(vehicleIDs)]

		event := Event{
			SourceName: "v2x",
			SourceType: "vehicle",
			Timestamp:  time.Now(),
			Severity:   "high",
			Category:   "v2x",
			Message:    fmt.Sprintf("Position jump anomaly detected for vehicle %s", vehicleID),
			Details: map[string]interface{}{
				"vehicle_id":   vehicleID,
				"message_type": "bsm",
				"protocol":     "DSRC",
				"source_ip":    fmt.Sprintf("192.168.1.%d", 100+i),
				"position": map[string]interface{}{
					"latitude":  37.7749 + rand.Float64()*0.02,
					"longitude": -122.4194 + rand.Float64()*0.02,
				},
				"anomalies": []map[string]interface{}{
					{
						"type":        "position_jump",
						"confidence":  0.75 + rand.Float64()*0.2,
						"description": fmt.Sprintf("Vehicle %s moved impossible distance", vehicleID),
					},
				},
				"attack": "position_jump_simulation",
			},
		}
		sendFunc(event)
		time.Sleep(time.Second * time.Duration(1+rand.Intn(2)))
	}
}

// GenerateV2XMessageFloodAttack creates message flooding events that trigger our SIEM rules
func GenerateV2XMessageFloodAttack(eventCount int, sendFunc func(Event)) {
	attackerVehicle := fmt.Sprintf("VEH-FLOOD-%X", rand.Intn(0xFFFF))

	log.Printf("Generating V2X message flood attack with %d events", eventCount*3)

	for i := 0; i < eventCount*3; i++ {
		event := Event{
			SourceName: "v2x",
			SourceType: "vehicle",
			Timestamp:  time.Now(),
			Severity:   "critical",
			Category:   "v2x",
			Message:    fmt.Sprintf("High frequency messaging detected from %s", attackerVehicle),
			Details: map[string]interface{}{
				"vehicle_id":   attackerVehicle,
				"message_type": "bsm",
				"protocol":     "DSRC",
				"source_ip":    fmt.Sprintf("192.168.1.%d", 100+(i%50)),
				"anomalies": []map[string]interface{}{
					{
						"type":        "high_frequency",
						"confidence":  0.9 + rand.Float64()*0.09,
						"description": "Abnormal message frequency detected",
					},
				},
				"attack": "message_flooding",
			},
		}
		sendFunc(event)
		time.Sleep(time.Millisecond * time.Duration(100+rand.Intn(200)))
	}
}

// GenerateV2XInvalidSignatureAttack creates invalid signature events that trigger our SIEM rules
func GenerateV2XInvalidSignatureAttack(eventCount int, sendFunc func(Event)) {
	log.Printf("Generating V2X invalid signature attack with %d events", eventCount)

	for i := 0; i < eventCount; i++ {
		vehicleID := fmt.Sprintf("VEH-ATTACK-%03d", i)

		event := Event{
			SourceName: "v2x",
			SourceType: "vehicle",
			Timestamp:  time.Now(),
			Severity:   "critical",
			Category:   "v2x",
			Message:    fmt.Sprintf("Invalid digital signature detected from %s", vehicleID),
			Details: map[string]interface{}{
				"vehicle_id":       vehicleID,
				"message_type":     "bsm",
				"protocol":         "DSRC",
				"source_ip":        fmt.Sprintf("192.168.1.%d", 100+i),
				"signature_valid":  false,
				"trust_level":      0,
				"validation_error": "Certificate verification failed",
				"attack":           "invalid_signature",
			},
		}
		sendFunc(event)
		time.Sleep(time.Second * time.Duration(1+rand.Intn(2)))
	}
}

// GenerateV2XSpeedAnomalyAttack creates speed anomaly events that trigger our SIEM rules
func GenerateV2XSpeedAnomalyAttack(eventCount int, sendFunc func(Event)) {
	log.Printf("Generating V2X speed anomaly attack with %d events", eventCount)

	for i := 0; i < eventCount; i++ {
		vehicleID := fmt.Sprintf("VEH-SPEED-%03d", i)

		event := Event{
			SourceName: "v2x",
			SourceType: "vehicle",
			Timestamp:  time.Now(),
			Severity:   "medium",
			Category:   "v2x",
			Message:    fmt.Sprintf("Speed anomaly detected for vehicle %s", vehicleID),
			Details: map[string]interface{}{
				"vehicle_id":   vehicleID,
				"message_type": "bsm",
				"protocol":     "DSRC",
				"source_ip":    fmt.Sprintf("192.168.1.%d", 100+i),
				"speed":        50 + rand.Intn(30), // High speeds
				"anomalies": []map[string]interface{}{
					{
						"type":        "speed_jump",
						"confidence":  0.85 + rand.Float64()*0.1,
						"description": "Unrealistic speed change detected",
					},
				},
				"attack": "speed_anomaly",
			},
		}
		sendFunc(event)
		time.Sleep(time.Second * time.Duration(1+rand.Intn(2)))
	}
}
