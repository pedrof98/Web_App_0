package siem

import (
	"time"
	"fmt"
	
	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
)

// V2XDashboardService provides data for V2X-specific dashboards
type V2XDashboardService struct {
	DB *gorm.DB
}

// NewV2XDashboardService creates a new V2X dashboard service
func NewV2XDashboardService(db *gorm.DB) *V2XDashboardService {
	return &V2XDashboardService{
		DB: db,
	}
}

// V2XSummary contains summary counts of V2X messages
type V2XSummary struct {
	Total         int64 `json:"total"`
	DSRC          int64 `json:"dsrc"`
	CV2X          int64 `json:"cv2x"`
	BSM           int64 `json:"bsm"`
	SPAT          int64 `json:"spat"`
	RSA           int64 `json:"rsa"`
	CAM           int64 `json:"cam"`
	DENM          int64 `json:"denm"`
	CPM           int64 `json:"cpm"`
	PC5Interface  int64 `json:"pc5_interface"`
	UuInterface   int64 `json:"uu_interface"`
}

// V2XSecuritySummary contains security-related counts for V2X messages
type V2XSecuritySummary struct {
	Total                int64 `json:"total"`
	ValidSignature       int64 `json:"valid_signature"`
	InvalidSignature     int64 `json:"invalid_signature"`
	HighTrustLevel       int64 `json:"high_trust_level"`
	LowTrustLevel        int64 `json:"low_trust_level"`
	DetectedAnomalies    int64 `json:"detected_anomalies"`
	HighConfidenceAnomaly int64 `json:"high_confidence_anomaly"`
}

// V2XAnomalySummary contains counts of different anomaly types
type V2XAnomalySummary struct {
	Total            int64 `json:"total"`
	PositionJump     int64 `json:"position_jump"`
	SpeedJump        int64 `json:"speed_jump"`
	HeadingJump      int64 `json:"heading_jump"`
	HighFrequency    int64 `json:"high_frequency"`
	ConflictingAlerts int64 `json:"conflicting_alerts"`
	TimingAnomaly    int64 `json:"timing_anomaly"`
	Other            int64 `json:"other"`
}

// VehicleLocation represents a vehicle's location data for mapping
type VehicleLocation struct {
	VehicleID  string    `json:"vehicle_id"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Speed      float32   `json:"speed"`
	Heading    float32   `json:"heading"`
	MessageType string   `json:"message_type"`
	Timestamp  time.Time `json:"timestamp"`
	RSSI       int16     `json:"rssi"`
	HasAnomaly bool      `json:"has_anomaly"`
}

// AlertLocation represents alert data for mapping
type AlertLocation struct {
	ID         uint      `json:"id"`
	AlertType  string    `json:"alert_type"` 
	Description string   `json:"description"`
	Priority   uint8     `json:"priority"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Radius     uint16    `json:"radius"`
	Timestamp  time.Time `json:"timestamp"`
}

