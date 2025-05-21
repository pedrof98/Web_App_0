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
	// ListAlerts retrieves alerts based on query parameters
	ListAlerts(ctx context.Context, query.dtoAlertQuery) ([]domain.Alert, *dto.MetaInfo, error)

	//GetAlert retrieves a single alert by ID
	GetAlert(ctx context.Context, id uint) (*domain.Alert, error)

	// CreateAlert creates a new alert
	CreateAlert(ctx context.Context, input *dto.CreateAlertRequest) (*domain.Alert, error)

	UpdateAlert(ctx context.Context, id uint, input *dto.UpdateAlertRequest) (*domain.Alert, error)

	DeleteAlert(ctx context.Context, id uint) error

	AssignAlert(ctx context.Context, id uint, userID uint) (*domain.Alert, error)
}


// SecurityEventservice defines operations for managing security events
type SecurityEventService interface {
	
	ListSecurityEvents(ctx context.Context, query dto.SecurityEventQuery) ([]domain.SecurityEvent, *dto.MetaInfo, error)

	GetSecurityEvent(ctx context.Context, id uint) (*domain.SecurityEvent, error)

	CreateSecurityEvent(ctx context.Context, input *dto.CreateSecurityEventRequest) (*domain.SecurityEvent, error)

	BatchCreateSecurityEvents(ctx context.Context, inputs []*dto.CreateSecurityEventRequest) ([]*domain.SecurityEvent, error)

	DeleteSecurityEvent(ctx context.Context, id uint) error
}


