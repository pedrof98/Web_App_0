package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

// Configuration parameters
var (
	siemAPIURL           string
	eventsPerMinute      int
	enableAttackSim      bool
	attackFrequency      int
	includeV2XEvents     bool
)

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

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Load configuration from environment variables
	loadConfig()

	log.Println("V2X SIEM Data Generator starting...")
	log.Printf("Configured to send events to: %s", siemAPIURL)
	log.Printf("Events per minute: %d", eventsPerMinute)
	log.Printf("Attack simulation enabled: %t", enableAttackSim)
	log.Printf("Attack frequency: %d minutes", attackFrequency)
	log.Printf("V2X events included: %t", includeV2XEvents)

	// Start the data generator
	log.Printf("Starting data generator. Sending to %s", siemAPIURL)
	log.Printf("Generating %d events per minute", eventsPerMinute)

	// Wait for SIEM to be available
	for {
		if isSIEMAvailable() {
			break
		}
		log.Println("Waiting for SIEM to be available... will retry in 5 seconds")
		time.Sleep(5 * time.Second)
	}

	log.Println("SIEM is available! Starting to send events...")

	// Set up ticker for normal events
	interval := time.Minute / time.Duration(eventsPerMinute)
	eventTicker := time.NewTicker(interval)

	// Set up ticker for attack events (if enabled)
	var attackTicker *time.Ticker
	if enableAttackSim {
		attackTicker = time.NewTicker(time.Duration(attackFrequency) * time.Minute)
	}

	// Main loop
	for {
		select {
		case <-eventTicker.C:
			event := generateRandomEvent()
			sendEvent(event)

		case <-attackTicker.C:
			if enableAttackSim {
				log.Println("Generating attack scenario events...")
				generateAttackScenario()
			}
		}
	}
}

// loadConfig loads configuration from environment variables
func loadConfig() {
	// Get SIEM API URL
	siemAPIURL = os.Getenv("SIEM_API_URL")
	if siemAPIURL == "" {
		siemAPIURL = "http://localhost:8080"
	}
	// Remove trailing slash if present
	siemAPIURL = strings.TrimSuffix(siemAPIURL, "/")

	// Get events per minute
	eventsPerMinuteStr := os.Getenv("EVENTS_PER_MINUTE")
	if eventsPerMinuteStr == "" {
		eventsPerMinute = 60 // Default: 1 event per second
	} else {
		fmt.Sscanf(eventsPerMinuteStr, "%d", &eventsPerMinute)
		if eventsPerMinute < 1 {
			eventsPerMinute = 1
		}
	}

	// Get attack simulation setting
	enableAttackSimStr := os.Getenv("ENABLE_ATTACK_SIMULATION")
	enableAttackSim = strings.ToLower(enableAttackSimStr) == "true"

	// Get attack frequency
	attackFrequencyStr := os.Getenv("ATTACK_FREQUENCY")
	if attackFrequencyStr == "" {
		attackFrequency = 30 // Default: 30 minutes
	} else {
		fmt.Sscanf(attackFrequencyStr, "%d", &attackFrequency)
		if attackFrequency < 1 {
			attackFrequency = 1
		}
	}

	// Get V2X events setting
	includeV2XEventsStr := os.Getenv("INCLUDE_V2X_EVENTS")
	includeV2XEvents = strings.ToLower(includeV2XEventsStr) == "true"
}

