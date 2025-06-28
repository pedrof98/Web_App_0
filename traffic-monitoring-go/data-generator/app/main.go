package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"traffic-monitoring-go/data-generator/app/api"
	"traffic-monitoring-go/data-generator/app/config"
	"traffic-monitoring-go/data-generator/app/events"
	"traffic-monitoring-go/data-generator/app/sender"
)

var (
	apiPort        = 8082 // Default API port
	eventSender    *sender.Sender
	eventGenerator *events.Generator
)

func main() {
	rand.Seed(time.Now().UnixNano())

	cfg := config.Load()
	eventSender = sender.New(cfg.SIEMAPIURL)
	eventGenerator = events.NewGenerator(cfg.IncludeV2XEvents)

	// Get API port from environment
	if envApiPort := os.Getenv("API_PORT"); envApiPort != "" {
		if p, err := strconv.Atoi(envApiPort); err == nil {
			apiPort = p
		}
	}

	log.Println("V2X SIEM Data Generator starting...")
	log.Printf("Configured to send events to: %s", cfg.SIEMAPIURL)
	log.Printf("Events per minute: %d", cfg.EventsPerMinute)
	log.Printf("Attack simulation enabled: %t", cfg.EnableAttackSim)
	log.Printf("V2X events included: %t", cfg.IncludeV2XEvents)
	log.Printf("Attack trigger API on port: %d", apiPort)

	// Wait for SIEM to be available
	for {
		if eventSender.IsSIEMAvailable() {
			break
		}
		log.Println("Waiting for SIEM to be available... will retry in 5 seconds")
		time.Sleep(5 * time.Second)
	}

	log.Println("SIEM is available! Starting to send events...")

	// Start API server and attack processor
	go startAPIServer()
	go processAttackTriggers()

	// Set up ticker for normal events
	interval := time.Minute / time.Duration(cfg.EventsPerMinute)
	eventTicker := time.NewTicker(interval)

	// Set up ticker for attack events (if enabled)
	var attackTicker *time.Ticker

	if cfg.EnableAttackSim {
		attackTicker = time.NewTicker(time.Duration(cfg.AttackFrequency) * time.Minute)
	}

	// Main loop
	for {
		select {
		case <-eventTicker.C:
			event := eventGenerator.GenerateRandomEvent()
			eventSender.SendEvent(event)
			// Record normal event in dataset
			api.DatasetExporter.AddEvent(event, false, "")

		case <-attackTicker.C:
			if cfg.EnableAttackSim {
				log.Println("Generating attack scenario events...")
				eventGenerator.GenerateAttackScenario(func(event events.Event) {
					eventSender.SendEvent(event)
					// Record attack event in dataset
					api.DatasetExporter.AddEvent(event, true, extractAttackType(event))
				})
			}
		}
	}
}

// startAPIServer starts the HTTP API for attack triggers
func startAPIServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/attack/trigger", corsHandler(api.HandleAttackTrigger))
	mux.HandleFunc("/attack/status", corsHandler(api.HandleAttackStatus))
	mux.HandleFunc("/health", corsHandler(healthHandler))
	mux.HandleFunc("/dataset/start", corsHandler(api.HandleDatasetStart))
	mux.HandleFunc("/dataset/stop", corsHandler(api.HandleDatasetStop))
	mux.HandleFunc("/dataset/export", corsHandler(api.HandleDatasetExport))
	mux.HandleFunc("/dataset/status", corsHandler(api.HandleDatasetStatus))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", apiPort),
		Handler: mux,
	}

	log.Printf("Attack trigger API listening on port %d", apiPort)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Error starting API server: %v", err)
	}
}

// corsHandler adds CORS headers
func corsHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// healthHandler handles health checks
func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":            "healthy",
		"generator_running": true,
		"siem_available":    eventSender.IsSIEMAvailable(),
		"timestamp":         time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processAttackTriggers processes on-demand attack requests
func processAttackTriggers() {
	for req := range api.AttackTriggerChan {
		log.Printf("Processing on-demand attack: %s (count: %d)", req.AttackType, req.Count)

		// Generate the specified number of attack events
		for i := 0; i < req.Count; i++ {
			event := api.GetAttackEvent(req.AttackType, req.Severity)
			eventSender.SendEvent(event)
			// Record triggered attack in dataset
			api.DatasetExporter.AddEvent(event, true, req.AttackType)

			// Small delay between events to prevent overwhelming the system
			if i < req.Count-1 {
				time.Sleep(time.Millisecond * 100)
			}
		}

		log.Printf("Completed on-demand attack: %s (%d events generated)", req.AttackType, req.Count)
	}
}

// extractAttackType extracts attack type from event details or message
func extractAttackType(event events.Event) string {
	// Try to extract from details first
	if attackType, ok := event.Details["attack"].(string); ok {
		return attackType
	}

	// Try to infer from message content
	message := strings.ToLower(event.Message)
	if strings.Contains(message, "brute") {
		return "brute_force"
	} else if strings.Contains(message, "scan") {
		return "port_scan"
	} else if strings.Contains(message, "malware") {
		return "malware_spread"
	} else if strings.Contains(message, "position") {
		return "v2x_position_jump"
	} else if strings.Contains(message, "flood") {
		return "v2x_message_flood"
	} else if strings.Contains(message, "signature") {
		return "v2x_invalid_signature"
	} else if strings.Contains(message, "speed") {
		return "v2x_speed_anomaly"
	}

	return "unknown_attack"
}
