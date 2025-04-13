package v2x

import (
	"log"
	"math"
	"time"

	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
)

// V2XAnomalyDetector detects anomalies in V2X messages
type V2XAnomalyDetector struct {
	DB *gorm.DB
}

// NewV2XAnomalyDetector creates a new anomaly detector
func NewV2XAnomalyDetector(db *gorm.DB) *V2XAnomalyDetector {
	return &V2XAnomalyDetector{
		DB: db,
	}
}

// AnomalyParams contains parameters for anomaly detection
type AnomalyParams struct {
	// Position jump threshold in meters
	PositionJumpThreshold float64
	// Speed jump threshold in m/s
	SpeedJumpThreshold float64
	// Heading jump threshold in degrees
	HeadingJumpThreshold float64
	// Message frequency threshold (messages per second)
	MessageFrequencyThreshold float64
	// Time window for frequency analysis in seconds
	FrequencyTimeWindow float64
}

// DefaultAnomalyParams returns default parameters for anomaly detection
func DefaultAnomalyParams() AnomalyParams {
	return AnomalyParams{
		PositionJumpThreshold:    100.0,  // 100 meters
		SpeedJumpThreshold:       10.0,   // 10 m/s
		HeadingJumpThreshold:     45.0,   // 45 degrees
		MessageFrequencyThreshold: 10.0,   // Max 10 messages per second
		FrequencyTimeWindow:      5.0,    // 5 second window
	}
}

// DetectAnomalies detects anomalies in a V2X message
func (d *V2XAnomalyDetector) DetectAnomalies(message *models.V2XMessage) ([]models.V2XAnomalyDetection, error) {
	var anomalies []models.V2XAnomalyDetection
	
	// Get detection parameters
	params := DefaultAnomalyParams()
	
	// Different detection strategies based on message type
	switch message.MessageType {
	case models.MessageTypeBSM, models.MessageTypeCV2XBSM:
		// For BSMs, check position/speed/heading anomalies
		positionAnomalies, err := d.detectBSMPositionAnomalies(message, params)
		if err != nil {
			log.Printf("Error detecting position anomalies: %v", err)
		} else {
			anomalies = append(anomalies, positionAnomalies...)
		}
		
		// Check message frequency anomalies
		frequencyAnomalies, err := d.detectMessageFrequencyAnomalies(message, params)
		if err != nil {
			log.Printf("Error detecting frequency anomalies: %v", err)
		} else {
			anomalies = append(anomalies, frequencyAnomalies...)
		}
		
	case models.MessageTypeSPAT:
		// For SPATs, check timing anomalies
		timingAnomalies, err := d.detectSPATTimingAnomalies(message)
		if err != nil {
			log.Printf("Error detecting SPAT timing anomalies: %v", err)
		} else {
			anomalies = append(anomalies, timingAnomalies...)
		}
		
	case models.MessageTypeRSA:
		// For RSAs, check geographic consistency
		geoAnomalies, err := d.detectRSAGeographicAnomalies(message)
		if err != nil {
			log.Printf("Error detecting RSA geographic anomalies: %v", err)
		} else {
			anomalies = append(anomalies, geoAnomalies...)
		}
	}
	
	// Save detected anomalies
	for i := range anomalies {
		anomalies[i].V2XMessageID = message.ID
		anomalies[i].CreatedAt = time.Now()
		
		if err := d.DB.Create(&anomalies[i]).Error; err != nil {
			log.Printf("Error saving anomaly detection: %v", err)
		}
	}
	
	return anomalies, nil
}

