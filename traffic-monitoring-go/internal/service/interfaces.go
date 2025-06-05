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
	UpdateRule(ctx context.Context, id uint, input *dto.UpdateRuleRequest) (*domain.Rule, error)

	// Deleterule removes a rule by ID
	DeleteRule(ctx context.Context, id uint) error
}

// AlertService defines operations for managing alerts
type AlertService interface {
	// ListAlerts retrieves alerts based on query parameters
	ListAlerts(ctx context.Context, query dto.AlertQuery) ([]domain.Alert, *dto.MetaInfo, error)

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

// AuthService defines operations for authentication
type AuthService interface {
	// Login authenticates a user and returns a JWT token
	Login(ctx context.Context, request *dto.LoginRequest) (*dto.LoginResponse, error)

	// Register creates a new user account
	Register(ctx context.Context, request *dto.RegisterRequest) (*domain.User, error)

	// RefreshToken generates a new access token
	RefreshToken(ctx context.Context, tokenString string) (*dto.LoginResponse, error)

	// GetCurrentUser retrieves the current user from context
	GetCurrentUser(ctx context.Context) (*domain.User, error)

	// ChangePassword changes a user's password
	ChangePassword(ctx context.Context, userID uint, request *dto.ChangePasswordRequest) error

	// UpdateProfile updates a user's profile information
	UpdateProfile(ctx context.Context, userID uint, request *dto.UpdateProfileRequest) (*domain.User, error)
}

// UserService defines operations for user management
type UserService interface {
	// ListUsers retrieves users based on query parameters
	ListUsers(ctx context.Context, query dto.UserQuery) ([]domain.User, *dto.MetaInfo, error)

	// GetUser retrieves a single user by ID
	GetUser(ctx context.Context, id uint) (*domain.User, error)

	// CreateUser creates a new user (admin only)
	CreateUser(ctx context.Context, request *dto.RegisterRequest) (*domain.User, error)

	// UpdateUser updates an existing user (admin only)
	UpdateUser(ctx context.Context, id uint, request *dto.UpdateProfileRequest) (*domain.User, error)

	// DeleteUser removes a user by ID (admin only)
	DeleteUser(ctx context.Context, id uint) error

	// ActivateUser activates a user account
	ActivateUser(ctx context.Context, id uint) error

	// DeactivateUser deactivates a user account
	DeactivateUser(ctx context.Context, id uint) error
}
