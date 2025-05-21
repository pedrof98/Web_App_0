package repository

import (
	"context"

	"traffic-monitoring-go/internal/domain"
	"traffic-monitoring-go/internal/dto"
)


// Rulerepository defines the interface for rule data operations
type RuleRepository interface {
	// Findrules retrieves rules based on query parameters
	FindRules(ctx context.Context, query dto.RuleQuery) ([]domain.Rule, int64, error)

	// GetRuleByID retrieves a single rule by ID
	GetRuleByID(ctx context.Context, id uint) (*domain.Rule, error)

	// CreateRule saves a new rule
	CreateRule(ctx context.Context, rule *domain.Rule) error

	// UpdateRule updates an existing rule
	UpdateRule(ctx context.Context, rule *domain.Rule) error

	// DeleteRule removes a rule by ID
	DeleteRule(ctx context.Context, id uint) error

	// CountAlertsByRuleID counts alerts associated with a rule
	CountAlertsByRuleID(ctx context.Context, ruleID uint) (int64, error)
}

// AlertRepository defines the interface for alert data operations
type AlertRepository interface {
	// FindAlerts retrieves alerts based on query parameters
	FindAlerts(ctx context.Context, query dto.AlertQuery) ) []domain.Alert, int64, error)

	// GetAlertyID retrieves a single alert by ID
	GetAlertByID(ctx context.Context, id uint) (*domain.Alert, error)

	// CreateAlert saves a new alert
	CreateAlert(ctx context.Context, alert *domain.Alert) error

	// UpdateAlert updates an existing alert
	UpdateAlert(ctx context.Context, alert *domain.Alert) error

	// DeleteAlert removes an alert by ID
	DeleteAlert(ctx context.Context, id uint) error

	// CountAlertByRuleID counts alerts associated with a rule
	CountAlertsByRuleID(ctx context.Context, ruleID uint) (int64, error)
}

// SecurityEventRepository defines the interface for security event data operations
type SecurityEventRepository interface {
	
	FindSecurityEvents(ctx context.Context, query dto.SecurityEventQuery) ([]domain.SecurityEvent, int64, error)

	GetSecurityEventByID(ctx context.Context, id uint) (*domain.SecurityEvent, error)

	CreateSecurityEvent(ctx context.Context, event *domain.SecurityEvent) error

	BatchCreateSecurityEvents(ctx context.Context, events []*domain.SecurityEvent) error

	DeleteSecurityEvent(ctx context.Context, id uint) error 
}


