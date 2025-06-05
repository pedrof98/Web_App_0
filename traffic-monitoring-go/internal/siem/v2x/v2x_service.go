package v2x

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"traffic-monitoring-go/internal/domain"
	"traffic-monitoring-go/internal/dto"
	"traffic-monitoring-go/internal/service"

	"github.com/sirupsen/logrus"
)

// V2XSIEMService integrates V2X threat detection with the main SIEM system
type V2XSIEMService struct {
	detectionEngine      *V2XDetectionEngine
	messageParser        *MessageParser
	securityEventService service.SecurityEventService
	alertService         service.AlertService
	logger               *logrus.Logger

	// Performance metrics
	metrics *V2XMetrics
}

// V2XMetrics tracks performance and detection statistics
type V2XMetrics struct {
	MessagesProcessed   uint64
	ThreatsDetected     uint64
	AlertsGenerated     uint64
	ProcessingLatencyMs float64
	LastProcessedTime   time.Time

	// Threat type counters
	PositionJumpAttacks uint64
	SpeedAnomalies      uint64
	MessageAnomalies    uint64
	BoundaryViolations  uint64
	ReplayAttacks       uint64
}

// V2XProcessingResult contains the outcome of processing a V2X message
type V2XProcessingResult struct {
	MessageID       string
	ProcessingTime  time.Duration
	ThreatsDetected []DetectionResult
	AlertsCreated   []uint // Alert IDs
	SecurityEventID uint
	Errors          []string
}

func NewV2XSIEMService(
	securityEventService service.SecurityEventService,
	alertService service.AlertService,
	logger *logrus.Logger,
) *V2XSIEMService {
	return &V2XSIEMService{
		detectionEngine:      NewV2XDetectionEngine(),
		messageParser:        NewMessageParser(),
		securityEventService: securityEventService,
		alertService:         alertService,
		logger:               logger,
		metrics:              &V2XMetrics{},
	}
}

// ProcessV2XMessage is the main entry point for processing V2X messages
func (s *V2XSIEMService) ProcessV2XMessage(ctx context.Context, rawData []byte) (*V2XProcessingResult, error) {
	startTime := time.Now()

	result := &V2XProcessingResult{
		MessageID:       fmt.Sprintf("v2x_%d", time.Now().UnixNano()),
		ThreatsDetected: make([]DetectionResult, 0),
		AlertsCreated:   make([]uint, 0),
		Errors:          make([]string, 0),
	}

	// Step 1: Parse the V2X message
	parsedMsg, err := s.messageParser.ParseMessage(rawData)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Message parsing failed: %v", err))
		return result, err
	}

	if !parsedMsg.Valid {
		result.Errors = append(result.Errors, "Invalid message format")
		return result, fmt.Errorf("invalid message format")
	}

	// Step 2: Convert to detection engine format
	v2xMsg := parsedMsg.ToV2XMessage()
	if v2xMsg == nil {
		result.Errors = append(result.Errors, "Failed to convert message format")
		return result, fmt.Errorf("message conversion failed")
	}

	result.MessageID = v2xMsg.ID

	// Step 3: Create security event for the message
	securityEventID, err := s.createSecurityEvent(ctx, v2xMsg, parsedMsg)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Security event creation failed: %v", err))
		s.logger.WithError(err).Error("Failed to create security event for V2X message")
		// Continue processing even if event creation fails
	} else {
		result.SecurityEventID = securityEventID
	}

	// Step 4: Run threat detection
	threats := s.detectionEngine.ProcessMessage(ctx, v2xMsg)
	result.ThreatsDetected = threats

	// Step 5: Create alerts for detected threats
	for _, threat := range threats {
		if threat.ThreatDetected {
			alertID, err := s.createAlert(ctx, threat, securityEventID)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Alert creation failed: %v", err))
				s.logger.WithError(err).WithField("threat_type", threat.ThreatType).Error("Failed to create alert")
			} else {
				result.AlertsCreated = append(result.AlertsCreated, alertID)
			}
		}
	}

	// Step 6: Update metrics
	result.ProcessingTime = time.Since(startTime)
	s.updateMetrics(result)

	s.logger.WithFields(logrus.Fields{
		"message_id":       result.MessageID,
		"vehicle_id":       v2xMsg.VehicleID,
		"processing_time":  result.ProcessingTime,
		"threats_detected": len(result.ThreatsDetected),
		"alerts_created":   len(result.AlertsCreated),
	}).Info("V2X message processed")

	return result, nil
}

