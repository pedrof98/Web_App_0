package models

import (
	"time"
)

// V2XMessageType represents the type of V2X message
type V2XMessageType string


const (
	// DSRC message types
	MessageTypeBSM		V2XMessageType = "basic_safety_message"
	MessageTypeSPAT		V2XMessageType = "signal_phase_and_timing"
	MessageTypeMAP		V2XMessageType = "map_data"
	MessageTypeTIM		V2XMessageType = "traveler_information"
	MessageTypeICA		V2XMessageType = "intersection_collision_alert"
	MessageTypeEVA		V2XMessageType = "emergency_vehicle_alert"
	MessageTypeRSA		V2XMessageType = "roadside_alert"


	// C-V2X specific message types
	MessageTypeCV2XBSM	V2XMessageType = "cv2x_basic_safety_message"
	MessageTypeCAM		V2XMessageType = "cooperative_awareness_message"
	MessageTypeDENM		V2XMessageType = "decentralized_environmental_notification_message"
	MessageTypeCPM		V2XMessageType = "collective_perception_message"
	MessageTypeMANEUVER	V2XMessageType = "maneuver_coordination_message"
)


// V2XProtocol represents the communication protocol used
type V2XProtocol string

const (
	ProtocolDSRC		V2XProtocol = "dsrc" //DSRC/WAVE (802.11p)
	ProtocolCV2XMode4	V2XProtocol = "cv2x_pc5" // C-V2X PC5 interface (direct)
	ProtocolCV2XUu		V2XProtocol = "cv2x_uu"  // C-V2X Uu interface (network)
)

// V2XMessage represents the base structure for all V2X messages
type V2XMessage struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	Protocol     V2XProtocol   `gorm:"not null" json:"protocol"`
	MessageType  V2XMessageType `gorm:"not null" json:"message_type"`
	RawData      []byte        `gorm:"type:bytea" json:"raw_data"`
	Timestamp    time.Time     `gorm:"not null;index" json:"timestamp"`
	ReceivedAt   time.Time     `gorm:"not null" json:"received_at"`
	RSSI         int16         `json:"rssi"` // Received Signal Strength Indicator
	SourceID     string        `json:"source_id"` // Vehicle ID or RSU ID
	Latitude     float64       `json:"latitude"`
	Longitude    float64       `json:"longitude"`
	Elevation    float64       `json:"elevation,omitempty"`
	CreatedAt    time.Time     `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for V2XMessage
func (V2XMessage) TableName() string {
	return "v2x_messages"
}

// BSMAcceleration represents the acceleration data in a BSM
type BSMAcceleration struct {
	LateralAccel      float32 `json:"lateral_accel" gorm:"column:lateral_accel"`
	LongitudinalAccel float32 `json:"longitudinal_accel" gorm:"column:longitudinal_accel"`
	VerticalAccel     float32 `json:"vertical_accel" gorm:"column:vertical_accel"`
	YawRate           float32 `json:"yaw_rate" gorm:"column:yaw_rate"`
}

// BSMBrakeStatus represents the brake status in a BSM
type BSMBrakeStatus struct {
	BrakeApplied      bool `json:"brake_applied" gorm:"column:brake_applied"`
	TractionControl   bool `json:"traction_control" gorm:"column:traction_control"`
	ABS               bool `json:"abs" gorm:"column:abs"`
	StabilityControl  bool `json:"stability_control" gorm:"column:stability_control"`
	BrakeBoost        bool `json:"brake_boost" gorm:"column:brake_boost"`
	AuxiliaryBrakes   bool `json:"auxiliary_brakes" gorm:"column:auxiliary_brakes"`
}

// BSMVehicleSize represents the vehicle size in a BSM
type BSMVehicleSize struct {
	Width  float32 `json:"width" gorm:"column:width"`
	Length float32 `json:"length" gorm:"column:length"`
	Height float32 `json:"height,omitempty" gorm:"column:height"`
}

// BasicSafetyMessage (BSM) represents J2735 BSM structure
type BasicSafetyMessage struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	V2XMessageID    uint      `gorm:"not null" json:"v2x_message_id"`
	V2XMessage      V2XMessage `gorm:"foreignKey:V2XMessageID" json:"v2x_message"`
	TemporaryID     uint32    `json:"temporary_id"` // Temporary vehicle ID
	MessageCount    uint8     `json:"message_count"`
	SecMark         uint16    `json:"sec_mark"` // milliseconds within the minute
	Speed           float32   `json:"speed"`    // in m/s
	Heading         float32   `json:"heading"`  // in degrees
	
	// Flatten the embedded structs to avoid GORM migration issues
	LateralAccel      float32 `json:"lateral_accel"`
	LongitudinalAccel float32 `json:"longitudinal_accel"`
	VerticalAccel     float32 `json:"vertical_accel"`
	YawRate           float32 `json:"yaw_rate"`
	
	BrakeApplied      bool    `json:"brake_applied"`
	TractionControl   bool    `json:"traction_control"`
	ABS               bool    `json:"abs"`
	StabilityControl  bool    `json:"stability_control"`
	BrakeBoost        bool    `json:"brake_boost"`
	AuxiliaryBrakes   bool    `json:"auxiliary_brakes"`
	
	Width             float32 `json:"width"`
	Length            float32 `json:"length"`
	Height            float32 `json:"height,omitempty"`
	
	VehicleClass    uint8     `json:"vehicle_class"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for BasicSafetyMessage
