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
	// Interface methods placeholder for now
}

// SecurityEventRepository defines the interface for security event data operations
type SecurityEventRepository interface {
	// placeholder for now
}


