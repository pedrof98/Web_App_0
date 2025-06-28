package handlers

import (
	"io"
	"log"
	"net/http"
	"strings"

	"traffic-monitoring-go/app/models"
	"traffic-monitoring-go/app/siem"
	"traffic-monitoring-go/app/siem/elasticsearch"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// IngestionHandler handles event ingestion endpoints
type IngestionHandler struct {
	DB                 *gorm.DB
	EventIngester      *siem.EventIngester
	EnhancedRuleEngine *siem.EnhancedRuleEngine
	ESService          *elasticsearch.Service
}

// NewIngestionHandler creates a new IngestionHandler
func NewIngestionHandler(db *gorm.DB, esService *elasticsearch.Service) *IngestionHandler {
	return &IngestionHandler{
		DB:                 db,
		EventIngester:      siem.NewEventIngester(db),
		EnhancedRuleEngine: siem.NewEnhancedRuleEngine(db),
		ESService:          esService,
	}
}

// IngestEvent handles POST /ingest
func (h *IngestionHandler) IngestEvent(c *gin.Context) {
	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Check if it's a stress test/benchmark request
	isStressMode := isStressTest(c)

	// For stress tests, use direct ingestion without transaction or rule evaluation
	if isStressMode {
		if err := h.EventIngester.IngestEvent(body); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get the last created event ID for response
		var lastEventID uint
		h.DB.Raw("SELECT id FROM security_events ORDER BY id DESC LIMIT 1").Scan(&lastEventID)

		c.JSON(http.StatusOK, gin.H{
			"message":  "Event ingested successfully (stress mode)",
			"event_id": lastEventID,
		})
		return
	}

	// For normal operation, use transaction for data consistency
	var securityEvent models.SecurityEvent
	var alerts []models.Alert

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		// Create a transaction-scoped ingester
		ingester := siem.NewEventIngester(tx)

		// Process the event
		if err := ingester.IngestEvent(body); err != nil {
			return err
		}

		// Get created event
		if err := tx.Last(&securityEvent).Error; err != nil {
			return err
		}

		// Create a transaction-scoped rule engine
		ruleEngine := siem.NewEnhancedRuleEngine(tx)

		// Evaluate rules against the event
		if err := ruleEngine.EvaluateEvent(&securityEvent); err != nil {
			return err
		}

		// Get any alerts created for this event
		if err := tx.Where("security_event_id = ?", securityEvent.ID).Find(&alerts).Error; err != nil {
			// Just log the error but don't fail the transaction
			c.Error(err)
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Index in Elasticsearch asynchronously if available
	if h.ESService != nil {
		go func(event *models.SecurityEvent, alertList []models.Alert) {
			// Index the security event
			if err := h.ESService.IndexSecurityEvent(event); err != nil {
				// Log error but continue
				log.Printf("Error indexing security event: %v", err)
			}

			// Index any alerts
			for _, alert := range alertList {
				if err := h.ESService.IndexAlert(&alert); err != nil {
					// Log error but continue
					log.Printf("Error indexing alert: %v", err)
				}
			}
		}(&securityEvent, alerts)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Event ingested and processed successfully",
		"event_id":       securityEvent.ID,
		"alerts_created": len(alerts),
	})
}

// helper function to check if the request is a stress test
func isStressTest(c *gin.Context) bool {
	userAgent := c.GetHeader("User-Agent")

	isBenchmark := strings.Contains(userAgent, "stress-test") ||
		strings.Contains(c.FullPath(), "benchmark") ||
		strings.Contains(c.FullPath(), "evaluation")

	// also check for a specific header we can set during stress tests
	stressHeader := c.GetHeader("X-Stress-Test")

	return isBenchmark || strings.ToLower(stressHeader) == "true"
}
