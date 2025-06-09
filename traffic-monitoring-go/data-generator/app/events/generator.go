package events

import (
	"fmt"
	"math/rand"
	"time"
)

type Generator struct {
	includeV2XEvents bool
}

func NewGenerator(includeV2XEvents bool) *Generator {
	return &Generator{
		includeV2XEvents: includeV2XEvents,
	}
}

func (g *Generator) GenerateRandomEvent() Event {
	// Choose a random severity, weighted toward lower severities
	severities := []string{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo}
	weights := []int{1, 3, 6, 10, 15}
	severity := g.weightedRandomChoice(severities, weights)

	// Choose a random category
	categories := []string{
		CategoryAuthentication,
		CategoryAuthorization,
		CategoryNetwork,
		CategoryMalware,
		CategorySystem,
	}

	// Include V2X categories if enabled
	if g.includeV2XEvents {
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
		message, sourceType = g.generateAuthEvent(details, sourceIP)
	case CategoryNetwork:
		message, sourceType = g.generateNetworkEvent(details, sourceIP, sourcePort, destIP, destPort)
	case CategoryMalware:
		message, sourceType = g.generateMalwareEvent(details, sourceIP)
	case CategorySystem:
		message, sourceType = g.generateSystemEvent(details, sourceIP)
	case CategoryVehicle:
		message, sourceType = g.generateVehicleEvent(details)
	case CategoryV2X:
		message, sourceType = g.generateV2XEvent(details)
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

func (g *Generator) generateAuthEvent(details map[string]interface{}, sourceIP string) (string, string) {
	usernames := []string{"admin", "root", "user", "guest", "system", "service"}
	username := usernames[rand.Intn(len(usernames))]
	status := []string{"success", "failure"}[rand.Intn(2)]

	details["username"] = username
	details["status"] = status

	var message string
	if status == "success" {
		message = fmt.Sprintf("User %s successfully authenticated from %s", username, sourceIP)
	} else {
		message = fmt.Sprintf("Failed authentication attempt for user %s from %s", username, sourceIP)
	}

	return message, "authentication"
}

func (g *Generator) generateNetworkEvent(details map[string]interface{}, sourceIP string, sourcePort int, destIP string, destPort int) (string, string) {
	protocols := []string{"TCP", "UDP", "HTTP", "HTTPS", "SSH", "FTP"}
	protocol := protocols[rand.Intn(len(protocols))]
	actions := []string{"allow", "block", "alert", "log"}
	action := actions[rand.Intn(len(actions))]

	details["protocol"] = protocol
	details["action"] = action

	message := fmt.Sprintf("%s connection from %s:%d to %s:%d %s",
		protocol, sourceIP, sourcePort, destIP, destPort, action)

	return message, "firewall"
}

func (g *Generator) generateMalwareEvent(details map[string]interface{}, sourceIP string) (string, string) {
	malwareTypes := []string{"trojan", "virus", "ransomware", "spyware", "worm"}
	malwareType := malwareTypes[rand.Intn(len(malwareTypes))]
	filenames := []string{"/bin/infected", "/tmp/suspicious.exe", "/var/malicious.sh", "/home/user/bad.pdf"}
	filename := filenames[rand.Intn(len(filenames))]

	details["malware_type"] = malwareType
	details["filename"] = filename

	message := fmt.Sprintf("Detected %s in file %s from host %s", malwareType, filename, sourceIP)

	return message, "antivirus"
}

func (g *Generator) generateSystemEvent(details map[string]interface{}, sourceIP string) (string, string) {
	eventTypes := []string{"startup", "shutdown", "error", "warning", "process_crash", "disk_full", "service_start", "service_stop"}
	eventType := eventTypes[rand.Intn(len(eventTypes))]
	services := []string{"httpd", "postgres", "mysql", "nginx", "systemd", "cron", "ssh"}
	service := services[rand.Intn(len(services))]

	details["event_type"] = eventType
	details["service"] = service

	message := fmt.Sprintf("System event: %s - %s on %s", eventType, service, sourceIP)

	return message, "system"
}

func (g *Generator) generateVehicleEvent(details map[string]interface{}) (string, string) {
	vehicleIDs := []string{"VEH001", "VEH002", "VEH003", "VEH004", "VEH005"}
	vehicleID := vehicleIDs[rand.Intn(len(vehicleIDs))]
	componentTypes := []string{"engine", "brakes", "transmission", "fuel", "electrical", "sensors"}
	component := componentTypes[rand.Intn(len(componentTypes))]
	severities := []string{"info", "warning", "error"}
	severity := severities[rand.Intn(len(severities))]

	details["vehicle_id"] = vehicleID
	details["component"] = component
	details["location"] = fmt.Sprintf("%f,%f", 37.7749+rand.Float64()*0.02, -122.4194+rand.Float64()*0.02)

	message := fmt.Sprintf("Vehicle %s reported %s %s event", vehicleID, severity, component)

	return message, "vehicle"
}

func (g *Generator) generateV2XEvent(details map[string]interface{}) (string, string) {
	messageTypes := []string{"basic_safety", "emergency_vehicle", "roadwork_warning", "traffic_signal", "hazard"}
	messageType := messageTypes[rand.Intn(len(messageTypes))]
	vehicleIDs := []string{"VEH001", "VEH002", "VEH003", "VEH004", "VEH005"}
	vehicleID := vehicleIDs[rand.Intn(len(vehicleIDs))]

	details["vehicle_id"] = vehicleID
	details["message_type"] = messageType
	details["protocol"] = "DSRC"
	details["location"] = fmt.Sprintf("%f,%f", 37.7749+rand.Float64()*0.02, -122.4194+rand.Float64()*0.02)
	details["speed"] = 35 + rand.Intn(30)
	details["signature_valid"] = true
	details["trust_level"] = 8

	message := fmt.Sprintf("V2X %s message from vehicle %s", messageType, vehicleID)

	return message, "v2x"
}

func (g *Generator) weightedRandomChoice(choices []string, weights []int) string {
	if len(choices) != len(weights) {
		return choices[rand.Intn(len(choices))]
	}

	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}

	r := rand.Intn(totalWeight)

	for i, w := range weights {
		r -= w
		if r < 0 {
			return choices[i]
		}
	}

	return choices[0]
}
