package dataset

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"v2x-simulator/app/attacks"
)

// DataPoint represents a single data point in our dataset
type DataPoint struct {
	Timestamp         time.Time `json:"timestamp"`
	VehicleID         uint32    `json:"vehicle_id"`
	Latitude          float64   `json:"latitude"`
	Longitude         float64   `json:"longitude"`
	Speed             float32   `json:"speed"`
	Heading           float32   `json:"heading"`
	Protocol          string    `json:"protocol"`     // DSRC or CV2X
	MessageType       string    `json:"message_type"` // BSM, SPAT, etc.
	IsAttack          bool      `json:"is_attack"`
	AttackType        string    `json:"attack_type,omitempty"`
	AttackSeverity    string    `json:"attack_severity,omitempty"`
	AttackDescription string    `json:"attack_description,omitempty"`
}

// DatasetExporter handles the export of simulation data to CSV
type DatasetExporter struct {
	mu         sync.Mutex
	dataPoints []DataPoint
	enabled    bool
}

// NewDatasetExporter creates a new dataset exporter
func NewDatasetExporter() *DatasetExporter {
	return &DatasetExporter{
		dataPoints: make([]DataPoint, 0),
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

// AddNormalMessage records a normal V2X message
func (de *DatasetExporter) AddNormalMessage(vehicleID uint32, lat, lon float64, speed, heading float32, protocol, messageType string) {
	de.mu.Lock()
	defer de.mu.Unlock()

	if !de.enabled {
		return
	}

	dataPoint := DataPoint{
		Timestamp:   time.Now(),
		VehicleID:   vehicleID,
		Latitude:    lat,
		Longitude:   lon,
		Speed:       speed,
		Heading:     heading,
		Protocol:    protocol,
		MessageType: messageType,
		IsAttack:    false,
	}

	de.dataPoints = append(de.dataPoints, dataPoint)
}

// AddAttackMessage records an attack message
func (de *DatasetExporter) AddAttackMessage(vehicleID uint32, lat, lon float64, speed, heading float32, protocol, messageType string, attack *attacks.AttackScenario) {
	de.mu.Lock()
	defer de.mu.Unlock()

	if !de.enabled {
		return
	}

	dataPoint := DataPoint{
		Timestamp:         time.Now(),
		VehicleID:         vehicleID,
		Latitude:          lat,
		Longitude:         lon,
		Speed:             speed,
		Heading:           heading,
		Protocol:          protocol,
		MessageType:       messageType,
		IsAttack:          true,
		AttackType:        string(attack.Type),
		AttackSeverity:    attack.Severity,
		AttackDescription: attack.Description,
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
		"vehicle_id",
		"latitude",
		"longitude",
		"speed",
		"heading",
		"protocol",
		"message_type",
		"is_attack",
		"attack_type",
		"attack_severity",
		"attack_description",
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	// Write data points
	for _, dp := range de.dataPoints {
		record := []string{
			dp.Timestamp.Format(time.RFC3339),
			fmt.Sprintf("%d", dp.VehicleID),
			fmt.Sprintf("%.8f", dp.Latitude),
			fmt.Sprintf("%.8f", dp.Longitude),
			fmt.Sprintf("%.2f", dp.Speed),
			fmt.Sprintf("%.2f", dp.Heading),
			dp.Protocol,
			dp.MessageType,
			fmt.Sprintf("%t", dp.IsAttack),
			dp.AttackType,
			dp.AttackSeverity,
			dp.AttackDescription,
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
	de.dataPoints = make([]DataPoint, 0)
}

// GetStats returns statistics about the collected data
func (de *DatasetExporter) GetStats() map[string]interface{} {
	de.mu.Lock()
	defer de.mu.Unlock()

	totalMessages := len(de.dataPoints)
	attackMessages := 0
	normalMessages := 0

	protocolCounts := make(map[string]int)
	attackTypeCounts := make(map[string]int)

	for _, dp := range de.dataPoints {
		protocolCounts[dp.Protocol]++

		if dp.IsAttack {
			attackMessages++
			attackTypeCounts[dp.AttackType]++
		} else {
			normalMessages++
		}
	}

	attackRate := 0.0
	if totalMessages > 0 {
		attackRate = float64(attackMessages) / float64(totalMessages) * 100
	}

	return map[string]interface{}{
		"total_messages":      totalMessages,
		"normal_messages":     normalMessages,
		"attack_messages":     attackMessages,
		"attack_rate_percent": attackRate,
		"protocol_counts":     protocolCounts,
		"attack_type_counts":  attackTypeCounts,
		"collection_enabled":  de.enabled,
	}
}
