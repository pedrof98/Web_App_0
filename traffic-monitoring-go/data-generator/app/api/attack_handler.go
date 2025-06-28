package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"traffic-monitoring-go/data-generator/app/dataset"
	"traffic-monitoring-go/data-generator/app/events"
)

// AttackTriggerRequest represents the request to trigger specific attacks
type AttackTriggerRequest struct {
	AttackType string `json:"attack_type"`
	Count      int    `json:"count,omitempty"`
	Severity   string `json:"severity,omitempty"`
}

// AttackTriggerResponse represents the response from triggering attacks
type AttackTriggerResponse struct {
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	AttackType  string                 `json:"attack_type"`
	EventsCount int                    `json:"events_count"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

var (
	AttackTriggerChan = make(chan AttackTriggerRequest, 10)
	validAttackTypes  = []string{
		"position_jump", "speed_jump", "invalid_signature",
		"message_flooding", "replay_attack", "low_trust_level",
	}
	DatasetExporter = dataset.NewDatasetExporter()
)

// HandleAttackTrigger handles POST /attack/trigger requests
func HandleAttackTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AttackTriggerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Validate attack type
	if !isValidAttackType(req.AttackType) {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid attack type. Valid types: %v", validAttackTypes))
		return
	}

	// Set defaults and limits
	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 100 {
		req.Count = 100
	}

	// Send trigger request (non-blocking)
	select {
	case AttackTriggerChan <- req:
		response := AttackTriggerResponse{
			Success:     true,
			Message:     fmt.Sprintf("Successfully triggered %s attack", req.AttackType),
			AttackType:  req.AttackType,
			EventsCount: req.Count,
			Details: map[string]interface{}{
				"trigger_time": time.Now().Format(time.RFC3339),
				"severity":     req.Severity,
			},
		}
		respondJSON(w, http.StatusOK, response)
	default:
		respondError(w, http.StatusServiceUnavailable, "Attack trigger queue is full, please try again")
	}
}

// HandleAttackStatus handles GET /attack/status requests
func HandleAttackStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := map[string]interface{}{
		"available_attacks": validAttackTypes,
		"queue_capacity":    cap(AttackTriggerChan),
		"queue_length":      len(AttackTriggerChan),
		"generator_running": true,
	}
	respondJSON(w, http.StatusOK, status)
}

// GetAttackEvent creates specific attack event
func GetAttackEvent(attackType string, severity string) events.Event {
	switch attackType {
	case "position_jump":
		return generatePositionJumpAttack(severity)
	case "speed_jump":
		return generateSpeedJumpAttack(severity)
	case "invalid_signature":
		return generateInvalidSignatureAttack(severity)
	case "message_flooding":
		return generateMessageFloodingAttack(severity)
	case "replay_attack":
		return generateReplayAttack(severity)
	case "low_trust_level":
		return generateLowTrustLevelAttack(severity)
	default:
		return generatePositionJumpAttack(severity) // fallback
	}
}

// Helper functions
func isValidAttackType(attackType string) bool {
	for _, valid := range validAttackTypes {
		if attackType == valid {
			return true
		}
	}
	return false
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// Attack generation functions
func generatePositionJumpAttack(severity string) events.Event {
	if severity == "" {
		severity = "high"
	}

	return events.Event{
		SourceName: "v2x",
		SourceType: "vehicle",
		Timestamp:  time.Now(),
		Severity:   severity,
		Category:   "v2x",
		Message:    fmt.Sprintf("Position jump attack detected from vehicle VEH-%06d", rand.Intn(999999)),
		Details: map[string]interface{}{
			"vehicle_id":   fmt.Sprintf("VEH-%06d", rand.Intn(999999)),
			"message_type": "bsm",
			"protocol":     "DSRC",
			"source_ip":    fmt.Sprintf("192.168.1.%d", rand.Intn(254)+1),
			"anomalies": []map[string]interface{}{
				{
					"type":       "position_jump",
					"confidence": 0.85 + rand.Float64()*0.15,
					"distance":   200 + rand.Intn(800),
					"time_diff":  0.2 + rand.Float64()*0.3,
				},
			},
			"attack": "position_jump",
		},
	}
}

func generateSpeedJumpAttack(severity string) events.Event {
	if severity == "" {
		severity = "medium"
	}

	return events.Event{
		SourceName: "v2x",
		SourceType: "vehicle",
		Timestamp:  time.Now(),
		Severity:   severity,
		Category:   "v2x",
		Message:    fmt.Sprintf("Speed jump attack detected from vehicle VEH-%06d", rand.Intn(999999)),
		Details: map[string]interface{}{
			"vehicle_id":   fmt.Sprintf("VEH-%06d", rand.Intn(999999)),
			"message_type": "bsm",
			"protocol":     "DSRC",
			"source_ip":    fmt.Sprintf("192.168.1.%d", rand.Intn(254)+1),
			"anomalies": []map[string]interface{}{
				{
					"type":       "speed_jump",
					"confidence": 0.8 + rand.Float64()*0.2,
					"speed_diff": 25 + rand.Intn(40),
					"time_diff":  0.1 + rand.Float64()*0.2,
				},
			},
			"attack": "speed_jump",
		},
	}
}

func generateInvalidSignatureAttack(severity string) events.Event {
	if severity == "" {
		severity = "critical"
	}

	return events.Event{
		SourceName: "v2x",
		SourceType: "vehicle",
		Timestamp:  time.Now(),
		Severity:   severity,
		Category:   "v2x",
		Message:    fmt.Sprintf("Invalid signature attack detected from vehicle VEH-%06d", rand.Intn(999999)),
		Details: map[string]interface{}{
			"vehicle_id":      fmt.Sprintf("VEH-%06d", rand.Intn(999999)),
			"message_type":    "bsm",
			"protocol":        "DSRC",
			"source_ip":       fmt.Sprintf("192.168.1.%d", rand.Intn(254)+1),
			"signature_valid": false,
			"signature_error": "Invalid certificate chain",
			"attack":          "invalid_signature",
		},
	}
}

func generateMessageFloodingAttack(severity string) events.Event {
	if severity == "" {
		severity = "critical"
	}

	return events.Event{
		SourceName: "v2x",
		SourceType: "vehicle",
		Timestamp:  time.Now(),
		Severity:   severity,
		Category:   "v2x",
		Message:    fmt.Sprintf("Message flooding attack detected from vehicle VEH-%06d", rand.Intn(999999)),
		Details: map[string]interface{}{
			"vehicle_id":   fmt.Sprintf("VEH-%06d", rand.Intn(999999)),
			"message_type": "bsm",
			"protocol":     "DSRC",
			"source_ip":    fmt.Sprintf("192.168.1.%d", rand.Intn(254)+1),
			"anomalies": []map[string]interface{}{
				{
					"type":       "high_frequency",
					"confidence": 0.9 + rand.Float64()*0.1,
					"frequency":  20 + rand.Intn(15),
					"threshold":  10,
				},
			},
			"attack": "message_flooding",
		},
	}
}

func generateReplayAttack(severity string) events.Event {
	if severity == "" {
		severity = "high"
	}

	return events.Event{
		SourceName: "v2x",
		SourceType: "vehicle",
		Timestamp:  time.Now(),
		Severity:   severity,
		Category:   "v2x",
		Message:    fmt.Sprintf("Replay attack detected from vehicle VEH-%06d", rand.Intn(999999)),
		Details: map[string]interface{}{
			"vehicle_id":   fmt.Sprintf("VEH-%06d", rand.Intn(999999)),
			"message_type": "bsm",
			"protocol":     "DSRC",
			"source_ip":    fmt.Sprintf("192.168.1.%d", rand.Intn(254)+1),
			"anomalies": []map[string]interface{}{
				{
					"type":       "timing_anomaly",
					"subtype":    "replay_attack",
					"confidence": 0.75 + rand.Float64()*0.25,
					"delay":      60 + rand.Intn(300),
				},
			},
			"is_replay":          true,
			"original_timestamp": time.Now().Add(-time.Duration(60+rand.Intn(300)) * time.Second),
			"attack":             "replay_attack",
		},
	}
}

func generateLowTrustLevelAttack(severity string) events.Event {
	if severity == "" {
		severity = "medium"
	}

	return events.Event{
		SourceName: "v2x",
		SourceType: "vehicle",
		Timestamp:  time.Now(),
		Severity:   severity,
		Category:   "v2x",
		Message:    fmt.Sprintf("Low trust level vehicle detected: VEH-%06d", rand.Intn(999999)),
		Details: map[string]interface{}{
			"vehicle_id":   fmt.Sprintf("VEH-%06d", rand.Intn(999999)),
			"message_type": "bsm",
			"protocol":     "DSRC",
			"source_ip":    fmt.Sprintf("192.168.1.%d", rand.Intn(254)+1),
			"trust_level":  rand.Intn(2), // 0 or 1 (very low trust)
			"trust_reason": "Expired certificate",
			"attack":       "low_trust_level",
		},
	}
}

// HandleDatasetStart handles POST /dataset/start
func HandleDatasetStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	DatasetExporter.Enable()
	DatasetExporter.ClearData() // Start fresh

	response := map[string]interface{}{
		"success":   true,
		"message":   "Dataset collection started",
		"timestamp": time.Now().Format(time.RFC3339),
	}
	respondJSON(w, http.StatusOK, response)
}

// HandleDatasetStop handles POST /dataset/stop
func HandleDatasetStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	DatasetExporter.Disable()
	stats := DatasetExporter.GetStats()

	response := map[string]interface{}{
		"success":   true,
		"message":   "Dataset collection stopped",
		"stats":     stats,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	respondJSON(w, http.StatusOK, response)
}

// HandleDatasetExport handles GET /dataset/export
func HandleDatasetExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("/app/datasets/data_generator_dataset_%s.csv", timestamp)

	// Export to CSV
	if err := DatasetExporter.ExportToCSV(filename); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("Export failed: %v", err))
		return
	}

	stats := DatasetExporter.GetStats()
	response := map[string]interface{}{
		"success":   true,
		"message":   "Dataset exported successfully",
		"filename":  filename,
		"stats":     stats,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	respondJSON(w, http.StatusOK, response)
}

// HandleDatasetStatus handles GET /dataset/status
func HandleDatasetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := DatasetExporter.GetStats()
	respondJSON(w, http.StatusOK, stats)
}
