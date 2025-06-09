package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

// SIEMSender handles sending V2X events to the SIEM system
type SIEMSender struct {
	endpoint   string
	httpClient *http.Client
}

// NewSIEMSender creates a new SIEM sender
func NewSIEMSender(endpoint string) *SIEMSender {
	return &SIEMSender{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// V2XEvent represents a V2X event to be sent to SIEM
type V2XEvent struct {
	SourceName string                 `json:"source_name"`
	SourceType string                 `json:"source_type"`
	Timestamp  time.Time              `json:"timestamp"`
	Severity   string                 `json:"severity"`
	Category   string                 `json:"category"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details"`
}

// SendV2XEvent sends a V2X event to the SIEM ingestion endpoint
func (s *SIEMSender) SendV2XEvent(event *V2XEvent) error {
	// Convert event to JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", s.endpoint, bytes.NewBuffer(eventJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SIEM returned status %d", resp.StatusCode)
	}

	return nil
}

// TestConnection tests connectivity to the SIEM endpoint
func (s *SIEMSender) TestConnection() error {
	// Send a simple GET request to check if SIEM is available
	healthEndpoint := s.endpoint[:len(s.endpoint)-len("/ingest")] + "/health"

	resp, err := s.httpClient.Get(healthEndpoint)
	if err != nil {
		return fmt.Errorf("SIEM not reachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SIEM health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// CreateV2XMessageEvent creates a V2X event from vehicle data and message type
func CreateV2XMessageEvent(vehicleID uint32, messageType string, latitude, longitude float64, speed float32, protocol string) *V2XEvent {
	return &V2XEvent{
		SourceName: "v2x-simulator",
		SourceType: "v2x",
		Timestamp:  time.Now(),
		Severity:   "info", // Default severity for normal messages
		Category:   "v2x",
		Message:    fmt.Sprintf("V2X %s message from vehicle VEH%03d", messageType, vehicleID%1000),
		Details: map[string]interface{}{
			"vehicle_id":       fmt.Sprintf("VEH%03d", vehicleID%1000),
			"message_type":     messageType,
			"protocol":         protocol,
			"location":         fmt.Sprintf("%.6f,%.6f", latitude, longitude),
			"speed":            speed,
			"signature_valid":  true, // Default to valid signature
			"trust_level":      8,    // Default to high trust level
			"source_ip":        generateRandomIP(),
			"source_port":      generateRandomPort(),
			"destination_ip":   generateRandomIP(),
			"destination_port": generateRandomPort(),
		},
	}
}

// CreateAttackEvent creates a V2X event with attack characteristics
func CreateAttackEvent(vehicleID uint32, messageType string, latitude, longitude float64, speed float32, protocol, severity string, attackData map[string]interface{}) *V2XEvent {
	event := CreateV2XMessageEvent(vehicleID, messageType, latitude, longitude, speed, protocol)

	// Override severity for attack
	event.Severity = severity

	// Add attack-specific data to details
	for key, value := range attackData {
		event.Details[key] = value
	}

	// Update message to reflect attack
	if anomalies, ok := attackData["anomalies"]; ok {
		event.Message = fmt.Sprintf("V2X %s message from vehicle VEH%03d [ATTACK DETECTED]", messageType, vehicleID%1000)
		event.Details["anomalies"] = anomalies
	}

	return event
}

// generateRandomIP generates a random IP address for simulation
func generateRandomIP() string {
	return fmt.Sprintf("192.168.%d.%d",
		1+rand.Intn(8),   // 192.168.1-8.x
		1+rand.Intn(254)) // x.1-254
}

// generateRandomPort generates a random port for simulation
func generateRandomPort() int {
	// Common ports: 80, 443, 22, 3306, 5432, 8080, 8443
	commonPorts := []int{80, 443, 22, 3306, 5432, 8080, 8443}
	if rand.Intn(3) == 0 { // 33% chance of using common port
		return commonPorts[rand.Intn(len(commonPorts))]
	}
	// Otherwise random high port
	return 1024 + rand.Intn(64512)
}