// GetV2XSummary returns summary counts of V2X messages
func (s *V2XDashboardService) GetV2XSummary(timeRange string) (*V2XSummary, error) {
	var summary V2XSummary
	
	// Build query based on time range
	query := s.DB.Model(&models.V2XMessage{})
	timeFilter := getTimeWindowFilter(timeRange)
	if timeFilter != "" {
		query = query.Where(timeFilter)
	}
	
	// Get total count
	if err := query.Count(&summary.Total).Error; err != nil {
		return nil, err
	}
	
	// Get counts by protocol
	if err := query.Where("protocol = ?", models.ProtocolDSRC).Count(&summary.DSRC).Error; err != nil {
		return nil, err
	}
	
	if err := query.Where("protocol LIKE ?", "cv2x%").Count(&summary.CV2X).Error; err != nil {
		return nil, err
	}
	
	// Get counts by message type
	if err := query.Where("message_type = ?", models.MessageTypeBSM).Count(&summary.BSM).Error; err != nil {
		return nil, err
	}
	
	if err := query.Where("message_type = ?", models.MessageTypeSPAT).Count(&summary.SPAT).Error; err != nil {
		return nil, err
	}
	
	if err := query.Where("message_type = ?", models.MessageTypeRSA).Count(&summary.RSA).Error; err != nil {
		return nil, err
	}
	
	if err := query.Where("message_type = ?", models.MessageTypeCAM).Count(&summary.CAM).Error; err != nil {
		return nil, err
	}
	
	if err := query.Where("message_type = ?", models.MessageTypeDENM).Count(&summary.DENM).Error; err != nil {
		return nil, err
	}
	
	if err := query.Where("message_type = ?", models.MessageTypeCPM).Count(&summary.CPM).Error; err != nil {
		return nil, err
	}
	
	// Get CV2X counts by interface type
	var cv2xIds []uint
	if err := query.Where("protocol LIKE ?", "cv2x%").Pluck("id", &cv2xIds).Error; err != nil {
		return nil, err
	}
	
	if len(cv2xIds) > 0 {
		cv2xQuery := s.DB.Model(&models.CV2XMessage{}).Where("v2x_message_id IN ?", cv2xIds)
		
		// Count PC5 (direct) interfaces
		if err := cv2xQuery.Where("interface_type = ?", "PC5").Count(&summary.PC5Interface).Error; err != nil {
			return nil, err
		}
		
		// Count Uu (network) interfaces
		if err := cv2xQuery.Where("interface_type = ?", "Uu").Count(&summary.UuInterface).Error; err != nil {
			return nil, err
		}
	}
	
	return &summary, nil
}

// GetV2XSecuritySummary returns security-related summary for V2X messages
func (s *V2XDashboardService) GetV2XSecuritySummary(timeRange string) (*V2XSecuritySummary, error) {
	var summary V2XSecuritySummary
	
	// Build query based on time range
	timeFilter := getTimeWindowFilter(timeRange)
	messageQuery := s.DB.Model(&models.V2XMessage{})
	if timeFilter != "" {
		messageQuery = messageQuery.Where(timeFilter)
	}
	
	// Get total count
	if err := messageQuery.Count(&summary.Total).Error; err != nil {
		return nil, err
	}
	
	// Get message IDs within the time range
	var messageIds []uint
	if err := messageQuery.Pluck("id", &messageIds).Error; err != nil {
		return nil, err
	}
	
	if len(messageIds) > 0 {
		// Query for security info
		securityQuery := s.DB.Model(&models.V2XSecurityInfo{}).Where("v2x_message_id IN ?", messageIds)
		
		// Count valid signatures
		if err := securityQuery.Where("signature_valid = ?", true).Count(&summary.ValidSignature).Error; err != nil {
			return nil, err
		}
		
		// Count invalid signatures
		if err := securityQuery.Where("signature_valid = ?", false).Count(&summary.InvalidSignature).Error; err != nil {
			return nil, err
		}
		
		// Count high trust level (>= 7)
		if err := securityQuery.Where("trust_level >= ?", 7).Count(&summary.HighTrustLevel).Error; err != nil {
			return nil, err
		}
		
		// Count low trust level (< 3)
		if err := securityQuery.Where("trust_level < ?", 3).Count(&summary.LowTrustLevel).Error; err != nil {
			return nil, err
		}
		
		// Query for anomaly info
		anomalyQuery := s.DB.Model(&models.V2XAnomalyDetection{}).Where("v2x_message_id IN ?", messageIds)
		
		// Count total anomalies
		if err := anomalyQuery.Count(&summary.DetectedAnomalies).Error; err != nil {
			return nil, err
		}
		
		// Count high confidence anomalies
		if err := anomalyQuery.Where("confidence_score >= ?", 0.8).Count(&summary.HighConfidenceAnomaly).Error; err != nil {
			return nil, err
		}
	}
	
	return &summary, nil
}

