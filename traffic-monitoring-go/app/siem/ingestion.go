package siem

import (
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
)

// EventIngester handles ingestion of security events from various sources
type EventIngester struct {
	DB *gorm.DB
}

// NewEventIngester creates a new EventIngester
func NewEventIngester(db *gorm.DB) *EventIngester {
	return &EventIngester{DB: db}
}


// RawEvent represents a raw security event before normalization
type RawEvent struct {
	SourceName 		string			`json:"source_name"`
	SourceType		string			`json:"source_type"`
	Timestamp		time.Time		`json:"timestamp"`
	Severity		string			`json:"severity"`
	Category		string			`json:"category"`
	Message			string			`json:"message"`
	Details			map[string]interface{}	`json:"details"`
}


// IngestEvent processes a raw event, normalizes it, and stores it
func (e *EventIngester) IngestEvent(rawEventData []byte) error {
	//Parse the raw event
	var rawEvent RawEvent
	if err := json.Unmarshal(rawEventData, &rawEvent); err != nil {
		return err
	}

	// Find or create the log source
	var logSource models.LogSource
	result := e.DB.Where("name = ?", rawEvent.SourceName).First(&logSource)
	if result.Error != nil {
		// create a new log source if it doesn't exist
		logSource = models.LogSource{
			Name:		rawEvent.SourceName,
			Type:		models.LogSourceType(rawEvent.SourceType),
			Description:	"Auto-created from ingested event",
			Enabled:	true,
		}
		if err := e.DB.Create(&logSource).Error; err != nil {
			return err
		}
	}

	// Create the security event
	securityEvent := models.SecurityEvent{
		Timestamp:	rawEvent.Timestamp,
		LogSourceID:	logSource.ID,
		Severity:	models.EventSeverity(rawEvent.Severity),
		Category:	models.EventCategory(rawEvent.Category),
		Message:	rawEvent.Message,
		RawData:	string(rawEventData),
	}

	// Extract common fields from details if present
	if rawEvent.Details != nil {
		if sourceIP, ok := rawEvent.Details["source_ip"].(string); ok {
			securityEvent.SourceIP = sourceIP
		}

		if sourcePort, ok := rawEvent.Details["source_port"].(float64); ok {
			port := int(sourcePort)
			securityEvent.SourcePort = &port
		}
		if destIP, ok := rawEvent.Details["destination_ip"].(string); ok {
			securityEvent.DestinationIP = destIP
		}
		if destPort, ok := rawEvent.Details["destination_port"].(float64); ok {
			port := int(destPort)
			securityEvent.DestinationPort = &port
		}
		if protocol, ok := rawEvent.Details["protocol"].(string); ok {
			securityEvent.Protocol = protocol
		}
		if action, ok := rawEvent.Details["action"].(string); ok {
			securityEvent.Action = action
		}
		if status, ok := rawEvent.Details["status"].(string); ok {
			securityEvent.Status = status
		}
		if deviceID, ok := rawEvent.Details["devide_id"].(string); ok {
			securityEvent.DeviceID = deviceID
		}
	}


	// save the security event
	if err := e.DB.Create(&securityEvent).Error; err != nil {
		return err
	}

	log.Printf("Ingested security event: %s (ID: %d)", securityEvent.Message, securityEvent.ID)
	return nil
}