// ProcessV2XBatch processes multiple V2X messages in a batch for better performance
func (s *V2XSIEMService) ProcessV2XBatch(ctx context.Context, rawMessages [][]byte) ([]*V2XProcessingResult, error) {
	results := make([]*V2XProcessingResult, len(rawMessages))

	for i, rawData := range rawMessages {
		result, err := s.ProcessV2XMessage(ctx, rawData)
		if err != nil {
			// Continue processing other messages even if one fails
			s.logger.WithError(err).WithField("batch_index", i).Error("Failed to process message in batch")
		}
		results[i] = result
	}

	return results, nil
}

// createSecurityEvent creates a security event from the V2X message
func (s *V2XSIEMService) createSecurityEvent(ctx context.Context, v2xMsg *V2XMessage, parsedMsg *ParsedMessage) (uint, error) {
	// Convert V2X message to security event format
	eventData := map[string]interface{}{
		"vehicle_id":    v2xMsg.VehicleID,
		"message_type":  string(v2xMsg.MessageType),
		"position":      v2xMsg.Position,
		"speed":         v2xMsg.Speed,
		"heading":       v2xMsg.Heading,
		"message_count": v2xMsg.MessageCount,
	}

	if parsedMsg.BSMData != nil {
		eventData["bsm_data"] = parsedMsg.BSMData
	}

	if parsedMsg.CAMData != nil {
		eventData["cam_data"] = parsedMsg.CAMData
	}

	eventDataJSON, _ := json.Marshal(eventData)

	request := &dto.CreateSecurityEventRequest{
		Timestamp:   v2xMsg.Timestamp,
		DeviceID:    v2xMsg.VehicleID,
		LogSourceID: 1,                   // V2X log source - would be configured in system
		Severity:    domain.SeverityInfo, // Default severity, may be updated by threat detection
		Category:    domain.CategoryV2X,
		Message:     fmt.Sprintf("V2X %s message from vehicle %s", v2xMsg.MessageType, v2xMsg.VehicleID),
		RawData:     string(eventDataJSON),
	}

	securityEvent, err := s.securityEventService.CreateSecurityEvent(ctx, request)
	if err != nil {
		return 0, err
	}

	return securityEvent.ID, nil
}

// createAlert creates an alert for a detected threat
func (s *V2XSIEMService) createAlert(ctx context.Context, threat DetectionResult, securityEventID uint) (uint, error) {
	// Convert threat data to alert format
	evidenceJSON, _ := json.Marshal(threat.Evidence)

	request := &dto.CreateAlertRequest{
		RuleID:          s.getThreatRuleID(threat.ThreatType), // Map threat type to rule ID
		SecurityEventID: securityEventID,
		Severity:        threat.Severity,
		Status:          domain.AlertStatusOpen,
	}

	alert, err := s.alertService.CreateAlert(ctx, request)
	if err != nil {
		return 0, err
	}

	// Update the alert with threat-specific information
	updateReq := &dto.UpdateAlertRequest{
		Resolution: &threat.Description,
	}

	_, err = s.alertService.UpdateAlert(ctx, alert.ID, updateReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to update alert with threat description")
	}

	s.logger.WithFields(logrus.Fields{
		"alert_id":    alert.ID,
		"threat_type": threat.ThreatType,
		"severity":    threat.Severity,
		"confidence":  threat.Confidence,
		"evidence":    string(evidenceJSON),
	}).Info("V2X threat alert created")

	return alert.ID, nil
}

