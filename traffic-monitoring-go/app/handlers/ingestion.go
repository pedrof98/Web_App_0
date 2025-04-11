package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
	"traffic-monitoring-go/app/siem"
	"traffic-monitoring-go/app/siem/elasticsearch"
)

// IngestionHandler handles event ingestion endpoints
type IngestionHandler struct {
	DB                *gorm.DB
	EventIngester     *siem.EventIngester
	EnhancedRuleEngine *siem.EnhancedRuleEngine
	ESService         *elasticsearch.Service
}

// NewIngestionHandler creates a new IngestionHandler
func NewIngestionHandler(db *gorm.DB, esService *elasticsearch.Service) *IngestionHandler {
	return &IngestionHandler{
		DB:                db,
		EventIngester:     siem.NewEventIngester(db),
		EnhancedRuleEngine: siem.NewEnhancedRuleEngine(db),
		ESService:         esService,
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

	// Use a transaction for both ingestion and rule evaluation
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

	// Index in Elasticsearch if available
	if h.ESService != nil {
		// Index the security event
		if err := h.ESService.IndexSecurityEvent(&securityEvent); err != nil {
			// Log the error but don't fail the request
			c.Error(err)
		}

		// Index any alerts
		for _, alert := range alerts {
			if err := h.ESService.IndexAlert(&alert); err != nil {
				// Log the error but don't fail the request
				c.Error(err)
			}
		}
	}

	// Check if there were Elasticsearch indexing errors
	if len(c.Errors) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Event ingested and processed with Elasticsearch indexing warnings",
			"event_id": securityEvent.ID,
			"alerts_created": len(alerts),
			"warnings": c.Errors.Errors(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event ingested and processed successfully",
		"event_id": securityEvent.ID,
		"alerts_created": len(alerts),
	})
}