// detectBSMPositionAnomalies detects anomalies in BSM position data
func (d *V2XAnomalyDetector) detectBSMPositionAnomalies(message *models.V2XMessage, params AnomalyParams) ([]models.V2XAnomalyDetection, error) {
	var anomalies []models.V2XAnomalyDetection
	
	// Get previous message from same source within a recent timeframe
	var previousMessage models.V2XMessage
	timeWindow := time.Duration(-10) * time.Second // Look for messages in the last 10 seconds
	
	err := d.DB.Where("source_id = ? AND id != ? AND timestamp > ?", 
		message.SourceID, message.ID, message.Timestamp.Add(timeWindow)).
		Order("timestamp DESC").
		First(&previousMessage).Error
		
	// If there's no previous message, we can't detect position anomalies
	if err != nil {
		return anomalies, nil
	}
	
	// Calculate time difference in seconds
	timeDiff := message.Timestamp.Sub(previousMessage.Timestamp).Seconds()
	if timeDiff <= 0 {
		// Messages out of order or same timestamp, can't calculate rates
		return anomalies, nil
	}
	
	// If both messages have lat/long, check position jumps
	if message.Latitude != 0 && message.Longitude != 0 && 
	   previousMessage.Latitude != 0 && previousMessage.Longitude != 0 {
		
		// Calculate distance in meters
		distance := haversineDistance(
			message.Latitude, message.Longitude,
			previousMessage.Latitude, previousMessage.Longitude)
		
		// Calculate travel speed in m/s based on the distance and time
		travelSpeed := distance / timeDiff
		
		// Check if there's a large position jump
		if travelSpeed > params.PositionJumpThreshold {
			// Create an anomaly detection
			anomalies = append(anomalies, models.V2XAnomalyDetection{
				AnomalyType:      "position_jump",
				ConfidenceScore:  calculateConfidence(travelSpeed / params.PositionJumpThreshold),
				Description:      "Unusual position change detected",
			})
		}
	}
	
	// Check for BSM-specific data anomalies
	if message.MessageType == models.MessageTypeBSM || message.MessageType == models.MessageTypeCV2XBSM {
		// Get the BSM data for both messages
		var currentBSM models.BasicSafetyMessage
		var previousBSM models.BasicSafetyMessage
		
		// Load current BSM
		err = d.DB.Where("v2_x_message_id = ?", message.ID).First(&currentBSM).Error
		if err != nil {
			return anomalies, nil // Can't check BSM specifics
		}
		
		// Load previous BSM
		err = d.DB.Where("v2_x_message_id = ?", previousMessage.ID).First(&previousBSM).Error
		if err != nil {
			return anomalies, nil // Can't compare BSMs
		}
		
		// Check for speed jumps
		speedDiff := math.Abs(float64(currentBSM.Speed - previousBSM.Speed))
		if speedDiff > params.SpeedJumpThreshold {
			anomalies = append(anomalies, models.V2XAnomalyDetection{
				AnomalyType:      "speed_jump",
				ConfidenceScore:  calculateConfidence(speedDiff / params.SpeedJumpThreshold),
				Description:      "Unusual speed change detected",
			})
		}
		
		// Check for heading jumps (accounting for 0/360 degrees wrap)
		headingDiff := math.Abs(float64(currentBSM.Heading - previousBSM.Heading))
		if headingDiff > 180 {
			headingDiff = 360 - headingDiff // Take the smaller angle
		}
		
		if headingDiff > params.HeadingJumpThreshold && currentBSM.Speed > 5.0 {
			// Only consider heading jumps significant if vehicle is moving
			anomalies = append(anomalies, models.V2XAnomalyDetection{
				AnomalyType:      "heading_jump",
				ConfidenceScore:  calculateConfidence(headingDiff / params.HeadingJumpThreshold),
				Description:      "Unusual heading change detected",
			})
		}
	}
	
	return anomalies, nil
}

// detectMessageFrequencyAnomalies detects abnormal message frequencies
func (d *V2XAnomalyDetector) detectMessageFrequencyAnomalies(message *models.V2XMessage, params AnomalyParams) ([]models.V2XAnomalyDetection, error) {
	var anomalies []models.V2XAnomalyDetection
	
	// Count messages from the same source within the time window
	timeWindow := time.Duration(-int64(params.FrequencyTimeWindow * float64(time.Second)))
	
	var count int64
	err := d.DB.Model(&models.V2XMessage{}).
		Where("source_id = ? AND timestamp > ?", 
			message.SourceID, message.Timestamp.Add(timeWindow)).
		Count(&count).Error
			
	if err != nil {
		return anomalies, err
	}
	
	// Calculate messages per second
	messagesPerSecond := float64(count) / params.FrequencyTimeWindow
	
	// Check if frequency exceeds threshold
	if messagesPerSecond > params.MessageFrequencyThreshold {
		anomalies = append(anomalies, models.V2XAnomalyDetection{
			AnomalyType:      "high_frequency",
			ConfidenceScore:  calculateConfidence(messagesPerSecond / params.MessageFrequencyThreshold),
			Description:      "Abnormally high message frequency detected",
		})
	}
	
	return anomalies, nil
}

