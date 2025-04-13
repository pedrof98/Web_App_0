package v2x

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
	"traffic-monitoring-go/app/siem"
)

// V2XRulesManager handles V2X-specific security rules
type V2XRulesManager struct {
	DB           *gorm.DB
	RuleEngine   *siem.EnhancedRuleEngine
}

// NewV2XRulesManager creates a new V2X rules manager
func NewV2XRulesManager(db *gorm.DB, ruleEngine *siem.EnhancedRuleEngine) *V2XRulesManager {
	return &V2XRulesManager{
		DB:         db,
		RuleEngine: ruleEngine,
	}
}

// InitializeDefaultRules creates default V2X security rules
func (m *V2XRulesManager) InitializeDefaultRules(adminUserID uint) error {
	// Check if there are already V2X rules in the database
	var count int64
	err := m.DB.Model(&models.Rule{}).
		Where("category = ?", models.CategoryV2X).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to check existing rules: %v", err)
	}

	// Only create default rules if none exist
	if count > 0 {
		log.Printf("Found %d existing V2X rules, skipping defaults", count)
		return nil
	}

	// Define default V2X security rules
	defaultRules := []models.Rule{
		// BSM position jump detection rule
		{
			Name:        "V2X Abnormal Position Jump",
			Description: "Detects when a vehicle's position changes unrealistically between messages",
			Condition:   "category = v2x AND raw_data.anomalies[0].type = position_jump AND raw_data.anomalies[0].confidence > 0.7",
			Severity:    models.SeverityHigh,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   adminUserID,
		},
		// BSM speed jump detection rule
		{
			Name:        "V2X Abnormal Speed Change",
			Description: "Detects when a vehicle's speed changes unrealistically between messages",
			Condition:   "category = v2x AND raw_data.anomalies[0].type = speed_jump AND raw_data.anomalies[0].confidence > 0.7",
			Severity:    models.SeverityMedium,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   adminUserID,
		},
		// Message flooding detection
		{
			Name:        "V2X Message Flooding",
			Description: "Detects abnormally high frequency of messages from a single source",
			Condition:   "category = v2x AND raw_data.anomalies[0].type = high_frequency AND raw_data.anomalies[0].confidence > 0.8",
			Severity:    models.SeverityHigh,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   adminUserID,
		},
		// Invalid signature detection
		{
			Name:        "V2X Invalid Signature",
			Description: "Detects messages with invalid security signatures",
			Condition:   "category = v2x AND raw_data.signature_valid = false",
			Severity:    models.SeverityCritical,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   adminUserID,
		},
		// DENM/RSA high priority alerts
		{
			Name:        "V2X High Priority Alert",
			Description: "Detects high priority roadside or DENM alerts",
			Condition:   "category = v2x AND (raw_data.message_type = denm OR raw_data.message_type = rsa) AND raw_data.priority >= 7",
			Severity:    models.SeverityHigh,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   adminUserID,
		},
		// Conflicting alerts in same area
		{
			Name:        "V2X Conflicting Alerts",
			Description: "Detects conflicting roadside or emergency alerts in the same area",
			Condition:   "category = v2x AND raw_data.anomalies[0].type = conflicting_alerts",
			Severity:    models.SeverityHigh,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   adminUserID,
		},
		// Geofence violations
		{
			Name:        "V2X Geofence Violation",
			Description: "Detects vehicles sending messages from restricted geographic areas",
			Condition:   "category = v2x AND isInGeofencedArea(raw_data.position.latitude, raw_data.position.longitude, 'restricted') = true",
			Severity:    models.SeverityHigh,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusDisabled, // Disabled by default as it requires geofence setup
			CreatedBy:   adminUserID,
		},
		// Traffic signal timing anomalies
		{
			Name:        "V2X Traffic Signal Timing Anomaly",
			Description: "Detects abnormal traffic signal timing patterns",
			Condition:   "category = v2x AND raw_data.anomalies[0].type = illogical_timing",
			Severity:    models.SeverityMedium,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   adminUserID,
		},
	}

	// Create each rule in the database
	for _, rule := range defaultRules {
		rule.CreatedAt = time.Now()
		rule.UpdatedAt = time.Now()
		
		if err := m.DB.Create(&rule).Error; err != nil {
			log.Printf("Error creating rule %s: %v", rule.Name, err)
			// Continue with other rules even if one fails
			continue
		}
	}

	log.Printf("Created %d default V2X security rules", len(defaultRules))
	return nil
}

// The RegisterV2XFunctions method is removed since it referred to a non-existent method
// Instead, we'll add a method to check if coordinates are in a geofenced area

// IsInGeofencedArea checks if coordinates are within a named geofenced area
func (m *V2XRulesManager) IsInGeofencedArea(latitude, longitude float64, areaName string) (bool, error) {
	// In a real implementation, this would query a geofence database
	// For demonstration, we'll use some hardcoded values
	
	// Example restricted areas (in a real system, these would be in the database)
	restrictedAreas := map[string][][]float64{
		"restricted": {
			{37.7749, -122.4194, 0.01}, // San Francisco area with 0.01 degree radius
			{34.0522, -118.2437, 0.02}, // Los Angeles area with 0.02 degree radius
		},
		"construction": {
			{40.7128, -74.0060, 0.005}, // New York area with 0.005 degree radius
		},
	}
	
	// Check if the area name exists
	areas, exists := restrictedAreas[areaName]
	if !exists {
		return false, fmt.Errorf("geofenced area '%s' not found", areaName)
	}
	
	// Check if coordinates are within any of the areas
	for _, area := range areas {
		if len(area) < 3 {
			continue // Skip invalid areas
		}
		
		// Simple circular geofence check using Euclidean distance
		// (in a real implementation, use proper geo calculations)
		centerLat := area[0]
		centerLon := area[1]
		radius := area[2]
		
		// Calculate simple Euclidean distance (approximation)
		latDiff := latitude - centerLat
		lonDiff := longitude - centerLon
		dist := (latDiff*latDiff + lonDiff*lonDiff)
		
		if dist <= radius*radius {
			return true, nil // Within geofence
		}
	}
	
	return false, nil // Not in any geofenced area
}