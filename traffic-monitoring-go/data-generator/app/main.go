package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
)

// Configuration for the data generator
type Config struct {
	SIEMURL               string
	EventsPerMinute       int
	EnableAttackSimulation bool
	AttackFrequency       int  // in minutes
	IncludeV2XEvents      bool
}

// EventGenerator is responsible for generating security events
type EventGenerator struct {
	Config         Config
	HTTPClient     *http.Client
	DeviceIDs      []string
	SourceIPs      []string
	DestinationIPs []string
}

// RawEvent represents a security event to be sent to the SIEM
type RawEvent struct {
	SourceName string                 `json:"source_name"`
	SourceType string                 `json:"source_type"`
	Timestamp  time.Time              `json:"timestamp"`
	Severity   string                 `json:"severity"`
	Category   string                 `json:"category"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details"`
}

// Initialize the generator with random data
func NewEventGenerator(config Config) *EventGenerator {
	// Seed the random generator
	rand.Seed(time.Now().UnixNano())
	
	// Create some random device IDs
	deviceIDs := make([]string, 25)
	for i := range deviceIDs {
		deviceIDs[i] = fmt.Sprintf("device-%s", uuid.New().String()[:8])
	}

	// Create some source IPs (internal network)
	sourceIPs := make([]string, 15)
	for i := range sourceIPs {
		sourceIPs[i] = fmt.Sprintf("10.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255))
	}

	// Create some destination IPs (external)
	destIPs := make([]string, 20)
	for i := range destIPs {
		destIPs[i] = fmt.Sprintf("%d.%d.%d.%d", rand.Intn(223)+1, rand.Intn(255), rand.Intn(255), rand.Intn(255))
	}

	return &EventGenerator{
		Config:         config,
		HTTPClient:     &http.Client{Timeout: 10 * time.Second},
		DeviceIDs:      deviceIDs,
		SourceIPs:      sourceIPs,
		DestinationIPs: destIPs,
	}
}

// Start generates and sends events at the configured rate
func (g *EventGenerator) Start() {
	log.Printf("Starting data generator. Sending to %s", g.Config.SIEMURL)
	log.Printf("Generating %d events per minute", g.Config.EventsPerMinute)

	// Calculate interval between events
	interval := time.Minute / time.Duration(g.Config.EventsPerMinute)
	
	// Start a timer for attack simulation if enabled
	if g.Config.EnableAttackSimulation {
		log.Printf("Attack simulation enabled, frequency: every %d minutes", g.Config.AttackFrequency)
		go g.scheduleAttacks()
	}

	// Main event generation loop
	ticker := time.NewTicker(interval)
	for range ticker.C {
		event := g.generateRandomEvent()
		if err := g.sendEvent(event); err != nil {
			log.Printf("Error sending event: %v", err)
		}
	}
}

// sendEvent sends a security event to the SIEM system
func (g *EventGenerator) sendEvent(event RawEvent) error {
	// Convert event to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshaling event: %v", err)
	}
	
	// Send to the SIEM API endpoint
	url := fmt.Sprintf("%s/ingest", g.Config.SIEMURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// Make the request
	resp, err := g.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received non-success status code: %d", resp.StatusCode)
	}
	
	return nil
}

// scheduleAttacks periodically triggers attack simulations
func (g *EventGenerator) scheduleAttacks() {
	ticker := time.NewTicker(time.Duration(g.Config.AttackFrequency) * time.Minute)
	for range ticker.C {
		attackType := g.getRandomAttackType()
		log.Printf("Simulating attack: %s", attackType)
		
		// Simulate attack with multiple events
		go g.simulateAttack(attackType)
	}
}

// getRandomAttackType returns a random attack type to simulate
func (g *EventGenerator) getRandomAttackType() string {
	attackTypes := []string{
		"port_scan",
		"brute_force",
		"malware_activity",
		"data_exfiltration",
		"denial_of_service",
	}
	
	if g.Config.IncludeV2XEvents {
		v2xAttacks := []string{
			"gps_spoofing",
			"message_injection",
			"v2x_signal_jamming",
			"unauthorized_access",
		}
		attackTypes = append(attackTypes, v2xAttacks...)
	}
	
	return attackTypes[rand.Intn(len(attackTypes))]
}

