package siem

import (
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
)

// RuleEngine evaluates ecurity events against rules
type RuleEngine struct {
	DB *gorm.DB
}


// NewRuleEngine creates a new RuleEngine
func NewRuleEngine(db *gorm.DB) *RuleEngine {
	return &RuleEngine{DB: db}
}


// EvaluateEvent checks an event against all enabled rules and creates alerts if matches
func (e *RuleEngine) EvaluateEvent(event *models.SecurityEvent) error {
	// get all enabled rules
	var rules []models.Rule
	if err := e.DB.Where("status = ?", models.RuleStatusEnabled).Find(&rules).Error; err != nil {
		return err
	}

	//Evaluate each rule against the event
	for _, rule := range rules {
		matched, err := e.evaluateRule(event, &rule)
		if err != nil {
			log.Printf("Error evaluating rule %s: %v", rule.Name, err)
			continue
		}

		if matched {
			// create an alert
			alert := models.Alert{
				RuleID:			rule.ID,
				SecurityEventID:	event.ID,
				Timestamp:		time.Now(),
				Severity:		rule.Severity,
				Status:			models.AlertStatusOpen,
			}

			if err := e.DB.Create(&alert).Error; err != nil {
				log.Printf("Error creating alert for rule %s: %v", rule.Name, err)
				continue
			}

			log.Printf("Created alert for rule: %s, event: %d", rule.Name, event.ID)
		}
	}
	
	return nil
}


// evaluateRule checks if an event matches a rule
// this is a simple implementatio that will be enhanced later
func (e *RuleEngine) evaluateRule(event *models.SecurityEvent, rule *models.Rule) (bool, error) {
	// split the condition into parts (very basic rule language for now)
	conditions := strings.Split(rule.Condition, " AND ")

	for _, condition := range conditions {
		parts := strings.Split(strings.TrimSpace(condition), " ")

		if len(parts) != 3 {
			continue // skip invalid conditions
		}

		field := parts[0]
		operator := parts[1]
		value := parts[2]

		// check field values based on rule conditions
		switch field {
		case "severity":
			if !evaluateCondition(string(event.Severity), operator, value) {
				return false, nil
			}
		case "category":
			if !evaluateCondition(string(event.Category), operator, value) {
				return false, nil
			}
		case "source_ip":
			if !evaluateCondition(event.SourceIP, operator, value) {
				return false, nil
			}
		case "destination_ip":
			if !evaluateCondition(event.DestinationIP, operator, value) {
				return false, nil
			}
		case "protocol":
			if !evaluateCondition(event.Protocol, operator, value) {
				return false, nil
			}
		case "action":
			if !evaluateCondition(event.Action, operator, value) {
				return false, nil
			}
		case "status":
			if !evaluateCondition(event.Status, operator, value) {
				return false, nil
			}
		}
	}

	// if we reach this point, all conditions matched
	return true, nil
}


// evaluateCondition compares values based on the operator
func evaluateCondition(fieldValue, operator, ruleValue string) bool {
	switch operator {
	case "=", "==":
		return fieldValue == ruleValue
	case "!=":
		return fieldValue != ruleValue
	case "contains":
		return strings.Contains(fieldValue, ruleValue)
	case "startswith":
		return strings.HasPrefix(fieldValue, ruleValue)
	case "endswith":
		return strings.HasSuffix(fieldValue, ruleValue)
	default:
		return false
	}
}






















