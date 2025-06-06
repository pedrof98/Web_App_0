package siem

import (
	"encoding/json"
	"log"
	"time"

	"traffic-monitoring-go/app/models"

	"gorm.io/gorm"
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
	SourceName string                 `json:"source_name"`
	SourceType string                 `json:"source_type"`
	Timestamp  time.Time              `json:"timestamp"`
	Severity   string                 `json:"severity"`
	Category   string                 `json:"category"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details"`
}

// IngestEvent processes a raw event, normalizes it, and stores it
func (e *EventIngester) IngestEvent(rawEventData []byte) error {
	//Parse the raw event
	var rawEvent RawEvent
	if err := json.Unmarshal(rawEventData, &rawEvent); err != nil {
		return err
	}

	// Find or create the log source
	//var logSource models.LogSource
	//result := e.DB.Where("name = ?", rawEvent.SourceName).First(&logSource)
	//if result.Error != nil {
	// create a new log source if it doesn't exist
	//	logSource = models.LogSource{
	//		Name:		rawEvent.SourceName,
	//		Type:		models.LogSourceType(rawEvent.SourceType),
	//		Description:	"Auto-created from ingested event",
	//		Enabled:	true,
	//	}
	//	if err := e.DB.Create(&logSource).Error; err != nil {
	//		return err
	//	}

	// Use Raw SQL for log source lookup (faster for large datasets)
	var logSourceID uint
	err := e.DB.Raw("SELECT id FROM log_sources WHERE name = ? LIMIT 1", rawEvent.SourceName).Scan(&logSourceID).Error

	if err != nil || logSourceID == 0 {
		// create a new log source witrh raw SQL
		result := e.DB.Exec(`
			INSERT INTO log_sources (name, type, description, enabled, created_at, updated_at)
			VALUES (?, ?, 'Auto-created', true, NOW(), NOW())
			ON CONFLICT (name) DO UPDATE SET updated_at = NOW()
			RETURNING id`, rawEvent.SourceName, rawEvent.SourceType)

		if result.Error != nil {
			return result.Error
		}

		// get the ID
		e.DB.Raw("SELECT id FROM log_sources WHERE name = ?", rawEvent.SourceName).Scan(&logSourceID)
	}

	// Create the security event
	securityEvent := models.SecurityEvent{
		Timestamp:   rawEvent.Timestamp,
		LogSourceID: logSourceID,
		Severity:    models.EventSeverity(rawEvent.Severity),
		Category:    models.EventCategory(rawEvent.Category),
		Message:     rawEvent.Message,
		RawData:     string(rawEventData),
	}

	// Extract common fields from details if present
	if details := rawEvent.Details; details != nil {
		if sourceIP, ok := details["source_ip"].(string); ok {
			securityEvent.SourceIP = sourceIP
		}
		if sourcePort, ok := details["source_port"].(float64); ok {
			port := int(sourcePort)
			securityEvent.SourcePort = &port
		}
		if destIP, ok := details["destination_ip"].(string); ok {
			securityEvent.DestinationIP = destIP
		}
		if destPort, ok := details["destination_port"].(float64); ok {
			port := int(destPort)
			securityEvent.DestinationPort = &port
		}
		if protocol, ok := details["protocol"].(string); ok {
			securityEvent.Protocol = protocol
		}
		if action, ok := details["action"].(string); ok {
			securityEvent.Action = action
		}
		if status, ok := details["status"].(string); ok {
			securityEvent.Status = status
		}
		if deviceID, ok := details["device_id"].(string); ok {
			securityEvent.DeviceID = deviceID
		}
	}

	// Save the security event with optimized performance
	if err := e.DB.Session(&gorm.Session{SkipDefaultTransaction: true}).Create(&securityEvent).Error; err != nil {
		return err
	}

	log.Printf("Ingested security event: %s (ID: %d)", securityEvent.Message, securityEvent.ID)
	return nil
}
