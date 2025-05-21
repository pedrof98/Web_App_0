package dto

import (
    "time"

    "github.com/gin-gonic/gin"
    "traffic-monitoring-go/internal/domain"
)

// SecurityEventQuery represents query parameters for filtering security events
type SecurityEventQuery struct {
    PaginationQuery
    Severity        string  `form:"severity" binding:"omitempty,oneof=critical high medium low info"`
    Category        string  `form:"category" binding:"omitempty"`
    SourceIP        string  `form:"source_ip" binding:"omitempty"`
    DestinationIP   string  `form:"destination_ip" binding:"omitempty"`
    Protocol        string  `form:"protocol" binding:"omitempty"`
    Action          string  `form:"action" binding:"omitempty"`
    Status          string  `form:"status" binding:"omitempty"`
    LogSourceID     *uint   `form:"log_source_id" binding:"omitempty"`
    DeviceID        string  `form:"device_id" binding:"omitempty"`
    FromDate        string  `form:"from_date" binding:"omitempty,datetime=2006-01-02"`
    ToDate          string  `form:"to_date" binding:"omitempty,datetime=2006-01-02"`
    Search          string  `form:"search" binding:"omitempty"`
}

// ParseSecurityEventQuery parses query parameters from the request
func ParseSecurityEventQuery(c *gin.Context) (SecurityEventQuery, error) {
    var q SecurityEventQuery
    if err := c.ShouldBindQuery(&q); err != nil {
        return q, err
    }

    // Set defaults if not provided
    if q.Page == 0 {
        q.Page = 1
    }
    if q.PageSize == 0 {
        q.PageSize = 50
    }

    return q, nil
}

// SecurityEventResponse represents a security event in API responses
type SecurityEventResponse struct {
    ID              uint                 `json:"id"`
    Timestamp       time.Time            `json:"timestamp"`
    SourceIP        string               `json:"source_ip,omitempty"`
    SourcePort      *int                 `json:"source_port,omitempty"`
    DestinationIP   string               `json:"destination_ip,omitempty"`
    DestinationPort *int                 `json:"destination_port,omitempty"`
    Protocol        string               `json:"protocol,omitempty"`
    Action          string               `json:"action,omitempty"`
    Status          string               `json:"status,omitempty"`
    UserID          *uint                `json:"user_id,omitempty"`
    DeviceID        string               `json:"device_id,omitempty"`
    LogSourceID     uint                 `json:"log_source_id"`
    Severity        domain.EventSeverity `json:"severity"`
    Category        domain.EventCategory `json:"category"`
    Message         string               `json:"message"`
    CreatedAt       time.Time            `json:"created_at"`
}

// CreateSecurityEventRequest represents the request to create a new security event
type CreateSecurityEventRequest struct {
    Timestamp       time.Time            `json:"timestamp" binding:"required"`
    SourceIP        string               `json:"source_ip" binding:"omitempty,ip"`
    SourcePort      *int                 `json:"source_port" binding:"omitempty,min=1,max=65535"`
    DestinationIP   string               `json:"destination_ip" binding:"omitempty,ip"`
    DestinationPort *int                 `json:"destination_port" binding:"omitempty,min=1,max=65535"`
    Protocol        string               `json:"protocol" binding:"omitempty"`
    Action          string               `json:"action" binding:"omitempty"`
    Status          string               `json:"status" binding:"omitempty"`
    UserID          *uint                `json:"user_id" binding:"omitempty"`
    DeviceID        string               `json:"device_id" binding:"omitempty"`
    LogSourceID     uint                 `json:"log_source_id" binding:"required"`
    Severity        domain.EventSeverity `json:"severity" binding:"required,oneof=critical high medium low info"`
    Category        domain.EventCategory `json:"category" binding:"required"`
    Message         string               `json:"message" binding:"required"`
    RawData         string               `json:"raw_data" binding:"omitempty"`
}

// ToDomain converts the create request to a domain security event
func (r *CreateSecurityEventRequest) ToDomain() *domain.SecurityEvent {
    return &domain.SecurityEvent{
        Timestamp:       r.Timestamp,
        SourceIP:        r.SourceIP,
        SourcePort:      r.SourcePort,
        DestinationIP:   r.DestinationIP,
        DestinationPort: r.DestinationPort,
        Protocol:        r.Protocol,
        Action:          r.Action,
        Status:          r.Status,
        UserID:          r.UserID,
        DeviceID:        r.DeviceID,
        LogSourceID:     r.LogSourceID,
        Severity:        r.Severity,
        Category:        r.Category,
        Message:         r.Message,
        RawData:         r.RawData,
    }
}

// SecurityEventToResponse converts a domain security event to a response DTO
func SecurityEventToResponse(event *domain.SecurityEvent) SecurityEventResponse {
    return SecurityEventResponse{
        ID:              event.ID,
        Timestamp:       event.Timestamp,
        SourceIP:        event.SourceIP,
        SourcePort:      event.SourcePort,
        DestinationIP:   event.DestinationIP,
        DestinationPort: event.DestinationPort,
        Protocol:        event.Protocol,
        Action:          event.Action,
        Status:          event.Status,
        UserID:          event.UserID,
        DeviceID:        event.DeviceID,
        LogSourceID:     event.LogSourceID,
        Severity:        event.Severity,
        Category:        event.Category,
        Message:         event.Message,
        CreatedAt:       event.CreatedAt,
    }
}

// SecurityEventsToResponses converts a slice of domain security events to response DTOs
func SecurityEventsToResponses(events []domain.SecurityEvent) []SecurityEventResponse {
    responses := make([]SecurityEventResponse, len(events))
    for i, event := range events {
        responses[i] = SecurityEventToResponse(&event)
    }
    return responses
}