package events

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

func (g *Generator) GenerateAttackScenario(sendFunc func(Event)) {
	attackTypes := []string{
		"brute_force",
		"port_scan",
		"malware_spread",
		"v2x_position_jump",
		"v2x_message_flood",
		"v2x_invalid_signature",
		"v2x_speed_anomaly",
	}

	attackType := attackTypes[rand.Intn(len(attackTypes))]

	// If V2X events are disabled, don't use v2x attacks
	if !g.includeV2XEvents && (attackType == "v2x_position_jump" || attackType == "v2x_message_flood" || attackType == "v2x_invalid_signature" || attackType == "v2x_speed_anomaly") {
		attackType = attackTypes[rand.Intn(3)] // Use only first 3 (non-V2X attacks)
	}

	eventCount := 5 + rand.Intn(10)
	log.Printf("Generating %s attack scenario with %d events", attackType, eventCount)

	switch attackType {
	case "brute_force":
		g.generateBruteForceAttack(eventCount, sendFunc)
	case "port_scan":
		g.generatePortScanAttack(eventCount, sendFunc)
	case "malware_spread":
		g.generateMalwareSpreadAttack(eventCount, sendFunc)
	case "v2x_position_jump":
		GenerateV2XPositionJumpAttack(eventCount, sendFunc)
	case "v2x_message_flood":
		GenerateV2XMessageFloodAttack(eventCount, sendFunc)
	case "v2x_invalid_signature":
		GenerateV2XInvalidSignatureAttack(eventCount, sendFunc)
	case "v2x_speed_anomaly":
		GenerateV2XSpeedAnomalyAttack(eventCount, sendFunc)
	}
}

func (g *Generator) generateBruteForceAttack(eventCount int, sendFunc func(Event)) {
	attackerIP := fmt.Sprintf("45.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255))
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
		sendFunc(event)
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
	sendFunc(event)
}

func (g *Generator) generatePortScanAttack(eventCount int, sendFunc func(Event)) {
	attackerIP := fmt.Sprintf("45.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255))
	targetIP := fmt.Sprintf("10.0.%d.%d", rand.Intn(10), rand.Intn(254)+1)
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
		sendFunc(event)
		time.Sleep(time.Millisecond * time.Duration(100+rand.Intn(200)))
	}
}

func (g *Generator) generateMalwareSpreadAttack(eventCount int, sendFunc func(Event)) {
	attackerIP := fmt.Sprintf("45.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255))
	malwareType := []string{"trojan", "ransomware", "worm"}[rand.Intn(3)]
	malwareName := fmt.Sprintf("MALWARE_%X", rand.Intn(0x1000000))
	hosts := []string{}

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
	sendFunc(event)
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
		sendFunc(event)
		time.Sleep(time.Second * time.Duration(1+rand.Intn(3)))
	}
}
