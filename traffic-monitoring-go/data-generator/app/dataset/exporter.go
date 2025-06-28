package dataset

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"traffic-monitoring-go/data-generator/app/events"
)

// EventDataPoint represents a single event data point in our dataset
type EventDataPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	SourceName string    `json:"source_name"`
	SourceType string    `json:"source_type"`
	Severity   string    `json:"severity"`
	Category   string    `json:"category"`
	Message    string    `json:"message"`
	SourceIP   string    `json:"source_ip,omitempty"`
	DestIP     string    `json:"dest_ip,omitempty"`
	SourcePort int       `json:"source_port,omitempty"`
	DestPort   int       `json:"dest_port,omitempty"`
	Protocol   string    `json:"protocol,omitempty"`
	VehicleID  string    `json:"vehicle_id,omitempty"`
	IsAttack   bool      `json:"is_attack"`
	AttackType string    `json:"attack_type,omitempty"`
}

// DatasetExporter handles the export of generated events to CSV
type DatasetExporter struct {
	mu         sync.Mutex
	dataPoints []EventDataPoint
	enabled    bool
}

// NewDatasetExporter creates a new dataset exporter
func NewDatasetExporter() *DatasetExporter {
	return &DatasetExporter{
		dataPoints: make([]EventDataPoint, 0),
		enabled:    false,
	}
}

// Enable starts data collection
func (de *DatasetExporter) Enable() {
	de.mu.Lock()
	defer de.mu.Unlock()
	de.enabled = true
}

// Disable stops data collection
func (de *DatasetExporter) Disable() {
	de.mu.Lock()
	defer de.mu.Unlock()
	de.enabled = false
}

// IsEnabled returns whether data collection is active
func (de *DatasetExporter) IsEnabled() bool {
	de.mu.Lock()
	defer de.mu.Unlock()
	return de.enabled
}

// AddEvent records an event in the dataset
func (de *DatasetExporter) AddEvent(event events.Event, isAttack bool, attackType string) {
	de.mu.Lock()
	defer de.mu.Unlock()

	if !de.enabled {
		return
	}

	// Extract details from event
	sourceIP := extractStringDetail(event.Details, "source_ip")
	destIP := extractStringDetail(event.Details, "destination_ip")
	protocol := extractStringDetail(event.Details, "protocol")
	vehicleID := extractStringDetail(event.Details, "vehicle_id")

	sourcePort := extractIntDetail(event.Details, "source_port")
	destPort := extractIntDetail(event.Details, "destination_port")

	dataPoint := EventDataPoint{
		Timestamp:  event.Timestamp,
		SourceName: event.SourceName,
		SourceType: event.SourceType,
		Severity:   event.Severity,
		Category:   event.Category,
		Message:    event.Message,
		SourceIP:   sourceIP,
		DestIP:     destIP,
		SourcePort: sourcePort,
		DestPort:   destPort,
		Protocol:   protocol,
		VehicleID:  vehicleID,
		IsAttack:   isAttack,
		AttackType: attackType,
	}

	de.dataPoints = append(de.dataPoints, dataPoint)
}

// GetDataCount returns the number of collected data points
func (de *DatasetExporter) GetDataCount() int {
	de.mu.Lock()
	defer de.mu.Unlock()
	return len(de.dataPoints)
}

// ExportToCSV exports the collected data to a CSV file
func (de *DatasetExporter) ExportToCSV(filename string) error {
	de.mu.Lock()
	defer de.mu.Unlock()

	if len(de.dataPoints) == 0 {
		return fmt.Errorf("no data to export")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Create CSV file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	header := []string{
		"timestamp",
		"source_name",
		"source_type",
		"severity",
		"category",
		"message",
		"source_ip",
		"dest_ip",
		"source_port",
		"dest_port",
		"protocol",
		"vehicle_id",
		"is_attack",
		"attack_type",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	// Write data points
	for _, dp := range de.dataPoints {
		record := []string{
			dp.Timestamp.Format(time.RFC3339),
			dp.SourceName,
			dp.SourceType,
			dp.Severity,
			dp.Category,
			dp.Message,
			dp.SourceIP,
			dp.DestIP,
			fmt.Sprintf("%d", dp.SourcePort),
			fmt.Sprintf("%d", dp.DestPort),
			dp.Protocol,
			dp.VehicleID,
			fmt.Sprintf("%t", dp.IsAttack),
			dp.AttackType,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %v", err)
		}
	}

	return nil
}

// ClearData removes all collected data points
func (de *DatasetExporter) ClearData() {
	de.mu.Lock()
	defer de.mu.Unlock()
	de.dataPoints = make([]EventDataPoint, 0)
}

// GetStats returns statistics about the collected data
func (de *DatasetExporter) GetStats() map[string]interface{} {
	de.mu.Lock()
	defer de.mu.Unlock()

	totalEvents := len(de.dataPoints)
	attackEvents := 0
	normalEvents := 0

	categoryCounts := make(map[string]int)
	severityCounts := make(map[string]int)
	attackTypeCounts := make(map[string]int)

	for _, dp := range de.dataPoints {
		categoryCounts[dp.Category]++
		severityCounts[dp.Severity]++

		if dp.IsAttack {
			attackEvents++
			attackTypeCounts[dp.AttackType]++
		} else {
			normalEvents++
		}
	}

	attackRate := 0.0
	if totalEvents > 0 {
		attackRate = float64(attackEvents) / float64(totalEvents) * 100
	}

	return map[string]interface{}{
		"total_events":        totalEvents,
		"normal_events":       normalEvents,
		"attack_events":       attackEvents,
		"attack_rate_percent": attackRate,
		"category_counts":     categoryCounts,
		"severity_counts":     severityCounts,
		"attack_type_counts":  attackTypeCounts,
		"collection_enabled":  de.enabled,
	}
}

// Helper functions to extract details safely
func extractStringDetail(details map[string]interface{}, key string) string {
	if val, ok := details[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func extractIntDetail(details map[string]interface{}, key string) int {
	if val, ok := details[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}