// getThreatRuleID maps threat types to rule IDs
func (s *V2XSIEMService) getThreatRuleID(threatType string) uint {
	// This would be configured in the database with actual rule IDs
	threatRuleMap := map[string]uint{
		"position_jump_attack":          100,
		"impossible_speed":              101,
		"speed_anomaly":                 102,
		"message_flooding":              103,
		"message_dropout":               104,
		"geographic_boundary_violation": 105,
		"replay_attack":                 106,
		"message_count_anomaly":         107,
	}

	if ruleID, exists := threatRuleMap[threatType]; exists {
		return ruleID
	}

	return 199 // Default V2X rule ID
}

// updateMetrics updates performance and detection metrics
func (s *V2XSIEMService) updateMetrics(result *V2XProcessingResult) {
	s.metrics.MessagesProcessed++
	s.metrics.LastProcessedTime = time.Now()
	s.metrics.ProcessingLatencyMs = float64(result.ProcessingTime.Nanoseconds()) / 1000000.0

	if len(result.ThreatsDetected) > 0 {
		s.metrics.ThreatsDetected++
	}

	s.metrics.AlertsGenerated += uint64(len(result.AlertsCreated))

	// Update threat type counters
	for _, threat := range result.ThreatsDetected {
		if threat.ThreatDetected {
			switch threat.ThreatType {
			case "position_jump_attack":
				s.metrics.PositionJumpAttacks++
			case "impossible_speed", "speed_anomaly":
				s.metrics.SpeedAnomalies++
			case "message_flooding", "message_dropout", "message_count_anomaly":
				s.metrics.MessageAnomalies++
			case "geographic_boundary_violation":
				s.metrics.BoundaryViolations++
			case "replay_attack":
				s.metrics.ReplayAttacks++
			}
		}
	}
}

// GetMetrics returns current performance metrics
func (s *V2XSIEMService) GetMetrics() *V2XMetrics {
	return s.metrics
}

// GetDetectionConfig returns the current detection configuration
func (s *V2XSIEMService) GetDetectionConfig() *DetectionConfig {
	return s.detectionEngine.config
}

// UpdateDetectionConfig updates the detection engine configuration
func (s *V2XSIEMService) UpdateDetectionConfig(config *DetectionConfig) {
	s.detectionEngine.config = config
	s.logger.Info("V2X detection configuration updated")
}

// SetGeographicBounds configures allowed geographic boundaries
func (s *V2XSIEMService) SetGeographicBounds(bounds []GeographicBound) {
	s.detectionEngine.config.AllowedBounds = bounds
	s.logger.WithField("bounds_count", len(bounds)).Info("Geographic bounds updated")
}

// GetVehicleStates returns current vehicle states for monitoring
func (s *V2XSIEMService) GetVehicleStates() map[string]*VehicleState {
	return s.detectionEngine.vehicleStates
}

// ClearVehicleState removes a vehicle from tracking (for cleanup)
func (s *V2XSIEMService) ClearVehicleState(vehicleID string) {
	delete(s.detectionEngine.vehicleStates, vehicleID)
	s.logger.WithField("vehicle_id", vehicleID).Info("Vehicle state cleared")
}

// GetThreatStatistics returns aggregated threat detection statistics
func (s *V2XSIEMService) GetThreatStatistics() map[string]interface{} {
	return map[string]interface{}{
		"total_messages_processed":   s.metrics.MessagesProcessed,
		"total_threats_detected":     s.metrics.ThreatsDetected,
		"total_alerts_generated":     s.metrics.AlertsGenerated,
		"average_processing_latency": s.metrics.ProcessingLatencyMs,
		"last_processed_time":        s.metrics.LastProcessedTime,
		"position_jump_attacks":      s.metrics.PositionJumpAttacks,
		"speed_anomalies":            s.metrics.SpeedAnomalies,
		"message_anomalies":          s.metrics.MessageAnomalies,
		"boundary_violations":        s.metrics.BoundaryViolations,
		"replay_attacks":             s.metrics.ReplayAttacks,
		"active_vehicles":            len(s.detectionEngine.vehicleStates),
	}
}