// GetV2XAnomalySummary returns counts of different anomaly types
func (s *V2XDashboardService) GetV2XAnomalySummary(timeRange string) (*V2XAnomalySummary, error) {
	var summary V2XAnomalySummary
	
	// Build query based on time range
	timeFilter := getTimeWindowFilter(timeRange)
	messageQuery := s.DB.Model(&models.V2XMessage{})
	if timeFilter != "" {
		messageQuery = messageQuery.Where(timeFilter)
	}
	
	// Get message IDs within the time range
	var messageIds []uint
	if err := messageQuery.Pluck("id", &messageIds).Error; err != nil {
		return nil, err
	}
	
	if len(messageIds) > 0 {
		// Query for anomaly info
		anomalyQuery := s.DB.Model(&models.V2XAnomalyDetection{}).Where("v2x_message_id IN ?", messageIds)
		
		// Count total anomalies
		if err := anomalyQuery.Count(&summary.Total).Error; err != nil {
			return nil, err
		}
		
		// Count position jump anomalies
		if err := anomalyQuery.Where("anomaly_type = ?", "position_jump").Count(&summary.PositionJump).Error; err != nil {
			return nil, err
		}
		
		// Count speed jump anomalies
		if err := anomalyQuery.Where("anomaly_type = ?", "speed_jump").Count(&summary.SpeedJump).Error; err != nil {
			return nil, err
		}
		
		// Count heading jump anomalies
		if err := anomalyQuery.Where("anomaly_type = ?", "heading_jump").Count(&summary.HeadingJump).Error; err != nil {
			return nil, err
		}
		
		// Count high frequency anomalies
		if err := anomalyQuery.Where("anomaly_type = ?", "high_frequency").Count(&summary.HighFrequency).Error; err != nil {
			return nil, err
		}
		
		// Count conflicting alerts anomalies
		if err := anomalyQuery.Where("anomaly_type = ?", "conflicting_alerts").Count(&summary.ConflictingAlerts).Error; err != nil {
			return nil, err
		}
		
		// Count timing anomalies
		if err := anomalyQuery.Where("anomaly_type LIKE ?", "%timing%").Count(&summary.TimingAnomaly).Error; err != nil {
			return nil, err
		}
		
		// Count other anomalies
		otherTypes := []string{"position_jump", "speed_jump", "heading_jump", "high_frequency", "conflicting_alerts"}
		
		if err := anomalyQuery.Where("anomaly_type NOT IN ? AND anomaly_type NOT LIKE ?", otherTypes, "%timing%").
			Count(&summary.Other).Error; err != nil {
			return nil, err
		}
		
	}
	
	return &summary, nil
}

// GetRecentVehicleLocations returns recent vehicle locations for mapping
func (s *V2XDashboardService) GetRecentVehicleLocations(limit int) ([]VehicleLocation, error) {
	if limit <= 0 {
		limit = 100 // Default limit
	}
	
	// Get the most recent BSM for each vehicle
	type Result struct {
		ID          uint
		VehicleID   string
		Latitude    float64
		Longitude   float64
		MessageType string
		Timestamp   time.Time
		RSSI        int16
	}
	
	var results []Result
	
	// This query finds the most recent message for each source_id (vehicle)
	// We use a subquery to get max IDs for each source_id
	subQuery := s.DB.Model(&models.V2XMessage{}).
		Select("MAX(id) as id").
		Where("latitude != 0 AND longitude != 0").
		Where("message_type IN ?", []string{
			string(models.MessageTypeBSM), 
			string(models.MessageTypeCV2XBSM), 
			string(models.MessageTypeCAM),
		}).
		Group("source_id").
		Limit(limit)
	
	// Main query to get vehicle locations
	query := s.DB.Model(&models.V2XMessage{}).
		Select("id, source_id as vehicle_id, latitude, longitude, message_type, timestamp, rssi").
		Where("id IN (?)", subQuery)
	
	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}
	
	// Convert to VehicleLocation objects
	locations := make([]VehicleLocation, 0, len(results))
	
	for _, result := range results {
		// Get additional BSM data (speed, heading)
		var speed float32
		var heading float32
		var hasAnomaly bool
		
		// Check if there are any anomalies for this message
		var anomalyCount int64
		s.DB.Model(&models.V2XAnomalyDetection{}).
			Where("v2x_message_id = ?", result.ID).
			Count(&anomalyCount)
		
		hasAnomaly = anomalyCount > 0
		
		// Get BSM data based on message type
		switch result.MessageType {
		case string(models.MessageTypeBSM), string(models.MessageTypeCV2XBSM), string(models.MessageTypeCAM):
			var bsm models.BasicSafetyMessage
			if err := s.DB.Where("v2x_message_id = ?", result.ID).First(&bsm).Error; err == nil {
				speed = bsm.Speed
				heading = bsm.Heading
			}
		}
		
		locations = append(locations, VehicleLocation{
			VehicleID:   result.VehicleID,
			Latitude:    result.Latitude,
			Longitude:   result.Longitude,
			Speed:       speed,
			Heading:     heading,
			MessageType: result.MessageType,
			Timestamp:   result.Timestamp,
			RSSI:        result.RSSI,
			HasAnomaly:  hasAnomaly,
		})
	}
	
	return locations, nil
}