// simulateAttack generates a series of related events to simulate an attack
func (g *EventGenerator) simulateAttack(attackType string) {
	// Number of events in the attack
	eventCount := rand.Intn(20) + 5 // 5 to 24 events
	
	// Common details for all events in this attack
	sourceIP := g.SourceIPs[rand.Intn(len(g.SourceIPs))]
	destIP := g.DestinationIPs[rand.Intn(len(g.DestinationIPs))]
	attackID := uuid.New().String()
	
	for i := 0; i < eventCount; i++ {
		var event RawEvent
		
		switch attackType {
		case "port_scan":
			event = g.generatePortScanEvent(sourceIP, destIP, attackID, i, eventCount)
		case "brute_force":
			event = g.generateBruteForceEvent(sourceIP, destIP, attackID, i, eventCount)
		case "malware_activity":
			event = g.generateMalwareEvent(sourceIP, destIP, attackID, i, eventCount)
		case "data_exfiltration":
			event = g.generateDataExfiltrationEvent(sourceIP, destIP, attackID, i, eventCount)
		case "denial_of_service":
			event = g.generateDoSEvent(sourceIP, destIP, attackID, i, eventCount)
		case "gps_spoofing":
			event = g.generateGPSSpoofingEvent(sourceIP, destIP, attackID, i, eventCount)
		case "message_injection":
			event = g.generateV2XMessageInjectionEvent(sourceIP, destIP, attackID, i, eventCount)
		case "v2x_signal_jamming":
			event = g.generateV2XJammingEvent(sourceIP, destIP, attackID, i, eventCount)
		case "unauthorized_access":
			event = g.generateV2XUnauthorizedAccessEvent(sourceIP, destIP, attackID, i, eventCount)
		default:
			event = g.generateRandomEvent()
		}
		
		// Add attack correlation ID
		event.Details["attack_id"] = attackID
		event.Details["attack_type"] = attackType
		
		// Send the event
		if err := g.sendEvent(event); err != nil {
			log.Printf("Error sending attack event: %v", err)
		}
		
		// Delay between events in the attack
		time.Sleep(time.Duration(rand.Intn(2000)+500) * time.Millisecond)
	}
}

// generatePortScanEvent creates events for a port scanning attack
func (g *EventGenerator) generatePortScanEvent(sourceIP, destIP, attackID string, eventIndex, totalEvents int) RawEvent {
	ports := []int{21, 22, 23, 25, 53, 80, 110, 111, 135, 139, 143, 443, 445, 993, 995, 1723, 3306, 3389, 5900, 8080}
	port := ports[eventIndex%len(ports)]
	
	severity := "medium"
	if eventIndex > totalEvents*2/3 {
		severity = "high" // Escalate severity as the attack progresses
	}
	
	return RawEvent{
		SourceName: "firewall",
		SourceType: "network",
		Timestamp:  time.Now().Add(-time.Duration(rand.Intn(60)) * time.Second),
		Severity:   severity,
		Category:   "network",
		Message:    fmt.Sprintf("Port scan detected from %s to %s:%d", sourceIP, destIP, port),
		Details: map[string]interface{}{
			"source_ip":       sourceIP,
			"destination_ip":  destIP,
			"destination_port": port,
			"protocol":        "TCP",
			"action":          "block",
			"rule_id":         rand.Intn(100) + 1000,
		},
	}
}

// generateBruteForceEvent creates events for a brute force login attack
func (g *EventGenerator) generateBruteForceEvent(sourceIP, destIP, attackID string, eventIndex, totalEvents int) RawEvent {
	severity := "medium"
	status := "failure"
	
	// Last event is a successful login
	if eventIndex == totalEvents-1 {
		severity = "critical"
		status = "success"
	} else if eventIndex > totalEvents*3/4 {
		severity = "high"
	}
	
	usernames := []string{"admin", "root", "administrator", "user", "guest"}
	username := usernames[rand.Intn(len(usernames))]
	
	return RawEvent{
		SourceName: "auth_service",
		SourceType: "application",
		Timestamp:  time.Now().Add(-time.Duration(rand.Intn(60)) * time.Second),
		Severity:   severity,
		Category:   "authentication",
		Message:    fmt.Sprintf("Authentication attempt for user '%s' from %s: %s", username, sourceIP, status),
		Details: map[string]interface{}{
			"source_ip":      sourceIP,
			"destination_ip": destIP,
			"username":       username,
			"status":         status,
			"service":        "ssh",
			"destination_port": 22,
		},
	}
}

