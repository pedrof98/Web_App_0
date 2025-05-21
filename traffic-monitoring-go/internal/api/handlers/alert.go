package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"traffic-monitoring-go/internal/dto"
	"traffic-monitoring-go/internal/pkg/respond"
	"traffic-monitoring-go/internal/service"
)

// AlertHandler handles HTTP requests for alerts
type AlertHandler struct {
	service service.AlertService
}

// NewAlertHandler creates a new AlertHandler
func NewAlertHandler(service service.AlertService) *AlertHandler {
	return &AlertHandler{
		service: service,
	}
}

// List handles GET /api/v1/alerts
// @Summary List alerts
// @Description Get a list of alerts with pagination and filtering
// @Tags alerts
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Filter by status (open, closed, in_progress, false_positive)"
// @Param severity query string false "Filter by severity (critical, high, medium, low, info)"
// @Param assigned_to query int false "Filter by assigned user ID"
// @Param rule_id query int false "Filter by rule ID"
// @Param from_date query string false "Filter from date (YYYY-MM-DD)"
// @Param to_date query string false "Filter to date (YYYY-MM-DD)"
// @Param search query string false "Search in resolution text"
// @Success 200 {object} dto.Success[[]dto.AlertResponse]
// @Failure 400 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/alerts [get]
func (h *AlertHandler) List(c *gin.Context) {
	// Parse query parameters
	query, err := dto.ParseAlertQuery(c)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	alerts, meta, err := h.service.ListAlerts(c.Request.Context(), query)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Transform domain models to response DTOs
	responses := dto.AlertsToResponses(alerts)

	// Send response
	respond.OK(c, responses, meta)
}

// Get handles GET /api/v1/alerts/:id
// @Summary Get an alert
// @Description Get a single alert by ID
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Success 200 {object} dto.Success[dto.AlertResponse]
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/alerts/{id} [get]
func (h *AlertHandler) Get(c *gin.Context) {
	// Parse alert ID from path
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	alert, err := h.service.GetAlert(c.Request.Context(), uint(id))
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Transform domain model to response DTO
	response := dto.AlertToResponse(alert)

	// Send response
	respond.OK(c, response, nil)
}

// Create handles POST /api/v1/alerts
// @Summary Create an alert
// @Description Create a new alert
// @Tags alerts
// @Accept json
// @Produce json
// @Param alert body dto.CreateAlertRequest true "Alert details"
// @Success 201 {object} dto.Success[dto.AlertResponse]
// @Failure 400 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/alerts [post]
func (h *AlertHandler) Create(c *gin.Context) {
	// Parse request body
	var request dto.CreateAlertRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	alert, err := h.service.CreateAlert(c.Request.Context(), &request)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Transform domain model to response DTO
	response := dto.AlertToResponse(alert)

	// Send response
	respond.Created(c, response)
}

// Update handles PUT /api/v1/alerts/:id
// @Summary Update an alert
// @Description Update an existing alert
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Param alert body dto.UpdateAlertRequest true "Alert details"
// @Success 200 {object} dto.Success[dto.AlertResponse]
// @Failure 400 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/alerts/{id} [put]
func (h *AlertHandler) Update(c *gin.Context) {
	// Parse alert ID from path
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Parse request body
	var request dto.UpdateAlertRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	alert, err := h.service.UpdateAlert(c.Request.Context(), uint(id), &request)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Transform domain model to response DTO
	response := dto.AlertToResponse(alert)

	// Send response
	respond.OK(c, response, nil)
}

// Delete handles DELETE /api/v1/alerts/:id
// @Summary Delete an alert
// @Description Delete an existing alert
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/alerts/{id} [delete]
func (h *AlertHandler) Delete(c *gin.Context) {
	// Parse alert ID from path
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	err = h.service.DeleteAlert(c.Request.Context(), uint(id))
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Send response
	respond.NoContent(c)
}

// Assign handles POST /api/v1/alerts/:id/assign
// @Summary Assign an alert
// @Description Assign an alert to a user
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path int true "Alert ID"
// @Param user_id body int true "User ID"
// @Success 200 {object} dto.Success[dto.AlertResponse]
// @Failure 400 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/alerts/{id}/assign [post]

// Assign handles POST /api/v1/alerts/:id/assign
func (h *AlertHandler) Assign(c *gin.Context) {
    // Parse alert ID from path
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        respond.BadRequest(c, err)
        return
    }

    // Parse user ID from request body
    var request struct {
        UserID uint `json:"user_id" binding:"required"`
    }
    if err := c.ShouldBindJSON(&request); err != nil {
        respond.BadRequest(c, err)
        return
    }

    // Call service
    alert, err := h.service.AssignAlert(c.Request.Context(), uint(id), request.UserID)
    if err != nil {
        respond.Error(c, err)
        return
    }

    // Transform domain model to response DTO
    response := dto.AlertToResponse(alert)

    // Send response
    respond.OK(c, response, nil)
}