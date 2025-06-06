package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"traffic-monitoring-go/app/models"
	"traffic-monitoring-go/app/siem"
	"traffic-monitoring-go/app/siem/elasticsearch"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// V2XTestHandler provides testing endpoints for V2X security rules
type V2XTestHandler struct {
	DB                 *gorm.DB
	EventIngester      *siem.EventIngester
	EnhancedRuleEngine *siem.EnhancedRuleEngine
	ESService          *elasticsearch.Service
}

// NewV2XTestHandler creates a new V2X test handler
func NewV2XTestHandler(db *gorm.DB, esService *elasticsearch.Service) *V2XTestHandler {
	return &V2XTestHandler{
		DB:                 db,
		EventIngester:      siem.NewEventIngester(db),
		EnhancedRuleEngine: siem.NewEnhancedRuleEngine(db),
		ESService:          esService,
	}
}

// TestV2XRules handles POST /test/v2x-rules - creates test events to trigger V2X security rules
func (h *V2XTestHandler) TestV2XRules(c *gin.Context) {
	// Create sample V2X events that should trigger our security rules
	testEvents := []map[string]interface{}{
		{
			"source_name": "v2x",
			"source_type": "vehicle",
			"timestamp":   time.Now(),
			"severity":    "high",
			"category":    "v2x",
			"message":     "Position jump anomaly detected in vehicle VEH-12345",
			"details": map[string]interface{}{
				"source_ip":    "192.168.1.100",
				"protocol":     "DSRC",
				"message_type": "bsm",
				"vehicle_id":   "VEH-12345",
				"position": map[string]interface{}{
					"latitude":  37.7749,
					"longitude": -122.4194,
				},
				"anomalies": []map[string]interface{}{
					{
						"type":        "position_jump",
						"confidence":  0.85,
						"description": "Vehicle moved 150m in 0.5 seconds - impossible speed",
					},
				},
			},
		},
		{
			"source_name": "v2x",
			"source_type": "vehicle",
			"timestamp":   time.Now(),
			"severity":    "critical",
			"category":    "v2x",
			"message":     "Invalid digital signature detected",
			"details": map[string]interface{}{
				"source_ip":        "192.168.1.101",
				"protocol":         "C-V2X",
				"message_type":     "bsm",
				"vehicle_id":       "VEH-67890",
				"signature_valid":  false,
				"trust_level":      0,
				"validation_error": "Certificate verification failed",
			},
		},
		{
			"source_name": "v2x",
			"source_type": "vehicle",
			"timestamp":   time.Now(),
			"severity":    "critical",
			"category":    "v2x",
			"message":     "Message flooding attack detected",
			"details": map[string]interface{}{
				"source_ip":    "192.168.1.102",
				"protocol":     "DSRC",
				"message_type": "bsm",
				"vehicle_id":   "VEH-FLOOD1",
				"anomalies": []map[string]interface{}{
					{
						"type":        "high_frequency",
						"confidence":  0.95,
						"description": "Vehicle sending 25 messages per second",
					},
				},
			},
		},
		{
			"source_name": "v2x",
			"source_type": "rsu",
			"timestamp":   time.Now(),
			"severity":    "high",
			"category":    "v2x",
			"message":     "High priority emergency alert",
			"details": map[string]interface{}{
				"source_ip":    "192.168.1.200",
				"protocol":     "C-V2X",
				"message_type": "denm",
				"priority":     9,
				"alert_type":   "emergency_vehicle",
				"description":  "Ambulance approaching intersection",
			},
		},
	}

	var results []map[string]interface{}
	var createdAlerts int

	// Process each test event
	for i, eventData := range testEvents {
		// Convert to JSON for ingestion
		eventJSON, err := json.Marshal(eventData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":       "Failed to marshal test event",
				"event_index": i,
			})
			return
		}

		// Use transaction to ensure event and alert creation
		var eventID uint
		var alertCount int64

		err = h.DB.Transaction(func(tx *gorm.DB) error {
			// Create transaction-scoped services
			ingester := siem.NewEventIngester(tx)
			ruleEngine := siem.NewEnhancedRuleEngine(tx)

			// Ingest the event
			if err := ingester.IngestEvent(eventJSON); err != nil {
				return err
			}

			// Get the created event
			var securityEvent models.SecurityEvent
			if err := tx.Last(&securityEvent).Error; err != nil {
				return err
			}
			eventID = securityEvent.ID

			// Evaluate rules
			if err := ruleEngine.EvaluateEvent(&securityEvent); err != nil {
				return err
			}

			// Count alerts created for this event
			if err := tx.Model(&models.Alert{}).Where("security_event_id = ?", eventID).Count(&alertCount).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			results = append(results, map[string]interface{}{
				"event_index": i,
				"status":      "failed",
				"error":       err.Error(),
			})
			continue
		}

		createdAlerts += int(alertCount)
		results = append(results, map[string]interface{}{
			"event_index":    i,
			"status":         "success",
			"event_id":       eventID,
			"alerts_created": alertCount,
			"event_type":     eventData["message"],
		})

		// Index in Elasticsearch if available
		if h.ESService != nil {
			var securityEvent models.SecurityEvent
			if err := h.DB.First(&securityEvent, eventID).Error; err == nil {
				go func(event models.SecurityEvent) {
					if err := h.ESService.IndexSecurityEvent(&event); err != nil {
						// Log error but don't fail the test
					}
				}(securityEvent)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":              "V2X security rules test completed",
		"events_processed":     len(testEvents),
		"total_alerts_created": createdAlerts,
		"results":              results,
		"note":                 "Check /alerts endpoint to see triggered alerts",
	})
}