// GetActiveAlerts returns active roadside alerts for mapping
func (s *V2XDashboardService) GetActiveAlerts(timeRange string) ([]AlertLocation, error) {
	// Build query based on time range
	timeFilter := getTimeWindowFilter(timeRange)
	
	// Find RSA and DENM messages within time window
	messageQuery := s.DB.Model(&models.V2XMessage{}).
		Where("message_type IN ?", []string{
			string(models.MessageTypeRSA),
			string(models.MessageTypeDENM),
		})
	
	if timeFilter != "" {
		messageQuery = messageQuery.Where(timeFilter)
	}
	
	// Get message IDs
	var messageIds []uint
	if err := messageQuery.Pluck("id", &messageIds).Error; err != nil {
		return nil, err
	}
	
	if len(messageIds) == 0 {
		return []AlertLocation{}, nil
	}
	
	// Find associated alerts (RSA model is used for both RSA and DENM)
	var alerts []struct {
		ID          uint
		V2XMessageID uint
		AlertType   uint16
		Description string
		Priority    uint8
		Radius      uint16
	}
	
	if err := s.DB.Model(&models.RoadsideAlert{}).
		Where("v2x_message_id IN ?", messageIds).
		Find(&alerts).Error; err != nil {
		return nil, err
	}
	
	// Create result objects with location data
	locations := make([]AlertLocation, 0, len(alerts))
	
	for _, alert := range alerts {
		var message models.V2XMessage
		if err := s.DB.First(&message, alert.V2XMessageID).Error; err != nil {
			continue
		}
		
		locations = append(locations, AlertLocation{
			ID:         alert.ID,
			AlertType:  formatAlertType(alert.AlertType),
			Description: alert.Description,
			Priority:   alert.Priority,
			Latitude:   message.Latitude,
			Longitude:  message.Longitude,
			Radius:     alert.Radius,
			Timestamp:  message.Timestamp,
		})
	}
	
	return locations, nil
}

// Helper function to format alert types
func formatAlertType(alertType uint16) string {
	switch alertType {
	case 1:
		return "accident"
	case 2:
		return "roadwork"
	case 3:
		return "weather"
	case 4:
		return "hazard"
	case 5:
		return "traffic"
	default:
		return fmt.Sprintf("other_%d", alertType)
	}
}

// Helper function to get time window filter
func getTimeWindowFilter(timeRange string) string {
	now := time.Now()
	
	switch timeRange {
	case "last_5_minutes":
		startTime := now.Add(-5 * time.Minute)
		return fmt.Sprintf("timestamp >= '%s'", startTime.Format("2006-01-02 15:04:05"))
		
	case "last_15_minutes":
		startTime := now.Add(-15 * time.Minute)
		return fmt.Sprintf("timestamp >= '%s'", startTime.Format("2006-01-02 15:04:05"))
		
	case "last_hour":
		startTime := now.Add(-1 * time.Hour)
		return fmt.Sprintf("timestamp >= '%s'", startTime.Format("2006-01-02 15:04:05"))
		
	case "last_day":
		startTime := now.AddDate(0, 0, -1)
		return fmt.Sprintf("timestamp >= '%s'", startTime.Format("2006-01-02 15:04:05"))
		
	default:
		// Default to last hour if no valid time range specified
		startTime := now.Add(-1 * time.Hour)
		return fmt.Sprintf("timestamp >= '%s'", startTime.Format("2006-01-02 15:04:05"))
	}
}