// detectSPATTimingAnomalies detects anomalies in SPAT timing
func (d *V2XAnomalyDetector) detectSPATTimingAnomalies(message *models.V2XMessage) ([]models.V2XAnomalyDetection, error) {
	var anomalies []models.V2XAnomalyDetection
	
	// Load the SPAT data
	var spat models.SignalPhaseAndTiming
	err := d.DB.Where("v2_x_message_id = ?", message.ID).First(&spat).Error
	if err != nil {
		return anomalies, err
	}
	
	// Load the phase states
	var phaseStates []models.PhaseState
	err = d.DB.Where("spat_message_id = ?", spat.ID).Find(&phaseStates).Error
	if err != nil {
		return anomalies, err
	}
	
	// Check for illogical timing patterns
	for _, phase := range phaseStates {
		// Check if min end time is after max end time
		if phase.MinEndTime > phase.MaxEndTime {
			anomalies = append(anomalies, models.V2XAnomalyDetection{
				AnomalyType:      "illogical_timing",
				ConfidenceScore:  0.95,
				Description:      "Minimum end time exceeds maximum end time",
			})
		}
		
		// Check for unreasonably short or long phases
		if phase.LightState == "green" || phase.LightState == "yellow" {
			phaseDuration := phase.MaxEndTime - phase.StartTime
			
			if phaseDuration < 1000 && phase.LightState == "green" { // Less than 1 second green
				anomalies = append(anomalies, models.V2XAnomalyDetection{
					AnomalyType:      "short_phase",
					ConfidenceScore:  0.85,
					Description:      "Unreasonably short green phase detected",
				})
			} else if phaseDuration > 60000 { // More than 1 minute
				anomalies = append(anomalies, models.V2XAnomalyDetection{
					AnomalyType:      "long_phase",
					ConfidenceScore:  0.75,
					Description:      "Unusually long phase detected",
				})
			}
		}
	}
	
	return anomalies, nil
}

// detectRSAGeographicAnomalies detects geographic anomalies in RSA messages
func (d *V2XAnomalyDetector) detectRSAGeographicAnomalies(message *models.V2XMessage) ([]models.V2XAnomalyDetection, error) {
	var anomalies []models.V2XAnomalyDetection
	
	// Load the RSA data
	var rsa models.RoadsideAlert
	err := d.DB.Where("v2_x_message_id = ?", message.ID).First(&rsa).Error
	if err != nil {
		return anomalies, err
	}
	
	// Check for conflicting alerts in the same area
	var conflictingAlerts []models.RoadsideAlert
	
	// Find alerts of different types in the same area
	err = d.DB.Joins("JOIN v2x_messages ON v2x_messages.id = roadside_alerts.v2_x_message_id").
		Where("alert_type != ? AND roadside_alerts.id != ?", rsa.AlertType, rsa.ID).
		Where("v2x_messages.latitude BETWEEN ? AND ?", 
			message.Latitude - 0.01, message.Latitude + 0.01).
		Where("v2x_messages.longitude BETWEEN ? AND ?", 
			message.Longitude - 0.01, message.Longitude + 0.01).
		Find(&conflictingAlerts).Error
	
	if err != nil {
		return anomalies, err
	}
	
	if len(conflictingAlerts) > 0 {
		// Check if any alert conflicts directly (e.g., "clear road" vs "blocked road")
		for _, alert := range conflictingAlerts {
			if isConflictingAlertType(rsa.AlertType, alert.AlertType) {
				anomalies = append(anomalies, models.V2XAnomalyDetection{
					AnomalyType:      "conflicting_alerts",
					ConfidenceScore:  0.85,
					Description:      "Conflicting roadside alerts in the same area",
				})
				break
			}
		}
	}
	
	return anomalies, nil
}

// isConflictingAlertType checks if two alert types logically conflict
// Note: This is a simplified implementation and would be more comprehensive in a real system
func isConflictingAlertType(type1, type2 uint16) bool {
	// Example conflicting pairs:
	// 1: Accident vs 100: Clear Road 
	// 2: Road Work vs 101: Road Clear of Construction
	conflictingPairs := map[uint16][]uint16{
		1: {100},
		2: {101},
		// Add more pairs as needed
	}
	
	if conflicts, ok := conflictingPairs[type1]; ok {
		for _, conflict := range conflicts {
			if conflict == type2 {
				return true
			}
		}
	}
	
	// Check the reverse direction
	if conflicts, ok := conflictingPairs[type2]; ok {
		for _, conflict := range conflicts {
			if conflict == type1 {
				return true
			}
		}
	}
	
	return false
}

// calculateConfidence calculates a confidence score between 0 and 1
// based on how much a value exceeds a threshold
func calculateConfidence(ratio float64) float32 {
	// Scale to between 0.5 and 0.99
	confidence := 0.5 + math.Min(0.49, 0.49*(ratio-1.0)/9.0)
	return float32(confidence)
}

// haversineDistance calculates the distance between two sets of coordinates in meters
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Earth's radius in meters
	const earthRadius = 6371000.0
	
	// Convert to radians
	lat1 = lat1 * math.Pi / 180.0
	lon1 = lon1 * math.Pi / 180.0
	lat2 = lat2 * math.Pi / 180.0
	lon2 = lon2 * math.Pi / 180.0
	
	// Haversine formula
	dLat := lat2 - lat1
	dLon := lon2 - lon1
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + 
		 math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	distance := earthRadius * c
	
	return distance
}