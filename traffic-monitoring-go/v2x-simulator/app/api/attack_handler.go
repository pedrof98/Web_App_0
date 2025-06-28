package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"v2x-simulator/app/attacks"
	"v2x-simulator/app/dataset"
)

// AttackTriggerRequest represents the request to trigger specific attacks
type AttackTriggerRequest struct {
	AttackType string `json:"attack_type"`
	Count      int    `json:"count,omitempty"`
	VehicleID  string `json:"vehicle_id,omitempty"`
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
	if req.Count > 50 {
		req.Count = 50
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
				"vehicle_id":   req.VehicleID,
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
		"simulator_running": true,
	}
	respondJSON(w, http.StatusOK, status)
}

// GetAttackScenario creates attack scenario from request
func GetAttackScenario(attackType string) *attacks.AttackScenario {
	switch attackType {
	case "position_jump":
		return attacks.GenerateSpecificAttackScenario(attacks.PositionJumpAttack)
	case "speed_jump":
		return attacks.GenerateSpecificAttackScenario(attacks.SpeedJumpAttack)
	case "invalid_signature":
		return attacks.GenerateSpecificAttackScenario(attacks.InvalidSignatureAttack)
	case "message_flooding":
		return attacks.GenerateSpecificAttackScenario(attacks.MessageFloodingAttack)
	case "replay_attack":
		return attacks.GenerateSpecificAttackScenario(attacks.ReplayAttack)
	case "low_trust_level":
		return attacks.GenerateSpecificAttackScenario(attacks.TrustLevelAttack)
	default:
		return attacks.GenerateAttackScenario() // fallback to random
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
	filename := fmt.Sprintf("/app/datasets/v2x_dataset_%s.csv", timestamp)

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