// isSIEMAvailable checks if the SIEM API is available
func isSIEMAvailable() bool {
	// Use the health endpoint instead of root
	resp, err := http.Get(siemAPIURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// generateRandomEvent creates a random security event
func generateRandomEvent() Event {
	// Choose a random severity, weighted toward lower severities
	severities := []string{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo}
	weights := []int{1, 3, 6, 10, 15}
	severity := weightedRandomChoice(severities, weights)

	// Choose a random category
	categories := []string{
		CategoryAuthentication,
		CategoryAuthorization,
		CategoryNetwork,
		CategoryMalware,
		CategorySystem,
	}
	
	// Include V2X categories if enabled
	if includeV2XEvents {
		categories = append(categories, CategoryVehicle, CategoryV2X)
	}
	
	category := categories[rand.Intn(len(categories))]

	// Generate event details based on category
	sourceIP := fmt.Sprintf("192.168.%d.%d", rand.Intn(10), rand.Intn(254)+1)
	sourcePort := 1024 + rand.Intn(64510)
	destIP := fmt.Sprintf("10.0.%d.%d", rand.Intn(10), rand.Intn(254)+1)
	destPort := []int{22, 80, 443, 3306, 5432, 8080, 8443}[rand.Intn(7)]
	
	details := map[string]interface{}{
		"source_ip":        sourceIP,
		"source_port":      sourcePort,
		"destination_ip":   destIP,
		"destination_port": destPort,
	}
	
	// Add category-specific details
	message := ""
	sourceType := "system"
	
	switch category {
	case CategoryAuthentication:
		usernames := []string{"admin", "root", "user", "guest", "system", "service"}
		username := usernames[rand.Intn(len(usernames))]
		status := []string{"success", "failure"}[rand.Intn(2)]
		sourceType = "authentication"
		
		details["username"] = username
		details["status"] = status
		
		if status == "success" {
			message = fmt.Sprintf("User %s successfully authenticated from %s", username, sourceIP)
		} else {
			message = fmt.Sprintf("Failed authentication attempt for user %s from %s", username, sourceIP)
		}
		
	case CategoryNetwork:
		protocols := []string{"TCP", "UDP", "HTTP", "HTTPS", "SSH", "FTP"}
		protocol := protocols[rand.Intn(len(protocols))]
		actions := []string{"allow", "block", "alert", "log"}
		action := actions[rand.Intn(len(actions))]
		sourceType = "firewall"
		
		details["protocol"] = protocol
		details["action"] = action
		
		message = fmt.Sprintf("%s connection from %s:%d to %s:%d %s", 
			protocol, sourceIP, sourcePort, destIP, destPort, action)
		
	case CategoryMalware:
		malwareTypes := []string{"trojan", "virus", "ransomware", "spyware", "worm"}
		malwareType := malwareTypes[rand.Intn(len(malwareTypes))]
		filenames := []string{"/bin/infected", "/tmp/suspicious.exe", "/var/malicious.sh", "/home/user/bad.pdf"}
		filename := filenames[rand.Intn(len(filenames))]
		sourceType = "antivirus"
		
		details["malware_type"] = malwareType
		details["filename"] = filename
		
		message = fmt.Sprintf("Detected %s in file %s from host %s", malwareType, filename, sourceIP)
		
	case CategorySystem:
		eventTypes := []string{"startup", "shutdown", "error", "warning", "process_crash", "disk_full", "service_start", "service_stop"}
		eventType := eventTypes[rand.Intn(len(eventTypes))]
		services := []string{"httpd", "postgres", "mysql", "nginx", "systemd", "cron", "ssh"}
		service := services[rand.Intn(len(services))]
		sourceType = "system"
		
		details["event_type"] = eventType
		details["service"] = service
		
		message = fmt.Sprintf("System event: %s - %s on %s", eventType, service, sourceIP)
		
	case CategoryVehicle:
		vehicleIDs := []string{"VEH001", "VEH002", "VEH003", "VEH004", "VEH005"}
		vehicleID := vehicleIDs[rand.Intn(len(vehicleIDs))]
		componentTypes := []string{"engine", "brakes", "transmission", "fuel", "electrical", "sensors"}
		component := componentTypes[rand.Intn(len(componentTypes))]
		sourceType = "vehicle"
		
		details["vehicle_id"] = vehicleID
		details["component"] = component
		details["location"] = fmt.Sprintf("%f,%f", 37.7749+rand.Float64()*0.02, -122.4194+rand.Float64()*0.02)
		
		message = fmt.Sprintf("Vehicle %s reported %s %s event", vehicleID, severity, component)
		
	case CategoryV2X:
		messageTypes := []string{"basic_safety", "emergency_vehicle", "roadwork_warning", "traffic_signal", "hazard"}
		messageType := messageTypes[rand.Intn(len(messageTypes))]
		vehicleIDs := []string{"VEH001", "VEH002", "VEH003", "VEH004", "VEH005"}
		vehicleID := vehicleIDs[rand.Intn(len(vehicleIDs))]
		sourceType = "v2x"
		
		details["vehicle_id"] = vehicleID
		details["message_type"] = messageType
		details["location"] = fmt.Sprintf("%f,%f", 37.7749+rand.Float64()*0.02, -122.4194+rand.Float64()*0.02)
		details["speed"] = 35 + rand.Intn(30)
		
		message = fmt.Sprintf("V2X %s message from vehicle %s", messageType, vehicleID)
	}
	
	return Event{
		SourceName: sourceType,
		SourceType: sourceType,
		Timestamp:  time.Now(),
		Severity:   severity,
		Category:   category,
		Message:    message,
		Details:    details,
	}
}

// generateAttackScenario simulates an attack by sending a series of related events
func generateAttackScenario() {
	// Choose attack type
	attackTypes := []string{"brute_force", "port_scan", "malware_spread", "v2x_spoofing"}
	attackType := attackTypes[rand.Intn(len(attackTypes))]
	
	// If V2X events are disabled, don't use v2x_spoofing attack
	if !includeV2XEvents && attackType == "v2x_spoofing" {
		attackType = attackTypes[rand.Intn(len(attackTypes)-1)]
	}
	
	// Common attack details
	attackerIP := fmt.Sprintf("45.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255))
	targetIP := fmt.Sprintf("10.0.%d.%d", rand.Intn(10), rand.Intn(254)+1)
	
	// Number of events in the attack
	eventCount := 5 + rand.Intn(10)
	
	log.Printf("Generating %s attack scenario with %d events", attackType, eventCount)
	
	switch attackType {
	case "brute_force":
		// Simulate brute force authentication attack
		username := []string{"admin", "root", "administrator", "system"}[rand.Intn(4)]
		
		// Several failed logins
		for i := 0; i < eventCount-1; i++ {
			event := Event{
				SourceName: "authentication",
				SourceType: "authentication",
				Timestamp:  time.Now(),
				Severity:   SeverityMedium,
				Category:   CategoryAuthentication,
				Message:    fmt.Sprintf("Failed authentication attempt for user %s from %s", username, attackerIP),
				Details: map[string]interface{}{
					"username":       username,
					"source_ip":      attackerIP,
					"status":         "failure",
					"attempt_number": i + 1,
					"attack":         "brute_force",
				},
			}
			sendEvent(event)
			time.Sleep(time.Millisecond * time.Duration(500+rand.Intn(500)))
		}
		
		// Final successful login
		event := Event{
			SourceName: "authentication",
			SourceType: "authentication",
			Timestamp:  time.Now(),
			Severity:   SeverityCritical,
			Category:   CategoryAuthentication,
			Message:    fmt.Sprintf("Successful authentication for user %s after multiple failures from %s", username, attackerIP),
			Details: map[string]interface{}{
				"username":        username,
				"source_ip":       attackerIP,
				"status":          "success",
				"previous_failed": eventCount - 1,
				"attack":          "brute_force",
			},
		}
		sendEvent(event)
		
	case "port_scan":
		// Simulate port scanning
		ports := []int{21, 22, 23, 25, 53, 80, 443, 445, 3306, 3389, 5432, 8080, 8443}
		
		for i := 0; i < eventCount; i++ {
			port := ports[i%len(ports)]
			event := Event{
				SourceName: "firewall",
				SourceType: "network",
				Timestamp:  time.Now(),
				Severity:   SeverityHigh,
				Category:   CategoryNetwork,
				Message:    fmt.Sprintf("Port scan detected from %s to %s:%d", attackerIP, targetIP, port),
				Details: map[string]interface{}{
					"source_ip":        attackerIP,
					"source_port":      rand.Intn(65535),
					"destination_ip":   targetIP,
					"destination_port": port,
					"protocol":         "TCP",
					"action":           "block",
					"attack":           "port_scan",
				},
			}
			sendEvent(event)
			time.Sleep(time.Millisecond * time.Duration(100+rand.Intn(200)))
		}
		
	case "malware_spread":
		// Simulate malware spreading across systems
		malwareType := []string{"trojan", "ransomware", "worm"}[rand.Intn(3)]
		malwareName := fmt.Sprintf("MALWARE_%X", rand.Intn(0x1000000))
		hosts := []string{}
		
		// Generate some random host IPs in the same subnet
		for i := 0; i < eventCount; i++ {
			hosts = append(hosts, fmt.Sprintf("10.0.5.%d", 10+i))
		}
		
		// Initial infection
		event := Event{
			SourceName: "antivirus",
			SourceType: "malware",
			Timestamp:  time.Now(),
			Severity:   SeverityCritical,
			Category:   CategoryMalware,
			Message:    fmt.Sprintf("Initial %s infection detected on %s", malwareType, hosts[0]),
			Details: map[string]interface{}{
				"malware_type": malwareType,
				"malware_name": malwareName,
				"source_ip":    attackerIP,
				"host":         hosts[0],
				"filename":     "/tmp/infected.bin",
				"attack":       "malware_spread",
				"stage":        "initial_infection",
			},
		}
		sendEvent(event)
		time.Sleep(time.Second * time.Duration(1+rand.Intn(2)))
		
		// Spreading across systems
		for i := 1; i < len(hosts); i++ {
			event := Event{
				SourceName: "antivirus",
				SourceType: "malware",
				Timestamp:  time.Now(),
				Severity:   SeverityHigh,
				Category:   CategoryMalware,
				Message:    fmt.Sprintf("%s spreading to %s from %s", malwareName, hosts[i], hosts[i-1]),
				Details: map[string]interface{}{
					"malware_type":     malwareType,
					"malware_name":     malwareName,
					"source_ip":        hosts[i-1],
					"destination_ip":   hosts[i],
					"filename":         "/tmp/infected.bin",
					"attack":           "malware_spread",
					"stage":            "propagation",
					"propagation_path": i,
				},
			}
			sendEvent(event)
			time.Sleep(time.Second * time.Duration(1+rand.Intn(3)))
		}
		
	case "v2x_spoofing":
		// Simulate V2X message spoofing
		vehicleIDs := []string{"VEH001", "VEH002", "VEH003", "VEH004", "VEH005"}
		attackerVehicle := fmt.Sprintf("UNKNOWN_%X", rand.Intn(0x1000000))
		messageTypes := []string{"emergency_vehicle", "traffic_signal", "hazard_warning"}
		messageType := messageTypes[rand.Intn(len(messageTypes))]
		
		// Initial spoofed message
		event := Event{
			SourceName: "v2x",
			SourceType: "v2x",
			Timestamp:  time.Now(),
			Severity:   SeverityCritical,
			Category:   CategoryV2X,
			Message:    fmt.Sprintf("Potentially spoofed V2X %s message detected from unregistered vehicle", messageType),
			Details: map[string]interface{}{
				"vehicle_id":   attackerVehicle,
				"message_type": messageType,
				"location":     fmt.Sprintf("%f,%f", 37.7749+rand.Float64()*0.02, -122.4194+rand.Float64()*0.02),
				"attack":       "v2x_spoofing",
				"stage":        "initial_detection",
			},
		}
		sendEvent(event)
		time.Sleep(time.Second * time.Duration(1+rand.Intn(2)))
		
		// Vehicle responses to spoofed message
		for i := 0; i < eventCount-1; i++ {
			victimVehicle := vehicleIDs[i%len(vehicleIDs)]
			event := Event{
				SourceName: "v2x",
				SourceType: "v2x",
				Timestamp:  time.Now(),
				Severity:   SeverityHigh,
				Category:   CategoryV2X,
				Message:    fmt.Sprintf("Vehicle %s responding to potentially spoofed message from %s", victimVehicle, attackerVehicle),
				Details: map[string]interface{}{
					"vehicle_id":        victimVehicle,
					"message_type":      "response",
					"malicious_source":  attackerVehicle,
					"location":          fmt.Sprintf("%f,%f", 37.7749+rand.Float64()*0.02, -122.4194+rand.Float64()*0.02),
					"speed_change":      -10 - rand.Intn(20),
					"attack":            "v2x_spoofing",
					"stage":             "vehicle_response",
					"response_sequence": i + 1,
				},
			}
			sendEvent(event)
			time.Sleep(time.Second * time.Duration(rand.Intn(2)))
		}
	}
}

// sendEvent sends an event to the SIEM API
func sendEvent(event Event) {
	jsonData, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}
	
	resp, err := http.Post(siemAPIURL+"/ingest", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		log.Printf("Error sending event: %v", err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("Error response from SIEM: %d", resp.StatusCode)
		return
	}
	
	// Successful send
	if rand.Intn(100) < 5 { // Only log ~5% of events to avoid flooding logs
		log.Printf("Sent %s %s event: %s", event.Severity, event.Category, event.Message)
	}
}

// weightedRandomChoice selects a random item from choices based on weights
func weightedRandomChoice(choices []string, weights []int) string {
	if len(choices) != len(weights) {
		return choices[rand.Intn(len(choices))]
	}
	
	// Calculate total weight
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}
	
	// Generate a random value between 0 and totalWeight
	r := rand.Intn(totalWeight)
	
	// Find the item that corresponds to this value
	for i, w := range weights {
		r -= w
		if r < 0 {
			return choices[i]
		}
	}
	
	// Fallback (should never reach here if weights are positive)
	return choices[0]
}