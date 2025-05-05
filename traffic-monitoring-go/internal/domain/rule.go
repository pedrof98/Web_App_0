package domain

import (
	"time"
)

// RuleStatus represents the status of a security rule
type RuleStatus string

const (
	RuleStatusEnabled	RuleStatus = "enabled"
	RuleStatusDisabled	RuleStatus = "disabled"
	RuleStatusTesting	RuleStatus = "testing"
)

// Rule represents a detection rule for security events
type Rule struct {
	ID		uint
	Name		string
	Description	string
	Condition	string
	Severity	EventSeverity
	Category	EventCategory
	Status		RuleStatus
	CreatedBy	uint
	CreatedAt	time.Time
	UpdatedAt	time.Time
}

//IsValid validates the rule's basic properties
func (r *Rule) IsValid() bool {
	return r.Name != "" && r.Condition != "" && r.Severity != "" && r.Category != ""
}


// CanbeDeleted determines if a rule can be safely deleted
func (r *Rule) CanBeDeleted(alertCount int64) bool {
	return alertCount == 0
}


