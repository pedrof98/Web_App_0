package dto

import (
	"time"

	"github.com/gin-gonic/gin"
	"traffic-monitoring-go/internal/domain"
)

// AlertQuery represents query parameters for filtering alerts
type AlertQuery struct {
	PaginationQuery
	Status     string `form:"status" binding:"omitempty,oneof=open closed in_progress false_positive"`
	Severity   string `form:"severity" binding:"omitempty,oneof=critical high medium low info"`
	AssignedTo *uint  `form:"assigned_to" binding:"omitempty"`
	RuleID     *uint  `form:"rule_id" binding:"omitempty"`
	FromDate   string `form:"from_date" binding:"omitempty,datetime=2006-01-02"`
	ToDate     string `form:"to_date" binding:"omitempty,datetime=2006-01-02"`
	Search     string `form:"search" binding:"omitempty"`
}

// ParseAlertQuery parses query parameters from the request
func ParseAlertQuery(c *gin.Context) (AlertQuery, error) {
	var q AlertQuery
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


// AlertResponse represents an alert in API responses
type AlertResponse struct {
	ID              uint               `json:"id"`
	RuleID          uint               `json:"rule_id"`
	Rule            *RuleResponse      `json:"rule,omitempty"`
	SecurityEventID uint               `json:"security_event_id"`
	Timestamp       time.Time          `json:"timestamp"`
	Severity        domain.EventSeverity `json:"severity"`
	Status          domain.AlertStatus   `json:"status"`
	AssignedTo      *uint              `json:"assigned_to,omitempty"`
	Resolution      string             `json:"resolution,omitempty"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
}

// CreateAlertRequest represents the request to create a new alert
type CreateAlertRequest struct {
	RuleID          uint               `json:"rule_id" binding:"required"`
	SecurityEventID uint               `json:"security_event_id" binding:"required"`
	Severity        domain.EventSeverity `json:"severity" binding:"required,oneof=critical high medium low info"`
	Status          domain.AlertStatus   `json:"status" binding:"omitempty,oneof=open in_progress closed false_positive"`
}

// UpdateAlertRequest represents the request to update an existing alert
type UpdateAlertRequest struct {
	Status      *domain.AlertStatus  `json:"status" binding:"omitempty,oneof=open in_progress closed false_positive"`
	AssignedTo  *uint              `json:"assigned_to" binding:"omitempty"`
	Resolution  *string            `json:"resolution" binding:"omitempty"`
}


// ToDomain converts the create request to a domain alert
func (r *CreateAlertRequest) ToDomain() *domain.Alert {
	status := r.Status
	if status == "" {
		status = domain.AlertStatusOpen
	}

	return &domain.Alert{
		RuleID:			r.RuleID,
		SecurityEventID:	r.SecurityEventID,
		Timestamp:		time.Now(),
		Severity:		r.Severity,
		Status:			status,
	}
}

// ApplyToAlert applies update request fields to a domain alert
func (r *UpdateAlertRequest) ApplyToAlert(alert *domain.Alert) {
	if r.Stauts != nil {
		alert.Status = *r.Status
	}
	if r.AssignedTo != nil {
		alert.AssignedTo = r.AssignedTo
	}
	if r.Resolution != nil {
		alert.Resolution = *r.Resolution
	}
}

// AlertToResponse converts a domain alert to a response DTO
func AlertToResponse(alert *domain.Alert) AlertResponse {
	response := AlertResponse{
		ID:			alert.ID,
		RuleID:			alert.RuleID,
		SecurityEventID:	alert.SecurityEventID,
		Timestamp:		alert.Timestamp,
		Severity:		alert.Severity,
		Status:			alert.Status,
		AssignedTo:		alert.AssignedTo,
		Resolution:		alert.Resolution,
		CreatedAt:		alert.CreatedAt,
		UpdatedAt:		alert.UpdatedAt,
	}

	if alert.Rule != nil {
		ruleResponse := RuleToResponse(alert.Rule)
		response.Rule = &ruleResponse
	}

	return response
}

// AlertsToResponses converts a slice of domain alerts to response DTOs
func AlertsToResponses(alerts []domain.Alert) []AlertResponse {
	responses := make([]AlertResponse, len(alerts))
	for i, alert := range alerts {
		responses[i] = AlertToResponse(&alert)
	}
	return responses
}


