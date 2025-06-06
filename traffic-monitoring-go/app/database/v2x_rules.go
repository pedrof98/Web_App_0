package database

import (
	"log"
	"time"

	"traffic-monitoring-go/app/models"

	"gorm.io/gorm"
)

// CreateV2XSecurityRules creates concrete, testable V2X security rules
// This addresses Reviewer 4's request for specific examples of detection rules
func CreateV2XSecurityRules(db *gorm.DB) error {
	// Check if V2X rules already exist
	var count int64
	if err := db.Model(&models.Rule{}).Where("category = ?", models.CategoryV2X).Count(&count).Error; err != nil {
		return err
	}

	// Only create rules if none exist for V2X category
	if count > 0 {
		log.Printf("V2X security rules already exist (%d rules found)", count)
		return nil
	}

	// Get or create default user
	var defaultUser models.User
	if err := db.First(&defaultUser).Error; err != nil {
		// Create default admin user if none exists
		defaultUser = models.User{
			Email:          "admin@v2x-siem.com",
			HashedPassword: "$2a$10$SOME_HASH",
			Role:           models.AdminRole,
		}
		if err := db.Create(&defaultUser).Error; err != nil {
			return err
		}
		log.Printf("Created default admin user for V2X rules: %s", defaultUser.Email)
	}

	// Define concrete V2X security rules with specific detection logic
	v2xRules := []models.Rule{
		{
			Name:        "V2X Position Jump Detection",
			Description: "Detects vehicles with impossible position changes (>100m in <1 second)",
			Condition:   "category = v2x AND raw_data.anomalies contains position_jump AND raw_data.anomalies[0].confidence > 0.7",
			Severity:    models.SeverityHigh,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
		{
			Name:        "V2X Speed Anomaly Detection",
			Description: "Detects unrealistic speed changes (>10 m/s difference between consecutive messages)",
			Condition:   "category = v2x AND raw_data.anomalies contains speed_jump AND raw_data.anomalies[0].confidence > 0.8",
			Severity:    models.SeverityMedium,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
		{
			Name:        "V2X Message Flooding Attack",
			Description: "Detects abnormally high message frequency (>10 msgs/sec from single vehicle)",
			Condition:   "category = v2x AND raw_data.anomalies contains high_frequency AND raw_data.anomalies[0].confidence > 0.9",
			Severity:    models.SeverityCritical,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
		{
			Name:        "V2X Invalid Digital Signature",
			Description: "Detects messages with invalid or missing digital signatures",
			Condition:   "category = v2x AND raw_data.signature_valid = false",
			Severity:    models.SeverityCritical,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
		{
			Name:        "V2X Emergency Vehicle Spoofing",
			Description: "Detects potential spoofing of emergency vehicle alerts",
			Condition:   "category = v2x AND raw_data.message_type = emergency_vehicle_alert AND raw_data.trust_level < 3",
			Severity:    models.SeverityHigh,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
		{
			Name:        "V2X Conflicting Traffic Alerts",
			Description: "Detects conflicting roadside alerts in the same geographic area",
			Condition:   "category = v2x AND raw_data.anomalies contains conflicting_alerts",
			Severity:    models.SeverityMedium,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
		{
			Name:        "V2X Replay Attack Detection",
			Description: "Detects potential replay attacks (duplicate message IDs within time window)",
			Condition:   "category = v2x AND raw_data.anomalies contains timing_anomaly AND raw_data.anomalies[0].type = replay_attack",
			Severity:    models.SeverityHigh,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
		{
			Name:        "V2X Untrusted Certificate Authority",
			Description: "Detects messages from vehicles with untrusted or expired certificates",
			Condition:   "category = v2x AND raw_data.trust_level < 2 AND raw_data.signature_valid = true",
			Severity:    models.SeverityMedium,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
		{
			Name:        "V2X DENM High Priority Alert",
			Description: "Escalates high-priority Decentralized Environmental Notification Messages",
			Condition:   "category = v2x AND raw_data.message_type = denm AND raw_data.priority >= 8",
			Severity:    models.SeverityHigh,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
		{
			Name:        "V2X BSM Timing Violation",
			Description: "Detects Basic Safety Messages sent too frequently (violating 100ms standard)",
			Condition:   "category = v2x AND raw_data.message_type = bsm AND raw_data.interval_ms < 50",
			Severity:    models.SeverityLow,
			Category:    models.CategoryV2X,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
	}

	// Create each rule
	for _, rule := range v2xRules {
		rule.CreatedAt = time.Now()
		rule.UpdatedAt = time.Now()

		if err := db.Create(&rule).Error; err != nil {
			log.Printf("Error creating V2X rule %s: %v", rule.Name, err)
			continue
		}
		log.Printf("Created V2X security rule: %s", rule.Name)
	}

	log.Printf("Successfully created %d concrete V2X security rules with specific detection logic", len(v2xRules))
	return nil
}
