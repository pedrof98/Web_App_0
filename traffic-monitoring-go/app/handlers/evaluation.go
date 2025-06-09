package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"traffic-monitoring-go/app/models"
	"traffic-monitoring-go/app/siem"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EvaluationHandler struct {
	DB                 *gorm.DB
	EventIngester      *siem.EventIngester
	EnhancedRuleEngine *siem.EnhancedRuleEngine
}

func NewEvaluationHandler(db *gorm.DB) *EvaluationHandler {
	return &EvaluationHandler{
		DB:                 db,
		EventIngester:      siem.NewEventIngester(db),
		EnhancedRuleEngine: siem.NewEnhancedRuleEngine(db),
	}
}

// EvaluateV2XScenarios handles POST /evaluation/v2x-scenarios
func (h *EvaluationHandler) EvaluateV2XScenarios(c *gin.Context) {
	var params struct {
		LegitimateEvents int    `json:"legitimate_events" default:"100"`
		MaliciousEvents  int    `json:"malicious_events" default:"20"`
		ScenarioType     string `json:"scenario_type" default:"mixed"`
	}

	if err := c.ShouldBindJSON(&params); err != nil {
		params.LegitimateEvents = 100
		params.MaliciousEvents = 20
		params.ScenarioType = "mixed"
	}

	// Record initial state
	var initialMem runtime.MemStats
	runtime.ReadMemStats(&initialMem)
	startTime := time.Now()

	var initialEvents, initialAlerts int64
	h.DB.Model(&models.SecurityEvent{}).Count(&initialEvents)
	h.DB.Model(&models.Alert{}).Count(&initialAlerts)

	// Generate evaluation dataset
	results, err := h.generateV2XDataset(params.LegitimateEvents, params.MaliciousEvents, params.ScenarioType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate final metrics
	var finalEvents, finalAlerts int64
	h.DB.Model(&models.SecurityEvent{}).Count(&finalEvents)
	h.DB.Model(&models.Alert{}).Count(&finalAlerts)

	var finalMem runtime.MemStats
	runtime.ReadMemStats(&finalMem)
	totalDuration := time.Since(startTime)

	// Calculate detection accuracy
	eventsProcessed := finalEvents - initialEvents
	alertsGenerated := finalAlerts - initialAlerts

	// Safe calculations to avoid division by zero
	var eventsPerSecond, alertRatePercent float64
	if totalDuration.Seconds() > 0 {
		eventsPerSecond = float64(eventsProcessed) / totalDuration.Seconds()
	}
	if eventsProcessed > 0 {
		alertRatePercent = float64(alertsGenerated) / float64(eventsProcessed) * 100
	}

	// Safe memory calculation
	var memoryUsedMB float64
	if finalMem.Alloc > initialMem.Alloc {
		memoryUsedMB = float64(finalMem.Alloc-initialMem.Alloc) / 1024 / 1024
	}

	evaluation := map[string]interface{}{
		"scenario_results": results,
		"performance_metrics": map[string]interface{}{
			"total_events_processed": eventsProcessed,
			"alerts_generated":       alertsGenerated,
			"processing_time_ms":     totalDuration.Milliseconds(),
			"events_per_second":      eventsPerSecond,
			"memory_used_mb":         memoryUsedMB,
		},
		"detection_metrics": map[string]interface{}{
			"expected_malicious":  params.MaliciousEvents,
			"expected_legitimate": params.LegitimateEvents,
			"total_alerts":        alertsGenerated,
			"alert_rate_percent":  alertRatePercent,
		},
	}

	c.JSON(http.StatusOK, evaluation)
}

func (h *EvaluationHandler) generateV2XDataset(legitimate, malicious int, scenarioType string) (map[string]interface{}, error) {
	var legitimateProcessed, maliciousProcessed int
	var legitimateAlerts, maliciousAlerts int64

	// Generate legitimate V2X traffic
	for i := 0; i < legitimate; i++ {
		event := h.createLegitimateV2XEvent(i)
		eventJSON, _ := json.Marshal(event)

		if err := h.EventIngester.IngestEvent(eventJSON); err != nil {
			fmt.Printf("Error ingesting legitimate event %d: %v\n", i, err)
			continue
		}
		legitimateProcessed++

		// Get the created event and evaluate rules
		var securityEvent models.SecurityEvent
		if err := h.DB.Last(&securityEvent).Error; err == nil {
			// Evaluate rules for this event
			if err := h.EnhancedRuleEngine.EvaluateEvent(&securityEvent); err != nil {
				fmt.Printf("Error evaluating rules for event %d: %v\n", securityEvent.ID, err)
			}

			// Count alerts for this specific event
			var alertCount int64
			h.DB.Model(&models.Alert{}).Where("security_event_id = ?", securityEvent.ID).Count(&alertCount)
			legitimateAlerts += alertCount
		}
	}

	// Generate malicious V2X traffic
	for i := 0; i < malicious; i++ {
		event := h.createMaliciousV2XEvent(i)
		eventJSON, _ := json.Marshal(event)

		if err := h.EventIngester.IngestEvent(eventJSON); err != nil {
			fmt.Printf("Error ingesting malicious event %d: %v\n", i, err)
			continue
		}
		maliciousProcessed++

		// Get the created event and evaluate rules
		var securityEvent models.SecurityEvent
		if err := h.DB.Last(&securityEvent).Error; err == nil {
			fmt.Printf("Evaluating rules for malicious event ID %d with category %s\n", securityEvent.ID, securityEvent.Category)

			// Evaluate rules for this event
			if err := h.EnhancedRuleEngine.EvaluateEvent(&securityEvent); err != nil {
				fmt.Printf("Error evaluating rules for event %d: %v\n", securityEvent.ID, err)
			} else {
				fmt.Printf("Successfully evaluated rules for event %d\n", securityEvent.ID)
			}

			// Count alerts for this specific event
			var alertCount int64
			h.DB.Model(&models.Alert{}).Where("security_event_id = ?", securityEvent.ID).Count(&alertCount)
			fmt.Printf("Generated %d alerts for event %d\n", alertCount, securityEvent.ID)
			maliciousAlerts += alertCount
		}
	}

	// Calculate detection rates with safety checks
	var falsePositiveRate, truePositiveRate float64

	if legitimateProcessed > 0 {
		falsePositiveRate = float64(legitimateAlerts) / float64(legitimateProcessed) * 100
	}

	if maliciousProcessed > 0 {
		truePositiveRate = float64(maliciousAlerts) / float64(maliciousProcessed) * 100
	}

	return map[string]interface{}{
		"legitimate_events": map[string]interface{}{
			"generated":           legitimateProcessed,
			"alerts_triggered":    legitimateAlerts,
			"false_positive_rate": falsePositiveRate,
		},
		"malicious_events": map[string]interface{}{
			"generated":        maliciousProcessed,
			"alerts_triggered": maliciousAlerts,
			"detection_rate":   truePositiveRate,
		},
		"scenario_type": scenarioType,
	}, nil
}

func (h *EvaluationHandler) createLegitimateV2XEvent(index int) map[string]interface{} {
	return map[string]interface{}{
		"source_name": "v2x",
		"source_type": "vehicle",
		"timestamp":   time.Now(),
		"severity":    "info",
		"category":    "v2x",
		"message":     "Normal BSM from vehicle",
		"details": map[string]interface{}{
			"vehicle_id":   "VEH-LEGIT-" + fmt.Sprintf("%03d", index),
			"message_type": "bsm",
			"protocol":     "DSRC",
			"speed":        25 + rand.Intn(15), // Normal speeds
			"position": map[string]interface{}{
				"latitude":  37.7749 + rand.Float64()*0.001,
				"longitude": -122.4194 + rand.Float64()*0.001,
			},
			"signature_valid": true,
			"trust_level":     8,
		},
	}
}

func (h *EvaluationHandler) createMaliciousV2XEvent(index int) map[string]interface{} {
	// Randomly select attack type
	attackTypes := []string{"position_jump", "speed_anomaly", "message_flood", "invalid_signature"}
	attackType := attackTypes[rand.Intn(len(attackTypes))]

	event := map[string]interface{}{
		"source_name": "v2x",
		"source_type": "vehicle",
		"timestamp":   time.Now(),
		"severity":    "high",
		"category":    "v2x",
		"message":     fmt.Sprintf("Malicious V2X activity: %s", attackType),
		"details": map[string]interface{}{
			"vehicle_id":   fmt.Sprintf("VEH-ATTACK-%03d", index),
			"message_type": "bsm",
			"protocol":     "DSRC",
			"source_ip":    fmt.Sprintf("192.168.1.%d", 100+index),
			"attack_type":  attackType,
		},
	}

	// Add attack-specific details that will trigger our V2X rules
	switch attackType {
	case "position_jump":
		event["details"].(map[string]interface{})["anomalies"] = []map[string]interface{}{
			{
				"type":        "position_jump",
				"confidence":  0.85,
				"description": "Impossible position change detected",
			},
		}
	case "invalid_signature":
		event["details"].(map[string]interface{})["signature_valid"] = false
		event["details"].(map[string]interface{})["trust_level"] = 0
	case "message_flood":
		event["details"].(map[string]interface{})["anomalies"] = []map[string]interface{}{
			{
				"type":        "high_frequency",
				"confidence":  0.95,
				"description": "Message flooding detected",
			},
		}
	case "speed_anomaly":
		event["details"].(map[string]interface{})["anomalies"] = []map[string]interface{}{
			{
				"type":        "speed_jump",
				"confidence":  0.88,
				"description": "Unrealistic speed change detected",
			},
		}
	}

	return event
}