// GetV2XRuleExamples handles GET /test/v2x-rules/examples - shows concrete examples of rules
func (h *V2XTestHandler) GetV2XRuleExamples(c *gin.Context) {
	var v2xRules []models.Rule

	// Get all V2X rules to show examples
	if err := h.DB.Where("category = ?", models.CategoryV2X).Find(&v2xRules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Format rules with examples of what triggers them
	ruleExamples := make([]map[string]interface{}, len(v2xRules))

	for i, rule := range v2xRules {
		var example map[string]interface{}

		// Provide specific examples based on rule name
		switch rule.Name {
		case "V2X Position Jump Detection":
			example = map[string]interface{}{
				"trigger_condition": "Vehicle moves >100m in <1 second",
				"example_scenario":  "Vehicle reports position change from (37.7749, -122.4194) to (37.7759, -122.4194) in 0.5 seconds = ~111m movement",
				"detection_logic":   "Distance calculation using haversine formula, time difference analysis",
			}
		case "V2X Speed Anomaly Detection":
			example = map[string]interface{}{
				"trigger_condition": "Speed difference >10 m/s between messages",
				"example_scenario":  "Vehicle reports 15 m/s then 30 m/s in next message (15 m/s jump)",
				"detection_logic":   "Compare speed values in consecutive BSM messages",
			}
		case "V2X Message Flooding Attack":
			example = map[string]interface{}{
				"trigger_condition": "Message frequency >10 per second",
				"example_scenario":  "Vehicle sends 25 BSM messages in 1 second window",
				"detection_logic":   "Count messages per source within sliding time window",
			}
		case "V2X Invalid Digital Signature":
			example = map[string]interface{}{
				"trigger_condition": "signature_valid = false",
				"example_scenario":  "Message received with corrupted or missing digital signature",
				"detection_logic":   "PKI signature verification against trusted CA certificates",
			}
		default:
			example = map[string]interface{}{
				"trigger_condition": rule.Condition,
				"description":       rule.Description,
			}
		}

		ruleExamples[i] = map[string]interface{}{
			"id":          rule.ID,
			"name":        rule.Name,
			"description": rule.Description,
			"condition":   rule.Condition,
			"severity":    rule.Severity,
			"status":      rule.Status,
			"example":     example,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Concrete examples of V2X security detection rules",
		"total_rules": len(v2xRules),
		"rules":       ruleExamples,
		"note":        "Use POST /test/v2x-rules to trigger these rules with test data",
	})
}
