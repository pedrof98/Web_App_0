package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "traffic-monitoring-go/internal/dto"
    "traffic-monitoring-go/internal/pkg/respond"
    "traffic-monitoring-go/internal/service"
)

// SecurityEventHandler handles HTTP requests for security events
type SecurityEventHandler struct {
    service service.SecurityEventService
}

// NewSecurityEventHandler creates a new SecurityEventHandler
func NewSecurityEventHandler(service service.SecurityEventService) *SecurityEventHandler {
    return &SecurityEventHandler{
        service: service,
    }
}

// List handles GET /api/v1/security-events
// @Summary List security events
// @Description Get a list of security events with pagination and filtering
// @Tags security-events
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param severity query string false "Filter by severity (critical, high, medium, low, info)"
// @Param category query string false "Filter by category"
// @Param source_ip query string false "Filter by source IP"
// @Param destination_ip query string false "Filter by destination IP"
// @Param protocol query string false "Filter by protocol"
// @Param action query string false "Filter by action"
// @Param status query string false "Filter by status"
// @Param log_source_id query int false "Filter by log source ID"
// @Param device_id query string false "Filter by device ID"
// @Param from_date query string false "Filter from date (YYYY-MM-DD)"
// @Param to_date query string false "Filter to date (YYYY-MM-DD)"
// @Param search query string false "Search in message text"
// @Success 200 {object} dto.Success[[]dto.SecurityEventResponse]
// @Failure 400 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/security-events [get]
func (h *SecurityEventHandler) List(c *gin.Context) {
    // Parse query parameters
    query, err := dto.ParseSecurityEventQuery(c)
    if err != nil {
        respond.BadRequest(c, err)
        return
    }

    // Call service
    events, meta, err := h.service.ListSecurityEvents(c.Request.Context(), query)
    if err != nil {
        respond.Error(c, err)
        return
    }

    // Transform domain models to response DTOs
    responses := dto.SecurityEventsToResponses(events)

    // Send response
    respond.OK(c, responses, meta)
}

// Get handles GET /api/v1/security-events/:id
// @Summary Get a security event
// @Description Get a single security event by ID
// @Tags security-events
// @Accept json
// @Produce json
// @Param id path int true "Security Event ID"
// @Success 200 {object} dto.Success[dto.SecurityEventResponse]
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/security-events/{id} [get]
// Get handles GET /api/v1/security-events/:id
func (h *SecurityEventHandler) Get(c *gin.Context) {
    // Parse event ID from path
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        respond.BadRequest(c, err)
        return
    }

    // Call service
    event, err := h.service.GetSecurityEvent(c.Request.Context(), uint(id))
    if err != nil {
        respond.Error(c, err)
        return
    }

    // Transform domain model to response DTO
    response := dto.SecurityEventToResponse(event)

    // Send response
    respond.OK(c, response, nil)
}

// Create handles POST /api/v1/security-events
// @Summary Create a security event
// @Description Create a new security event
// @Tags security-events
// @Accept json
// @Produce json
// @Param event body dto.CreateSecurityEventRequest true "Security Event details"
// @Success 201 {object} dto.Success[dto.SecurityEventResponse]
// @Failure 400 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/security-events [post]
func (h *SecurityEventHandler) Create(c *gin.Context) {
    // Parse request body
    var request dto.CreateSecurityEventRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        respond.BadRequest(c, err)
        return
    }

    // Call service
    event, err := h.service.CreateSecurityEvent(c.Request.Context(), &request)
    if err != nil {
        respond.Error(c, err)
        return
    }

    // Transform domain model to response DTO
    response := dto.SecurityEventToResponse(event)

    // Send response
    respond.Created(c, response)
}

// BatchCreate handles POST /api/v1/security-events/batch
// @Summary Create multiple security events
// @Description Create multiple security events in a single request
// @Tags security-events
// @Accept json
// @Produce json
// @Param events body []dto.CreateSecurityEventRequest true "Security Event details"
// @Success 201 {object} dto.Success[[]dto.SecurityEventResponse]
// @Failure 400 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/security-events/batch [post]
func (h *SecurityEventHandler) BatchCreate(c *gin.Context) {
    // Parse request body
    var requests []*dto.CreateSecurityEventRequest
    if err := c.ShouldBindJSON(&requests); err != nil {
        respond.BadRequest(c, err)
        return
    }

    // Call service
    events, err := h.service.BatchCreateSecurityEvents(c.Request.Context(), requests)
    if err != nil {
        respond.Error(c, err)
        return
    }

    // Transform domain models to response DTOs
    responses := make([]dto.SecurityEventResponse, len(events))
    for i, event := range events {
        responses[i] = dto.SecurityEventToResponse(event)
    }

    // Send response
    respond.Created(c, responses)
}

// Delete handles DELETE /api/v1/security-events/:id
// @Summary Delete a security event
// @Description Delete an existing security event
// @Tags security-events
// @Accept json
// @Produce json
// @Param id path int true "Security Event ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/security-events/{id} [delete]
func (h *SecurityEventHandler) Delete(c *gin.Context) {
    // Parse event ID from path
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        respond.BadRequest(c, err)
        return
    }

    // Call service
    err = h.service.DeleteSecurityEvent(c.Request.Context(), uint(id))
    if err != nil {
        respond.Error(c, err)
        return
    }

    // Send response
    respond.NoContent(c)
}