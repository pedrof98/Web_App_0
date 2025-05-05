package service

import (
	"context"

	"traffic-monitoring-go/internal/domain"
	"traffic-monitoring-go/internal/dto"
)

// RuleService defines operations for managing rules
type RuleService interface {
	// ListRules retrieves rules based on query parameters
	ListRules(ctx context.Context, query dto.RuleQuery) ([]domain.Rule, *dto.MetaInfo, error)

	// GetRule retrieves a single rule by ID
	GetRule(ctx context.Context, id uint) (*domain.Rule, error)

	// CreateRule creates a new rule
	CreateRule(ctx context.Context, input *dto.CreateRuleRequest, userID uint) (*domain.Rule, error)

	// UpdateRule updates an existing rule
	UpdateRule(ctx context.Context, id uint, input *dto.UpdateRulerequest) (*domain.Rule, error)

	// Deleterule removes a rule by ID
	DeleteRule(ctx context.Context, id uint) error
}

// AlertService defines operations for managing alerts
type AlertService interface {
	// placeholder
}


// SecurityEventservice defines operations for managing security events
type SecurityEventService interface {
	// placeholder
}