// generateMalwareEvent creates events for malware activity
func (g *EventGenerator) generateMalwareEvent(sourceIP, destIP, attackID string, eventIndex, totalEvents int) RawEvent {
	malwareTypes := []string{"trojan", "ransomware", "spyware", "worm", "rootkit"}
	malwareName := fmt.Sprintf("%s.%s", gofakeit.AnimalType(), malwareTypes[rand.Intn(len(malwareTypes))])
	
	severity := "high"
	if eventIndex > totalEvents/2 {
		severity = "critical"
	}
	
	actions := []string{"detected", "blocked", "quarantined", "file deleted"}
	action := actions[rand.Intn(len(actions))]
	
	return RawEvent{
		SourceName: "antivirus",
		SourceType: "system",
		Timestamp:  time.Now().Add(-time.Duration(rand.Intn(60)) * time.Second),
		Severity:   severity,
		Category:   "malware",
		Message:    fmt.Sprintf("Malware '%s' %s on device", malwareName, action),
		Details: map[string]interface{}{
			"source_ip":    sourceIP,
			"device_id":    g.DeviceIDs[rand.Intn(len(g.DeviceIDs))],
			"malware_name": malwareName,
			"malware_type": malwareTypes[rand.Intn(len(malwareTypes))],
			"file_path":    fmt.Sprintf("/tmp/%s.exe", gofakeit.UUID()),
			"action":       action,
		},
	}
}

// generateDataExfiltrationEvent creates events for data exfiltration
func (g *EventGenerator) generateDataExfiltrationEvent(sourceIP, destIP, attackID string, eventIndex, totalEvents int) RawEvent {
	dataTypes := []string{"PII", "credentials", "financial", "intellectual property", "customer data"}
	dataType := dataTypes[rand.Intn(len(dataTypes))]
	
	severity := "high"
	sizeKB := rand.Intn(1000) + 50
	
	if sizeKB > 500 {
		severity = "critical"
	}
	
	return RawEvent{
		SourceName: "dlp",
		SourceType: "network",
		Timestamp:  time.Now().Add(-time.Duration(rand.Intn(60)) * time.Second),
		Severity:   severity,
		Category:   "data",
		Message:    fmt.Sprintf("Potential data exfiltration: %s data transferred to external IP", dataType),
		Details: map[string]interface{}{
			"source_ip":      sourceIP,
			"destination_ip": destIP,
			"protocol":       "HTTPS",
			"destination_port": 443,
			"data_type":      dataType,
			"size_kb":        sizeKB,
			"user_id":        rand.Intn(100) + 1,
		},
	}
}

// generateDoSEvent creates events for a denial of service attack
func (g *EventGenerator) generateDoSEvent(sourceIP, destIP, attackID string, eventIndex, totalEvents int) RawEvent {
	severity := "medium"
	packetCount := rand.Intn(1000) + 100
	
	if eventIndex > totalEvents/2 {
		severity = "high"
		packetCount = rand.Intn(10000) + 1000
	}
	
	if eventIndex > totalEvents*3/4 {
		severity = "critical"
		packetCount = rand.Intn(50000) + 10000
	}
	
	return RawEvent{
		SourceName: "ids",
		SourceType: "network",
		Timestamp:  time.Now().Add(-time.Duration(rand.Intn(60)) * time.Second),
		Severity:   severity,
		Category:   "network",
		Message:    fmt.Sprintf("Potential DoS attack detected from %s", sourceIP),
		Details: map[string]interface{}{
			"source_ip":      sourceIP,
			"destination_ip": destIP,
			"packet_count":   packetCount,
			"protocol":       "UDP",
			"attack_vector":  "flood",
		},
	}
}

// generateGPSSpoofingEvent creates events for GPS spoofing in a V2X context
func (g *EventGenerator) generateGPSSpoofingEvent(sourceIP, destIP, attackID string, eventIndex, totalEvents int) RawEvent {
	severity := "medium"
	if eventIndex > totalEvents/2 {
		severity = "high"
	}
	
	vehicle := g.DeviceIDs[rand.Intn(len(g.DeviceIDs))]
	location := fmt.Sprintf("%f,%f", gofakeit.Latitude(), gofakeit.Longitude())
	
	return RawEvent{
		SourceName: "v2x_security",
		SourceType: "vehicle",
		Timestamp:  time.Now().Add(-time.Duration(rand.Intn(60)) * time.Second),
		Severity:   severity,
		Category:   "v2x",
		Message:    fmt.Sprintf("Potential GPS spoofing detected for vehicle %s", vehicle),
		Details: map[string]interface{}{
			"device_id":     vehicle,
			"location":      location,
			"speed":         rand.Float64() * 120,
			"reported_lat":  gofakeit.Latitude(),
			"reported_lon":  gofakeit.Longitude(),
			"expected_lat":  gofakeit.Latitude(),
			"expected_lon":  gofakeit.Longitude(),
			"discrepancy_m": rand.Float64() * 1000,
		},
	}
}

