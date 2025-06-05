package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"traffic-monitoring-go/internal/dto"
	"traffic-monitoring-go/internal/pkg/respond"
	"traffic-monitoring-go/internal/siem/v2x"

	"github.com/gin-gonic/gin"
)

// V2XHandler handles HTTP requests for V2X-specific functionality
type V2XHandler struct {
	v2xService *v2x.V2XSIEMService
}

// NewV2XHandler creates a new V2XHandler
func NewV2XHandler(v2xService *v2x.V2XSIEMService) *V2XHandler {
	return &V2XHandler{
		v2xService: v2xService,
	}
}

// ProcessMessage handles POST /api/v1/v2x/messages
// @Summary Process a V2X message
// @Description Processes a single V2X message for threat detection
// @Tags v2x
// @Accept json
// @Produce json
// @Param message body dto.V2XMessageRequest true "V2X message data"
// @Success 200 {object} dto.Success[dto.V2XProcessingResponse]
// @Failure 400 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/v2x/messages [post]
func (h *V2XHandler) ProcessMessage(c *gin.Context) {
	var request dto.V2XMessageRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Convert hex string to bytes if provided
	var rawData []byte
	var err error

	if request.RawDataHex != "" {
		// Parse hex string
		rawData = make([]byte, len(request.RawDataHex)/2)
		for i := 0; i < len(rawData); i++ {
			val, err := strconv.ParseUint(request.RawDataHex[i*2:i*2+2], 16, 8)
			if err != nil {
				respond.BadRequest(c, fmt.Errorf("invalid hex data: %v", err))
				return
			}
			rawData[i] = byte(val)
		}
	} else if request.MessageData != nil {
		// Convert JSON message to bytes
		rawData, err = json.Marshal(request.MessageData)
		if err != nil {
			respond.BadRequest(c, fmt.Errorf("invalid message data: %v", err))
			return
		}
	} else {
		respond.BadRequest(c, fmt.Errorf("either raw_data_hex or message_data must be provided"))
		return
	}

	// Process the message
	result, err := h.v2xService.ProcessV2XMessage(c.Request.Context(), rawData)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Convert to response format
	response := &dto.V2XProcessingResponse{
		MessageID:        result.MessageID,
		ProcessingTimeMs: float64(result.ProcessingTime.Nanoseconds()) / 1000000.0,
		ThreatsDetected:  len(result.ThreatsDetected),
		AlertsCreated:    len(result.AlertsCreated),
		SecurityEventID:  result.SecurityEventID,
		Threats:          convertDetectionResults(result.ThreatsDetected),
		Errors:           result.Errors,
	}

	respond.OK(c, response, nil)
}

// ProcessBatch handles POST /api/v1/v2x/messages/batch
// @Summary Process multiple V2X messages
// @Description Processes multiple V2X messages in a batch for better performance
// @Tags v2x
// @Accept json
// @Produce json
// @Param messages body dto.V2XBatchRequest true "Batch of V2X messages"
// @Success 200 {object} dto.Success[dto.V2XBatchResponse]
// @Failure 400 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/v2x/messages/batch [post]
func (h *V2XHandler) ProcessBatch(c *gin.Context) {
	var request dto.V2XBatchRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}

	if len(request.Messages) > 1000 {
		respond.BadRequest(c, fmt.Errorf("batch size exceeds maximum of 1000 messages"))
		return
	}

	// Convert messages to raw data
	rawMessages := make([][]byte, len(request.Messages))
	for i, msg := range request.Messages {
		if msg.RawDataHex != "" {
			rawData := make([]byte, len(msg.RawDataHex)/2)
			for j := 0; j < len(rawData); j++ {
				val, err := strconv.ParseUint(msg.RawDataHex[j*2:j*2+2], 16, 8)
				if err != nil {
					respond.BadRequest(c, fmt.Errorf("invalid hex data at index %d: %v", i, err))
					return
				}
				rawData[j] = byte(val)
			}
			rawMessages[i] = rawData
		} else if msg.MessageData != nil {
			rawData, err := json.Marshal(msg.MessageData)
			if err != nil {
				respond.BadRequest(c, fmt.Errorf("invalid message data at index %d: %v", i, err))
				return
			}
			rawMessages[i] = rawData
		}
	}

	// Process the batch
	results, err := h.v2xService.ProcessV2XBatch(c.Request.Context(), rawMessages)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Convert to response format
	responses := make([]dto.V2XProcessingResponse, len(results))
	totalThreats := 0
	totalAlerts := 0

	for i, result := range results {
		if result != nil {
			responses[i] = dto.V2XProcessingResponse{
				MessageID:        result.MessageID,
				ProcessingTimeMs: float64(result.ProcessingTime.Nanoseconds()) / 1000000.0,
				ThreatsDetected:  len(result.ThreatsDetected),
				AlertsCreated:    len(result.AlertsCreated),
				SecurityEventID:  result.SecurityEventID,
				Threats:          convertDetectionResults(result.ThreatsDetected),
				Errors:           result.Errors,
			}
			totalThreats += len(result.ThreatsDetected)
			totalAlerts += len(result.AlertsCreated)
		}
	}

	batchResponse := &dto.V2XBatchResponse{
		ProcessedCount: len(results),
		TotalThreats:   totalThreats,
		TotalAlerts:    totalAlerts,
		Results:        responses,
	}

	respond.OK(c, batchResponse, nil)
}

