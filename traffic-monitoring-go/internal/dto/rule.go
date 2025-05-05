package dto

import (
	"time"

	"github.com/gin-gonic/gin"
	"traffic-monitoring-go/internal/domain"
)

// RuleQuery represents query parameters for filtering rules
type RuleQuery struct {
	PaginationQuery
	Status		string	`form:"status" binding:"omitempty,oneof=enabled disabled testing"`
	Category	string	`form:"category" binding:"omitempty"`
	Search		string	`form:"search" binding:"omitempty"`
}


// ParseRuleQuery parses query parameters from the request
func ParseRuleQuery(c *gin.Context) (RuleQuery, error) {
	var q RuleQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		return q, err
	}

	// set defaults if not provided
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 50
	}

	return q, nil
}


// Ruleresponse represents a rule in API responses
type RuleResponse struct {
	ID		uint			`json:"id"`
	Name		string			`json:"name"`
	Description	string			`json:"description,omitempty"`
	Condition	string			`json:"condition"`
	Severity	domain.EventSeverity	`json:"severity"`
	Category	domain.EventCategory	`json:"category"`
	Status		domain.RuleStatus	`json:"status"`
	CreatedBy	uint			`json:"created_by"`
	CreatedAt	time.Time		`json:"created_at"`
	UpdatedAt	time.Time		`json:"updated_at"`
}


// CreateRuleRequest represents the request to create a new rule
type CreateRuleRequest struct {
	Name		string			`json:"name" binding:"required,min=3,max=100"`
	Description	string			`json:"description" binding:"max=500"`
	Condition	string			`json:"condition" binding:"required"`
	Severity	domain.EventSeverity	`json:"severity" binding:"required,oneof=critical high medium low info"`
	Category	domain.EventCategory	`json:"category" binding:"required"`
	Status		domain.RuleStatus	`json:"status" binding:"omitempty,oneof=enabled disabled testing"`
}


// UpdateRulerequest represents the request to update an existing rule
type UpdateRuleRequest struct {
	Name		*string			`json:"name" binding:"omitempty,min=3,max=100"`
	Description	*string			`json:"description" binding:"omitempty,max=500"`
	Condition	*string			`json:"condition" binding:"omitempty"`
	Severity	*domain.EventSeverity	`json:"severity" binding:"omitempty,oneof=critical high medium low info"`
	Category	*domain.EventCategory	`json:"category" binding:"omitempty"`
	Status		*domain.RuleStatus	`json:"status" binding:"omitempty,oneof=enabled disabled testing"`
}


// ToDomain converts the create request to a domain rule
func (r *CreateRuleRequest) ToDomain(userID uint) *domain.Rule {
	status := r.Status
	if status == "" {
		status = domain.RuleStatusDisabled
	}

	return &domain.Rule{
		Name:		r.Name,
		Description	r.Description,
		Condition	r.Condition,
		Severity	r.Severity,
		Category	r.Category,
		Status:		status,
		CreatedBy:	userID,
	}
}


// ApplyToRule applies update request fields to a domain rule
func (r *UpdateRuleRequest) ApplyToRule(rule *domain.Rule) {
	if r.Name != nil {
		rule.Name = *r.Name
	}
	if r.Description != nil {
		rule.Description = *r.Description
	}
	if r.Condition != nil {
		rule.Condition = *r.Condition
	}
	if r.Severity != nil {
		rule.Severity = *r.Severity
	}
	if r.Category != nil {
		rule.Category = *r.Category
	}
	if r.Status != nil {
		rule.Status = *r.Status
	}
}

// FromDomain converts a domain rule to a response DTO
func RuleToResponse(rule *domain.Rule) RuleResponse {
	return RuleResponse{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Condition:   rule.Condition,
		Severity:    rule.Severity,
		Category:    rule.Category,
		Status:      rule.Status,
		CreatedBy:   rule.CreatedBy,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}
}

// RulesToResponses converts a slice of a domain rules to response DTOs
func RulesToResponses(rules []domain.Rule) []RuleResponse {
	responses := make([]RuleResponse, len(rules))
	for i, rule := range rules {
		responses[i] = RuleToResponse(&rule)
	}
	return responses
}