// generateV2XMessageInjectionEvent creates events for message injection in V2X network
func (g *EventGenerator) generateV2XMessageInjectionEvent(sourceIP, destIP, attackID string, eventIndex, totalEvents int) RawEvent {
	severity := "high"
	messageTypes := []string{"BSM", "SPaT", "MAP", "RSA", "PSM", "SSM"}
	messageType := messageTypes[rand.Intn(len(messageTypes))]
	
	if messageType == "BSM" || messageType == "RSA" {
		severity = "critical"
	}
	
	return RawEvent{
		SourceName: "v2x_idsus",
		SourceType: "v2x",
		Timestamp:  time.Now().Add(-time.Duration(rand.Intn(60)) * time.Second),
		Severity:   severity,
		Category:   "v2x",
		Message:    fmt.Sprintf("Unauthorized %s message detected in V2X network", messageType),
		Details: map[string]interface{}{
			"source_id":        sourceIP,
			"message_type":     messageType,
			"signature_valid":  false,
			"certificate_id":   gofakeit.UUID(),
			"message_content":  "INVALID",
			"detection_point":  fmt.Sprintf("RSU-%d", rand.Intn(100)),
		},
	}
}

// generateV2XJammingEvent creates events for signal jamming in V2X network
func (g *EventGenerator) generateV2XJammingEvent(sourceIP, destIP, attackID string, eventIndex, totalEvents int) RawEvent {
	severity := "medium"
	if eventIndex > totalEvents/2 {
		severity = "high"
	}
	
	jammingStrength := rand.Float64() * 10
	if jammingStrength > 7 {
		severity = "critical"
	}
	
	return RawEvent{
		SourceName: "v2x_monitor",
		SourceType: "v2x",
		Timestamp:  time.Now().Add(-time.Duration(rand.Intn(60)) * time.Second),
		Severity:   severity,
		Category:   "v2x",
		Message:    "Signal jamming detected in V2X communication channel",
		Details: map[string]interface{}{
			"location":          fmt.Sprintf("%f,%f", gofakeit.Latitude(), gofakeit.Longitude()),
			"frequency_band":    "5.9GHz",
			"channel":           rand.Intn(10) + 170,
			"signal_strength":   -1 * (rand.Float64() * 50 + 30), // in dBm
			"jamming_strength":  jammingStrength,
			"affected_vehicles": rand.Intn(50) + 1,
			"affected_radius_m": rand.Float64() * 300 + 50,
		},
	}
}

// generateV2XUnauthorizedAccessEvent creates events for unauthorized access to V2X infrastructure
func (g *EventGenerator) generateV2XUnauthorizedAccessEvent(sourceIP, destIP, attackID string, eventIndex, totalEvents int) RawEvent {
	severity := "high"
	
	components := []string{"RSU", "OBU", "certificate authority", "traffic management system"}
	component := components[rand.Intn(len(components))]
	
	if component == "certificate authority" || component == "traffic management system" {
		severity = "critical"
	}
	
	return RawEvent{
		SourceName: "v2x_auth",
		SourceType: "v2x",
		Timestamp:  time.Now().Add(-time.Duration(rand.Intn(60)) * time.Second),
		Severity:   severity,
		Category:   "authentication",
		Message:    fmt.Sprintf("Unauthorized access attempt to %s", component),
		Details: map[string]interface{}{
			"source_ip":      sourceIP,
			"destination_ip": destIP,
			"component":      component,
			"component_id":   fmt.Sprintf("%s-%d", strings.ToLower(strings.Split(component, " ")[0]), rand.Intn(100)),
			"attempt_count":  eventIndex + 1,
			"user_agent":     gofakeit.UserAgent(),
			"method":         "POST",
			"endpoint":       "/admin/control",
		},
	}
}

