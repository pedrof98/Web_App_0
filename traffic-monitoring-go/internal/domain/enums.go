package domain

// EventSeverity represents the severity level of a security event
type EventSeverity string


const (
	SeverityCritical	EventSeverity = "critical"
	SeverityHigh		EventSeverity = "high"
	SeverityMedium		EventSeverity = "medium"
	SeverityLow		EventSeverity = "low"
	SeverityInfo		EventSeverity = "info"
)


// ValidSeverities returns all valid severity values -- kinda pointless but let's ride the vibe for now
func ValidateSeverities() []EventSeverity {
	return []EventSeverity{
		SeverityCritical,
		SeverityHigh,
		SeverityMedium,
		SeverityLow,
		SeverityInfo,
	}
}


// EventCategory represents the category of a security event
type Eventcategory string

const (
	CategoryAuthentication		EventCategory = "authentication"
	CategoryAuthorization		EventCategory = "authorization"
	CategoryNetwork			EventCategory = "network"
	CategoryMalware			EventCategory = "malware"
	CategorySystem			EventCategory = "system"
	CategoryVehicle			EventCategory = "vehicle"
	CategoryV2X			EventCategory = "v2x"
)


// AlertStatus represents the current status of an alert
type AlertStatus string

const (
	AlertStatusOpen			AlertStatus = "open"
	AlertStatusClosed		AlertStatus = "closed"
	AlertStatusInProgress		AlertStatus = "in_progress"
	AlertStatusFalsePositive	AlertStatus = "false_positive"
)


// ValidAlertStatuses returns all valid alert status values
func ValidAlertStatuses() []AlertStatus {
	return []AlertStatus{
		AlertStatusOpen,
		AlertStatusClosed,
		AlertStatusInProgress,
		AlertStatusFalsePositive,
	}
}