// GetMetrics handles GET /api/v1/v2x/metrics
// @Summary Get V2X processing metrics
// @Description Returns performance and detection metrics for V2X processing
// @Tags v2x
// @Accept json
// @Produce json
// @Success 200 {object} dto.Success[dto.V2XMetricsResponse]
// @Failure 500 {object} dto.Error
// @Router /api/v1/v2x/metrics [get]
func (h *V2XHandler) GetMetrics(c *gin.Context) {
	metrics := h.v2xService.GetMetrics()
	stats := h.v2xService.GetThreatStatistics()

	response := &dto.V2XMetricsResponse{
		MessagesProcessed:       metrics.MessagesProcessed,
		ThreatsDetected:         metrics.ThreatsDetected,
		AlertsGenerated:         metrics.AlertsGenerated,
		AverageProcessingTimeMs: metrics.ProcessingLatencyMs,
		LastProcessedTime:       metrics.LastProcessedTime,
		ThreatBreakdown: dto.ThreatBreakdown{
			PositionJumpAttacks: metrics.PositionJumpAttacks,
			SpeedAnomalies:      metrics.SpeedAnomalies,
			MessageAnomalies:    metrics.MessageAnomalies,
			BoundaryViolations:  metrics.BoundaryViolations,
			ReplayAttacks:       metrics.ReplayAttacks,
		},
		ActiveVehicles: uint64(stats["active_vehicles"].(int)),
	}

	respond.OK(c, response, nil)
}

// GetConfiguration handles GET /api/v1/v2x/config
// @Summary Get V2X detection configuration
// @Description Returns current V2X threat detection configuration
// @Tags v2x
// @Accept json
// @Produce json
// @Success 200 {object} dto.Success[dto.V2XConfigResponse]
// @Failure 500 {object} dto.Error
// @Router /api/v1/v2x/config [get]
func (h *V2XHandler) GetConfiguration(c *gin.Context) {
	config := h.v2xService.GetDetectionConfig()

	response := &dto.V2XConfigResponse{
		MaxPositionJumpKm:      config.MaxPositionJumpKm,
		PositionJumpTimeWindow: config.PositionJumpTimeWindow.String(),
		MaxSpeedKmh:            config.MaxSpeedKmh,
		MaxAcceleration:        config.MaxAcceleration,
		SpeedChangeThreshold:   config.SpeedChangeThreshold,
		MinMessageInterval:     config.MinMessageInterval.String(),
		MaxMessageInterval:     config.MaxMessageInterval.String(),
		AllowedBounds:          convertGeographicBounds(config.AllowedBounds),
		MessageCountWindow:     config.MessageCountWindow,
	}

	respond.OK(c, response, nil)
}

// UpdateConfiguration handles PUT /api/v1/v2x/config
// @Summary Update V2X detection configuration
// @Description Updates V2X threat detection configuration parameters
// @Tags v2x
// @Accept json
// @Produce json
// @Param config body dto.V2XConfigRequest true "Detection configuration"
// @Success 200 {object} dto.Success[string]
// @Failure 400 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/v2x/config [put]
func (h *V2XHandler) UpdateConfiguration(c *gin.Context) {
	var request dto.V2XConfigRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Convert request to detection config
	config := h.v2xService.GetDetectionConfig()

	if request.MaxPositionJumpKm != nil {
		config.MaxPositionJumpKm = *request.MaxPositionJumpKm
	}
	if request.MaxSpeedKmh != nil {
		config.MaxSpeedKmh = *request.MaxSpeedKmh
	}
	if request.MaxAcceleration != nil {
		config.MaxAcceleration = *request.MaxAcceleration
	}
	if request.SpeedChangeThreshold != nil {
		config.SpeedChangeThreshold = *request.SpeedChangeThreshold
	}

	h.v2xService.UpdateDetectionConfig(config)

	respond.OK(c, "Configuration updated successfully", nil)
}

// GetVehicleStates handles GET /api/v1/v2x/vehicles
// @Summary Get active vehicle states
// @Description Returns current state information for all tracked vehicles
// @Tags v2x
// @Accept json
// @Produce json
// @Success 200 {object} dto.Success[[]dto.VehicleStateResponse]
// @Failure 500 {object} dto.Error
// @Router /api/v1/v2x/vehicles [get]
func (h *V2XHandler) GetVehicleStates(c *gin.Context) {
	states := h.v2xService.GetVehicleStates()

	responses := make([]dto.VehicleStateResponse, 0, len(states))
	for _, state := range states {
		responses = append(responses, dto.VehicleStateResponse{
			VehicleID:     state.VehicleID,
			LastPosition:  state.LastPosition,
			LastTimestamp: state.LastTimestamp,
			LastSpeed:     state.LastSpeed,
			LastHeading:   state.LastHeading,
			MessageCount:  state.MessageCount,
			HistorySize:   len(state.MessageHistory),
		})
	}

	respond.OK(c, responses, nil)
}

// Helper functions for response conversion

func convertDetectionResults(results []v2x.DetectionResult) []dto.ThreatResponse {
	threats := make([]dto.ThreatResponse, 0)
	for _, result := range results {
		if result.ThreatDetected {
			threats = append(threats, dto.ThreatResponse{
				Type:        result.ThreatType,
				Severity:    string(result.Severity),
				Description: result.Description,
				Confidence:  result.Confidence,
				Evidence:    result.Evidence,
			})
		}
	}
	return threats
}

func convertGeographicBounds(bounds []v2x.GeographicBound) []dto.GeographicBoundResponse {
	responses := make([]dto.GeographicBoundResponse, len(bounds))
	for i, bound := range bounds {
		responses[i] = dto.GeographicBoundResponse{
			Name:   bound.Name,
			MinLat: bound.MinLat,
			MaxLat: bound.MaxLat,
			MinLon: bound.MinLon,
			MaxLon: bound.MaxLon,
		}
	}
	return responses
}