// generateRandomEvent creates a random security event
func (g *EventGenerator) generateRandomEvent() RawEvent {
	// Create some variations of event types
	sourceTypes := []string{"system", "network", "application", "vehicle", "v2x", "sensor", "station"}
	categories := []string{"authentication", "authorization", "network", "malware", "system", "vehicle", "v2x"}
	severities := []string{"info", "low", "medium", "high", "critical"}
	
	// Weight probabilities toward less severe events
	severityIndex := 0
	r := rand.Float64()
	if r < 0.4 {
		severityIndex = 0 // 40% info
	} else if r < 0.7 {
		severityIndex = 1 // 30% low
	} else if r < 0.85 {
		severityIndex = 2 // 15% medium
	} else if r < 0.95 {
		severityIndex = 3 // 10% high
	} else {
		severityIndex = 4 // 5% critical
	}
	
	// Filter V2X events if not enabled
	sourceTypeIndex := rand.Intn(len(sourceTypes))
	categoryIndex := rand.Intn(len(categories))
	
	if !g.Config.IncludeV2XEvents {
		// Retry if we get V2X-related types
		for sourceTypes[sourceTypeIndex] == "v2x" || sourceTypes[sourceTypeIndex] == "vehicle" {
			sourceTypeIndex = rand.Intn(len(sourceTypes))
		}
		for categories[categoryIndex] == "v2x" || categories[categoryIndex] == "vehicle" {
			categoryIndex = rand.Intn(len(categories))
		}
	}
	
	sourceType := sourceTypes[sourceTypeIndex]
	category := categories[categoryIndex]
	severity := severities[severityIndex]
	
	// Generate different event source names based on source type
	var sourceName string
	switch sourceType {
	case "system":
		options := []string{"syslog", "kernel", "systemd", "cron", "audit"}
		sourceName = options[rand.Intn(len(options))]
	case "network":
		options := []string{"firewall", "router", "switch", "proxy", "ids", "waf"}
		sourceName = options[rand.Intn(len(options))]
	case "application":
		options := []string{"webapp", "database", "api", "auth_service", "logging"}
		sourceName = options[rand.Intn(len(options))]
	case "vehicle":
		options := []string{"onboard_unit", "sensors", "can_bus", "vehicle_control"}
		sourceName = options[rand.Intn(len(options))]
	case "v2x":
		options := []string{"v2x_gateway", "v2x_security", "v2x_messaging"}
		sourceName = options[rand.Intn(len(options))]
	default:
		sourceName = sourceType
	}
	
	// Generate message based on category and severity
	var message string
	details := make(map[string]interface{})
	
	// Add common fields
	deviceID := g.DeviceIDs[rand.Intn(len(g.DeviceIDs))]
	sourceIP := g.SourceIPs[rand.Intn(len(g.SourceIPs))]
	
	details["device_id"] = deviceID
	details["source_ip"] = sourceIP
	
	// Add specific message and details based on category
	switch category {
	case "authentication":
		usernames := []string{"admin", "user", "system", "service", "guest"}
		username := usernames[rand.Intn(len(usernames))]
		status := "success"
		if severity == "medium" || severity == "high" || severity == "critical" {
			status = "failure"
		}
		message = fmt.Sprintf("Authentication %s for user %s", status, username)
		details["username"] = username
		details["status"] = status
		details["auth_method"] = "password"
		
	case "authorization":
		resources := []string{"file", "database", "api", "admin console", "configuration"}
		resource := resources[rand.Intn(len(resources))]
		message = fmt.Sprintf("Access denied to %s", resource)
		details["resource"] = resource
		details["required_permission"] = "admin"
		
	case "network":
		protocols := []string{"TCP", "UDP", "HTTP", "HTTPS", "DNS"}
		protocol := protocols[rand.Intn(len(protocols))]
		destIP := g.DestinationIPs[rand.Intn(len(g.DestinationIPs))]
		port := rand.Intn(65535) + 1
		
		details["destination_ip"] = destIP
		details["destination_port"] = port
		details["protocol"] = protocol
		
		if severity == "info" || severity == "low" {
			message = fmt.Sprintf("Connection from %s to %s:%d", sourceIP, destIP, port)
			details["status"] = "allowed"
		} else {
			message = fmt.Sprintf("Blocked connection from %s to %s:%d", sourceIP, destIP, port)
			details["status"] = "blocked"
			details["rule_id"] = rand.Intn(1000) + 1
		}
		
	case "malware":
		malwareTypes := []string{"virus", "trojan", "ransomware", "spyware", "adware"}
		malwareType := malwareTypes[rand.Intn(len(malwareTypes))]
		message = fmt.Sprintf("Potential %s detected", malwareType)
		details["malware_type"] = malwareType
		details["file_path"] = fmt.Sprintf("/tmp/%s", gofakeit.LoremIpsumWord())
		
	case "system":
		events := []string{"startup", "shutdown", "crash", "update", "resource usage"}
		event := events[rand.Intn(len(events))]
		message = fmt.Sprintf("System %s event", event)
		details["event_type"] = event
		details["system_component"] = "kernel"
		
	case "vehicle":
		components := []string{"engine", "brakes", "transmission", "sensors", "navigation"}
		component := components[rand.Intn(len(components))]
		actions := []string{"warning", "error", "status change", "calibration"}
		action := actions[rand.Intn(len(actions))]
		message = fmt.Sprintf("Vehicle %s: %s", component, action)
		details["component"] = component
		details["action"] = action
		details["vehicle_id"] = deviceID
		details["speed"] = rand.Float64() * 120
		details["location"] = fmt.Sprintf("%f,%f", gofakeit.Latitude(), gofakeit.Longitude())
		
	case "v2x":
		messageTypes := []string{"BSM", "SPaT", "MAP", "RSA", "PSM"}
		msgType := messageTypes[rand.Intn(len(messageTypes))]
		events := []string{"received", "sent", "processed", "validated", "error"}
		event := events[rand.Intn(len(events))]
		message = fmt.Sprintf("V2X %s message %s", msgType, event)
		details["message_type"] = msgType
		details["event"] = event
		details["vehicle_id"] = deviceID
		details["location"] = fmt.Sprintf("%f,%f", gofakeit.Latitude(), gofakeit.Longitude())
	}
	
	// Generate a timestamp with slight randomness (mostly recent)
	timestamp := time.Now().Add(-time.Duration(rand.Intn(300)) * time.Second)
	
	return RawEvent{
		SourceName: sourceName,
		SourceType: sourceType,
		Timestamp:  timestamp,
		Severity:   severity,
		Category:   category,
		Message:    message,
		Details:    details,
	}
}

