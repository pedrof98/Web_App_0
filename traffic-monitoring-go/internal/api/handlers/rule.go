package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"traffic-monitoring-go/internal/dto"
	"traffic-monitoring-go/internal/pkg/respond"
	"traffic-monitoring-go/internal/service"
)

// RuleHandler handles HTTP requests for rules
type RuleHandler struct {
	service service.RuleService
}

// NewRuleHandler creates a new RuleHandler
func NewRuleHandler(service service.RuleService) *RuleHandler {
	return &RuleHandler{
		service: service,
	}
}

// List handles GET /api/v1/rules
// @Summary List rules
// @Description Get a list of rules with pagination and filtering
// @Tags rules
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query string false "Filter by status (enabled, disabled, testing)"
// @Param category query string false "Filter by category"
// @Param search query string false "Search in name and description"
// @Success 200 {object} dto.Success[[]dto.RuleResponse]
// @Failure 400 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/rules [get]

func (h *RuleHandler) List(c *gin.Context) {
	// Parse query parameters
	query, err := dto.ParseRuleQuery(c)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// call service
	rules, meta, err := h.service.ListRules(c.Request.Context(), query)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// transform domain models to response DTOs
	responses := dto.RulesToResponses(rules)

	// send response
	respond.OK(c, responses, meta)
}

// Get handles GET /api/v1/rules/:id
// @Summary Get a rule
// @Description Get a single rule by ID
// @Tags rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Success 200 {object} dto.Success[dto.RuleResponse]
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/rules/{id} [get]

func (h *RuleHandler) Get(c *gin.Context) {
	// Parse rule ID from path 
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// transform domain model to response DTO
	response := dto.RuleToResponse(rule)

	// send response
	respond.OK(c, reponse, nil)
}


// Create handles POST /api/v1/rules
// @Summary Create a rule
// @Description Create a new rule
// @Tags rules
// @Accept json
// @Produce json
// @Param rule body dto.CreateRuleRequest true "Rule details"
// @Success 201 {object} dto.Success[dto.RuleResponse]
// @Failure 400 {object} dto.Error
// @Failure 409 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/rules [post]


func (h *RuleHandler) Create(c *gin.Context) {
	// parse request body
	var request dto.CreateRuleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}


	// TODO: Get user ID from context after auth middleware is implemented
	userID := uint(1) // temporary hardcoded value

	// call service
	rule, err := h.service.CreateRule(c.Request.Context(), &request, userID)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// transform domain model to response DTO
	response := dto.RuleToResponse(rule)

	// send response
	respond.Created(c, response)
}

// Update handles PUT /api/v1/rules/:id
// @Summary Update a rule
// @Description Update an existing rule
// @Tags rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Param rule body dto.UpdateRuleRequest true "Rule details"
// @Success 200 {object} dto.Success[dto.RuleResponse]
// @Failure 400 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 409 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/rules/{id} [put]

func (h *RuleHandler) Update(c *gin.Context) {
	// parse rule ID from path
	id, err := strconv.ParseUint(c.Para,("id"), 10, 32)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// parse request body
	var request dto.UpdateRuleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}

	// call service
	rule, err := h.service.UpdateRule(c.Request.Context(), uint(id), &request)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// transform domain model to response DTO
	response := dto.RuleToResponse(rule)

	// send response
	respond.OK(c, response, nil)
}

// Delete handles DELETE /api/v1/rules/:id
// @Summary Delete a rule
// @Description Delete an existing rule
// @Tags rules
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/rules/{id} [delete]

func (h *RuleHandler) Delete(c *gin.Context) {
	// parse rule ID from path
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// call service
	err = h.service.DeleteRule(c.Request.Context(), uint(id))
	if err != nil {
		respond.Error(c, err)
		return
	}

	// send response
	respond.NoContet(c)
}





