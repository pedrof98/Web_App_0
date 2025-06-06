package database

import (
	"log"

	"traffic-monitoring-go/app/models"

	"gorm.io/gorm"
)

// CreateDefaultRules creates and enables default rules if none exist
func CreateDefaultRules(db *gorm.DB) error {
	// First create basic default rules
	if err := createBasicDefaultRules(db); err != nil {
		return err
	}

	// Then create enhanced V2X security rules
	if err := CreateV2XSecurityRules(db); err != nil {
		log.Printf("Warning: Failed to create V2X security rules: %v", err)
	}

	return nil
}

// createBasicDefaultRules creates the basic default rules (moved from original function)
func createBasicDefaultRules(db *gorm.DB) error {
	// Check if there are any basic rules (non-V2X)
	var count int64
	if err := db.Model(&models.Rule{}).Where("category != ?", models.CategoryV2X).Count(&count).Error; err != nil {
		return err
	}

	// If there are basic rules, skip creation
	if count > 0 {
		log.Printf("Basic default rules already exist (%d rules found)", count)
		return nil
	}

	defaultUser := models.User{
		Email:          "admin@example.com",
		HashedPassword: "$2a$10$SOME_HASH", // Use proper password hashing
		Role:           models.AdminRole,
	}

	// Create a default user if none exists
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return err
	}

	if userCount == 0 {
		if err := db.Create(&defaultUser).Error; err != nil {
			return err
		}
		log.Printf("Created default admin user: %s", defaultUser.Email)
	} else {
		// Get the first user
		if err := db.First(&defaultUser).Error; err != nil {
			return err
		}
	}

	rules := []models.Rule{
		{
			Name:        "Critical Severity Events",
			Description: "Alert on all critical severity events",
			Condition:   "severity = critical",
			Severity:    models.SeverityCritical,
			Category:    models.CategorySystem,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
		{
			Name:        "Authentication Failures",
			Description: "Alert on repeated authentication failures",
			Condition:   "category = authentication AND status = failure",
			Severity:    models.SeverityMedium,
			Category:    models.CategoryAuthentication,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
		{
			Name:        "Malware Detection",
			Description: "Alert on malware detection events",
			Condition:   "category = malware",
			Severity:    models.SeverityHigh,
			Category:    models.CategoryMalware,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
		{
			Name:        "Suspicious Network Activity",
			Description: "Alert on blocked network connections",
			Condition:   "category = network AND status = blocked",
			Severity:    models.SeverityMedium,
			Category:    models.CategoryNetwork,
			Status:      models.RuleStatusEnabled,
			CreatedBy:   defaultUser.ID,
		},
	}

	for _, rule := range rules {
		if err := db.Create(&rule).Error; err != nil {
			return err
		}
		log.Printf("Created default rule: %s", rule.Name)
	}

	log.Printf("Successfully created %d basic default rules", len(rules))
	return nil
}
