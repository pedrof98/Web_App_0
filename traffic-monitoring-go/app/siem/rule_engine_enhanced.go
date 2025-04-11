package siem

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"strconv"

	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
)

// EnhancedRuleEngine is an improved rule evaluation engine
type EnhancedRuleEngine struct {
	DB *gorm.DB
}


// NewEnhancedRuleEngine creates a new EnhancedRuleEngine
func NewEnhancedRuleEngine(db *gorm.DB) *EnhancedRuleEngine {
	return &EnhancedRuleEngine{DB: db}
}


// EvaluateEvent checks an event against all enabled rules and creates alerts if matched
func (e *EnhancedRuleEngine) EvaluateEvent(event *models.SecurityEvent) error {
	// get all enabled rules
	var rules []models.Rule
	if err := e.DB.Where("status = ?", models.RuleStatusEnabled).Find(&rules).Error; err != nil {
		return err
	}

	// evaluate each rule against the event
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
func (e *EnhancedRuleEngine) evaluateRule(event *models.SecurityEvent, rule *models.Rule) (bool, error) {
	// Parse rule condition
	condition := rule.Condition

	// support for complex conditions with AND, OR, and NOT operators
	// simplified parser, in a real system you'd use a proper expression parser

	// first handle NOT operators
	notPattern := regexp.MustCompile(`NOT\s+\(([^)]+)\)`)
	for notPattern.MatchString(condition) {
		condition = notPattern.ReplaceAllStringFunc(condition, func(match string) string {
			// extract the condition inside NOT()
			subExpr := notPattern.FindStringSubmatch(match)[1]

			// evaluate the sub-expression
			result, err := e.evaluateSimpleCondition(event, subExpr)
			if err != nil {
				log.Printf("Error evaluating NOT condition: %v", err)
				return "false" // default to false on error
			}

			// return the negated result
			return strconv.FormatBool(!result)
		})
	}

	// handle AND operators
	if strings.Contains(condition, " AND ") {
		parts := strings.Split(condition, " AND ")
		for _, part := range parts {
			result, err := e.evaluateSimpleCondition(event, strings.TrimSpace(part))
			if err != nil {
				return false, err
			}
			if !result {
				return false, nil // short-circuit on first false condition
			}
		}
		return true, nil // all conditions were true
	}

	// handle OR operators
	if strings.Contains(condition, " OR ") {
		parts := strings.Split(condition, " OR ")
		for _, part := range parts {
			result, err := e.evaluateSimpleCondition(event, strings.TrimSpace(part))
			if err != nil {
				return false, err
			}
			if result {
				return true, nil // short-circuit on first true condition
			}
		}
		return false, nil // no conditions were true
	}

	// if no AND or OR, it's a simple condition
	return e.evaluateSimpleCondition(event, condition)
}


// evaluateSimpleCondition evaluates a single condition against an event
func (e *EnhancedRuleEngine) evaluateSimpleCondition(event *models.SecurityEvent, condition string) (bool, error) {
	// handle true/false literals
	if condition == "true" {
		return true, nil
	}
	if condition == "false" {
		return false, nil
	}

	// Parse condition in the format "field operator value"
	parts := strings.SplitN(condition, " ", 3)
	if len(parts) != 3 {
		return false, fmt.Errorf("invalid condition format: %s", condition)
	}

	field := parts[0]
	operator := parts[1]
	value := parts[2]

	// extract value from event based on field
	var fieldValue interface{}

	// handle nested JSON fields
	if strings.Contains(field, ".") {
		// if the field refers to the raw data as JSON
		if strings.HasPrefix(field, "raw_data.") {
			var rawData map[string]interface{}
			if err := json.Unmarshal([]byte(event.RawData), &rawData); err != nil {
				return false, fmt.Errorf("error parsing raw data JSON: %v", err)
			}

			// extract nested field
			nestedField := field[9:] // remove "raw_data." prefix
			parts := strings.Split(nestedField, ".")

			// Navigate through the nested structure
			current := rawData
			for i, part := range parts {
				if i == len(parts)-1 {
					fieldValue = current[part]
					break
				}

				next, ok := current[part].(map[string]interface{})
				if !ok {
					return false, fmt.Errorf("field not found or not an object: %s", part)
				}
				current = next
			}
		}
	} else {
		// handle direct fields
		switch field {
		case "severity":
			fieldValue = string(event.Severity)
		case "category":
			fieldValue = string(event.Category)
		case "source_ip":
			fieldValue = event.SourceIP
		case "destination_ip":
			fieldValue = event.DestinationIP
		case "protocol":
			fieldValue = event.Protocol
		case "action":
			fieldValue = event.Action
		case "status":
			fieldValue = event.Status
		case "message":
			fieldValue = event.Message
		case "source_port":
			if event.SourcePort != nil {
				fieldValue = *event.SourcePort
			}
		case "destination_port":
			if event.DestinationPort != nil {
				fieldValue = *event.DestinationPort
			}
		case "device_id":
			fieldValue = event.DeviceID
		default:
			return false, fmt.Errorf("unknown field: %s", field)
		}
	}

	// Handle null/nil values
	if fieldValue == nil {
		// Special case for operators that work with null
		switch operator {
		case "is", "=", "==":
			return strings.ToLower(value) == "null", nil
		case "is not", "!=", "<>":
			return strings.ToLower(value) != "null", nil
		default:
			return false, nil // All other operations on null return false
		}
	}


	// Compare based on field type and operator
	switch v := fieldValue.(type) {
	case string:
		return compareString(v, operator, value)
	case int, int32, int64, uint, uint32, uint64:
		numValue := fmt.Sprintf("%d", v)
		return compareNumber(numValue, operator, value)
	case float32, float64:
		numValue := fmt.Sprintf("%f", v)
		return compareNumber(numValue, operator, value)
	case bool:
		switch strings.ToLower(value) {
		case "true":
			return compareBoolean(v, operator, true)
		case "false":
			return compareBoolean(v, operator, false)
		default:
			return false, fmt.Errorf("invalid boolean value: %s", value)
		}
	case time.Time:
		return compareTime(v, operator, value)
	default:
		// Convert to string as fallback
		strValue := fmt.Sprintf("%v", v)
		return compareString(strValue, operator, value)
	}
}



// compareString compares string values
func compareString(fieldValue, operator, ruleValue string) (bool, error) {
	switch operator {
	case "=", "==", "is":
		return fieldValue == ruleValue, nil
	case "!=", "<>", "is not":
		return fieldValue != ruleValue, nil
	case "contains":
		return strings.Contains(fieldValue, ruleValue), nil
	case "not contains":
		return !strings.Contains(fieldValue, ruleValue), nil
	case "startswith":
		return strings.HasPrefix(fieldValue, ruleValue), nil
	case "endswith":
		return strings.HasSuffix(fieldValue, ruleValue), nil
	case "matches":
		// Regular expression matching
		matched, err := regexp.MatchString(ruleValue, fieldValue)
		if err != nil {
			return false, fmt.Errorf("invalid regex: %v", err)
		}
		return matched, nil
	default:
		return false, fmt.Errorf("unsupported string operator: %s", operator)
	}
}

// compareNumber compares numeric values
func compareNumber(fieldValue, operator, ruleValue string) (bool, error) {
	// Parse the numbers
	fieldNum, err := strconv.ParseFloat(fieldValue, 64)
	if err != nil {
		return false, fmt.Errorf("failed to parse field value as number: %v", err)
	}

	ruleNum, err := strconv.ParseFloat(ruleValue, 64)
	if err != nil {
		return false, fmt.Errorf("failed to parse rule value as number: %v", err)
	}

	switch operator {
	case "=", "==", "is":
		return fieldNum == ruleNum, nil
	case "!=", "<>", "is not":
		return fieldNum != ruleNum, nil
	case ">":
		return fieldNum > ruleNum, nil
	case ">=":
		return fieldNum >= ruleNum, nil
	case "<":
		return fieldNum < ruleNum, nil
	case "<=":
		return fieldNum <= ruleNum, nil
	default:
		return false, fmt.Errorf("unsupported numeric operator: %s", operator)
	}
}

// compareBoolean compares boolean values
func compareBoolean(fieldValue bool, operator string, ruleValue bool) (bool, error) {
	switch operator {
	case "=", "==", "is":
		return fieldValue == ruleValue, nil
	case "!=", "<>", "is not":
		return fieldValue != ruleValue, nil
	default:
		return false, fmt.Errorf("unsupported boolean operator: %s", operator)
	}
}

// compareTime compares time values
func compareTime(fieldValue time.Time, operator, ruleValue string) (bool, error) {
	// Parse the rule time value
	var ruleTime time.Time
	var err error

	// Check for special time values
	switch ruleValue {
	case "now":
		ruleTime = time.Now()
	case "today":
		now := time.Now()
		ruleTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "yesterday":
		now := time.Now()
		ruleTime = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
	default:
		// Try various time formats
		formats := []string{
			time.RFC3339,
			"2006-01-02",
			"2006-01-02 15:04:05",
			"2006/01/02",
			"01/02/2006",
		}

		for _, format := range formats {
			ruleTime, err = time.Parse(format, ruleValue)
			if err == nil {
				break
			}
		}

		if err != nil {
			return false, fmt.Errorf("failed to parse time value: %s", ruleValue)
		}
	}

	// Special handling for relative time expressions
	if strings.HasPrefix(ruleValue, "-") && strings.Contains(ruleValue, " ") {
		// e.g., "-1 hour", "-30 minutes"
		parts := strings.SplitN(ruleValue, " ", 2)
		if len(parts) != 2 {
			return false, fmt.Errorf("invalid relative time format: %s", ruleValue)
		}

		num, err := strconv.Atoi(parts[0][1:]) // Remove the "-" and parse
		if err != nil {
			return false, fmt.Errorf("invalid relative time quantity: %s", parts[0])
		}

		unit := strings.TrimSpace(parts[1])
		switch unit {
		case "second", "seconds":
			ruleTime = time.Now().Add(time.Duration(-num) * time.Second)
		case "minute", "minutes":
			ruleTime = time.Now().Add(time.Duration(-num) * time.Minute)
		case "hour", "hours":
			ruleTime = time.Now().Add(time.Duration(-num) * time.Hour)
		case "day", "days":
			ruleTime = time.Now().AddDate(0, 0, -num)
		case "month", "months":
			ruleTime = time.Now().AddDate(0, -num, 0)
		case "year", "years":
			ruleTime = time.Now().AddDate(-num, 0, 0)
		default:
			return false, fmt.Errorf("unknown time unit: %s", unit)
		}
	}

	switch operator {
	case "=", "==", "is":
		return fieldValue.Equal(ruleTime), nil
	case "!=", "<>", "is not":
		return !fieldValue.Equal(ruleTime), nil
	case ">":
		return fieldValue.After(ruleTime), nil
	case ">=":
		return fieldValue.After(ruleTime) || fieldValue.Equal(ruleTime), nil
	case "<":
		return fieldValue.Before(ruleTime), nil
	case "<=":
		return fieldValue.Before(ruleTime) || fieldValue.Equal(ruleTime), nil
	default:
		return false, fmt.Errorf("unsupported time operator: %s", operator)
	}
}












































































