package domain

import (
	"time"
)

// SecurityEvent represents a security-related event in the system
type SecurityEvent struct {
	ID			uint
	Timestamp		time.Time
	SourceIP		string
	SourcePort		*int
	DestinationIP		string
	DestinationPort		*int
	Protocol		string
	Action			string
	Status			string
	UserID			*uint
	DeviceID		string
	LogSourceID		uint
	Severity		EventSeverity
	Category		EventCategory
	Message			string
	RawData			string
	CreatedAt		time.Time
}

func (e *SecurityEvent) IsValid() bool {
	return e.Timestamp.Unix() > 0 && e.LogSourceID != 0 &&
		e.Severity != "" && e.Category != "" && e.Message != ""
}

func (e *SecurityEvent) ShouldTriggerAlert() bool {
	// built-in logic for critical events
	return e.Severity == SeverityCritical
}