// getConfigFromEnv loads configuration from environment variables
func getConfigFromEnv() Config {
	config := Config{
		SIEMURL:               "http://app:8080",
		EventsPerMinute:       100,
		EnableAttackSimulation: true,
		AttackFrequency:       30,
		IncludeV2XEvents:      true,
	}
	
	// Override with environment variables if present
	if url := os.Getenv("SIEM_API_URL"); url != "" {
		config.SIEMURL = url
	}
	
	if epm := os.Getenv("EVENTS_PER_MINUTE"); epm != "" {
		if val, err := strconv.Atoi(epm); err == nil && val > 0 {
			config.EventsPerMinute = val
		}
	}
	
	if sim := os.Getenv("ENABLE_ATTACK_SIMULATION"); sim != "" {
		config.EnableAttackSimulation = strings.ToLower(sim) == "true"
	}
	
	if freq := os.Getenv("ATTACK_FREQUENCY"); freq != "" {
		if val, err := strconv.Atoi(freq); err == nil && val > 0 {
			config.AttackFrequency = val
		}
	}
	
	if v2x := os.Getenv("INCLUDE_V2X_EVENTS"); v2x != "" {
		config.IncludeV2XEvents = strings.ToLower(v2x) == "true"
	}
	
	return config
}

func main() {
	// Initialize configuration from environment variables
	config := getConfigFromEnv()
	
	// Create and start the event generator
	generator := NewEventGenerator(config)
	
	// Log startup information
	log.Printf("V2X SIEM Data Generator starting...")
	log.Printf("Configured to send events to: %s", config.SIEMURL)
	log.Printf("Events per minute: %d", config.EventsPerMinute)
	log.Printf("Attack simulation enabled: %v", config.EnableAttackSimulation)
	if config.EnableAttackSimulation {
		log.Printf("Attack frequency: %d minutes", config.AttackFrequency)
	}
	log.Printf("V2X events included: %v", config.IncludeV2XEvents)
	
	// Start generating events
	generator.Start()
}