func (BasicSafetyMessage) TableName() string {
	return "basic_safety_messages"
}

// SignalPhaseAndTiming (SPAT) represents traffic signal data
type SignalPhaseAndTiming struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	V2XMessageID    uint      `gorm:"not null" json:"v2x_message_id"`
	V2XMessage      V2XMessage `gorm:"foreignKey:V2XMessageID" json:"v2x_message"`
	IntersectionID  uint32    `json:"intersection_id"`
	MsgCount        uint8     `json:"msg_count"`
	PhaseStates     []PhaseState `gorm:"foreignKey:SPATMessageID" json:"phase_states"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for SignalPhaseAndTiming
func (SignalPhaseAndTiming) TableName() string {
	return "signal_phase_and_timing"
}

// PhaseState represents the state of a traffic signal phase
type PhaseState struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	SPATMessageID   uint      `gorm:"not null" json:"spat_message_id"`
	PhaseID         uint8     `json:"phase_id"` // Signal group ID
	LightState      string    `json:"light_state"` // Red, Yellow, Green, etc.
	StartTime       uint16    `json:"start_time"` // Start time in seconds from midnight
	MinEndTime      uint16    `json:"min_end_time"`
	MaxEndTime      uint16    `json:"max_end_time"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for PhaseState
func (PhaseState) TableName() string {
	return "phase_states"
}

// RoadsideAlert (RSA) represents alert messages from roadside equipment
type RoadsideAlert struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	V2XMessageID    uint      `gorm:"not null" json:"v2x_message_id"`
	V2XMessage      V2XMessage `gorm:"foreignKey:V2XMessageID" json:"v2x_message"`
	AlertType       uint16    `json:"alert_type"`
	Description     string    `json:"description"`
	Priority        uint8     `json:"priority"`
	Radius          uint16    `json:"radius"` // Range of the alert in meters
	Duration        uint16    `json:"duration"` // Duration in seconds
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for RoadsideAlert
func (RoadsideAlert) TableName() string {
	return "roadside_alerts"
}

// CV2XMessage represents cellular V2X specific message attributes
type CV2XMessage struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	V2XMessageID    uint      `gorm:"not null" json:"v2x_message_id"`
	V2XMessage      V2XMessage `gorm:"foreignKey:V2XMessageID" json:"v2x_message"`
	InterfaceType   string    `json:"interface_type"` // PC5 or Uu
	QoSInfo         uint8     `json:"qos_info"`      // Quality of Service information
	PLMNInfo        string    `gorm:"type:varchar(50)" json:"plmn_info"`      // Public Land Mobile Network info
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for CV2XMessage
func (CV2XMessage) TableName() string {
	return "cv2x_messages"
}

// V2XSecurityInfo represents security-related metadata for V2X messages
type V2XSecurityInfo struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	V2XMessageID    uint      `gorm:"not null;uniqueIndex" json:"v2x_message_id"`
	V2XMessage      V2XMessage `gorm:"foreignKey:V2XMessageID" json:"v2x_message"`
	SignatureValid  bool      `json:"signature_valid"`
	CertificateID   string    `json:"certificate_id"`
	TrustLevel      uint8     `json:"trust_level"`
	ValidationError string    `json:"validation_error"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for V2XSecurityInfo
func (V2XSecurityInfo) TableName() string {
	return "v2x_security_info"
}

// V2XAnomalyDetection represents anomaly detection results for V2X messages
type V2XAnomalyDetection struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	V2XMessageID    uint      `gorm:"not null" json:"v2x_message_id"`
	V2XMessage      V2XMessage `gorm:"foreignKey:V2XMessageID" json:"v2x_message"`
	AnomalyType     string    `json:"anomaly_type"`
	ConfidenceScore float32   `json:"confidence_score"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName returns the table name for V2XAnomalyDetection
func (V2XAnomalyDetection) TableName() string {
	return "v2x_anomaly_detections"
}