package collectors

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"time"
	
	"traffic-monitoring-go/app/models"
)

// CV2XParser provides parsing functionality for C-V2X messages
type CV2XParser struct{}

// NewCV2XParser creates a new C-V2X parser
func NewCV2XParser() *CV2XParser {
	return &CV2XParser{}
}

// Constants for C-V2X message types
const (
	CV2XMessageTypeBSM  byte = 1 // C-V2X version of BSM
	CV2XMessageTypeCAM  byte = 2 // Cooperative Awareness Message (European)
	CV2XMessageTypeDENM byte = 3 // Decentralized Environmental Notification Message
	CV2XMessageTypeCPM  byte = 4 // Collective Perception Message
)

// InterfaceType represents the C-V2X interface type (PC5 direct or Uu network)
type InterfaceType byte

const (
	InterfaceTypePC5 InterfaceType = 0    // Direct communications (PC5)
	InterfaceTypeUu  InterfaceType = 0x80 // Network communications (Uu)
)

// ParseMessageType determines the type of C-V2X message
func (p *CV2XParser) ParseMessageType(data []byte) (byte, InterfaceType, error) {
	if len(data) < 2 {
		return 0, 0, fmt.Errorf("message too short")
	}
	messageType := data[0]
	interfaceType := InterfaceType(data[1] & 0x80) // Extract interface type bit
	return messageType, interfaceType, nil
}

// ParseCV2XBSM parses a C-V2X Basic Safety Message
func (p *CV2XParser) ParseCV2XBSM(data []byte, interfaceType InterfaceType) (*models.BasicSafetyMessage, *models.CV2XMessage, *models.V2XMessage, error) {
	if len(data) < 24 {
		return nil, nil, nil, fmt.Errorf("CV2X BSM data too short, expected at least 24 bytes, got %d", len(data))
	}
	
	// Skip the message type and interface type bytes
	data = data[2:]
	
	bsm := &models.BasicSafetyMessage{}
	
	// Parse temporary ID (4 bytes)
	bsm.TemporaryID = binary.BigEndian.Uint32(data[0:4])
	
	// Parse message count (1 byte)
	bsm.MessageCount = uint8(data[4])
	
	// Parse timestamp (4 bytes) - Unix timestamp
	timestamp := binary.BigEndian.Uint32(data[5:9])
	
	// Parse latitude and longitude (8 bytes each - in IEEE 754 double precision)
	latitude := math.Float64frombits(binary.BigEndian.Uint64(data[9:17]))
	longitude := math.Float64frombits(binary.BigEndian.Uint64(data[17:25]))
	
	// Parse speed (4 bytes - IEEE 754 single precision)
	speed := math.Float32frombits(binary.BigEndian.Uint32(data[25:29]))
	
	// Parse heading (4 bytes - IEEE 754 single precision)
	heading := math.Float32frombits(binary.BigEndian.Uint32(data[29:33]))
	
	// Additional CV2X-specific field - QoS info (1 byte)
	qosInfo := uint8(0)
	if len(data) >= 34 {
		qosInfo = uint8(data[33])
	}
	
	// Set values
	bsm.Speed = speed
	bsm.Heading = heading
	
	// Parse acceleration, brake status, and vehicle size if available
	if len(data) >= 45 {
		// Acceleration (lateral, longitudinal, vertical) - in IEEE 754 single precision
		lateralAccel := math.Float32frombits(binary.BigEndian.Uint32(data[34:38]))
		longAccel := math.Float32frombits(binary.BigEndian.Uint32(data[38:42]))
		bsm.LateralAccel = lateralAccel
		bsm.LongitudinalAccel = longAccel
		
		// Brake status (1 byte bit flags)
		brakeStatus := data[42]
		bsm.BrakeApplied = (brakeStatus & 0x01) != 0
		bsm.TractionControl = (brakeStatus & 0x02) != 0
		bsm.ABS = (brakeStatus & 0x04) != 0
		bsm.StabilityControl = (brakeStatus & 0x08) != 0
		bsm.BrakeBoost = (brakeStatus & 0x10) != 0
		bsm.AuxiliaryBrakes = (brakeStatus & 0x20) != 0
	}
	
	// Create timestamp based on the parsed Unix timestamp
	messageTime := time.Unix(int64(timestamp), 0)
	
	// Create a v2x message with the parsed data
	v2xMessage := &models.V2XMessage{
		Protocol:    models.ProtocolCV2XMode4,
		MessageType: models.MessageTypeCV2XBSM,
		RawData:     append([]byte{}, data...), // Make a copy to avoid data races
		Timestamp:   messageTime,
		ReceivedAt:  time.Now(),
		SourceID:    fmt.Sprintf("CVX-%08X", bsm.TemporaryID),
		Latitude:    latitude,
		Longitude:   longitude,
	}
	
	// TODO: In a real system, calculate RSSI from received signal
	v2xMessage.RSSI = int16(-75 + rand.Intn(30)) // Simulated RSSI between -75 and -45 dBm
	
	// Create CV2X specific info
	interfaceTypeStr := "PC5"
	plmnInfo := ""
	
	// Add PLMN info if it's a Uu (network) interface
	if interfaceType == InterfaceTypeUu {
		interfaceTypeStr = "Uu"
		plmnInfo = "310-410" // Example PLMN ID (simulated)
	}
	
	cv2xInfo := &models.CV2XMessage{
		InterfaceType: interfaceTypeStr,
		QoSInfo:       qosInfo,
		PLMNInfo:      plmnInfo,
	}
	
	return bsm, cv2xInfo, v2xMessage, nil
}

// ParseDENM parses a Decentralized Environmental Notification Message
func (p *CV2XParser) ParseDENM(data []byte, interfaceType InterfaceType) (*models.RoadsideAlert, *models.CV2XMessage, *models.V2XMessage, error) {
	if len(data) < 28 {
		return nil, nil, nil, fmt.Errorf("DENM data too short, expected at least 28 bytes, got %d", len(data))
	}
	
	// Skip the message type and interface type bytes
	data = data[2:]
	
	// Create a RoadsideAlert struct to store the data
	// (We reuse the RoadsideAlert model since DENM is functionally similar)
	rsa := &models.RoadsideAlert{}
	
	// Parse event ID (4 bytes)
	eventID := binary.BigEndian.Uint32(data[0:4])
	
	// Parse message count (1 byte)
	_ = uint8(data[4]) // Unused in DENM but included for compatibility
	
	// Parse event type (1 byte)
	eventType := uint16(data[5])
	rsa.AlertType = eventType
	
	// Map event type to description
	var description string
	switch eventType {
	case 1:
		description = "Traffic accident"
	case 2:
		description = "Roadworks"
	case 3:
		description = "Adverse weather condition"
	case 4:
		description = "Hazardous location"
	case 5:
		description = "Traffic condition"
	default:
		description = fmt.Sprintf("Unknown event type: %d", eventType)
	}
	rsa.Description = description
	
	// Parse timestamp (4 bytes)
	timestamp := binary.BigEndian.Uint32(data[6:10])
	
	// Parse location (8 bytes for latitude and longitude each)
	latitude := math.Float64frombits(binary.BigEndian.Uint64(data[10:18]))
	longitude := math.Float64frombits(binary.BigEndian.Uint64(data[18:26]))
	
	// Parse radius (2 bytes)
	radius := binary.BigEndian.Uint16(data[26:28])
	rsa.Radius = radius
	
	// Parse duration (2 bytes) if available
	if len(data) >= 30 {
		duration := binary.BigEndian.Uint16(data[28:30])
		rsa.Duration = duration
	}
	
	// Calculate priority based on event type (simplified approach)
	// In a real implementation, this would come from the message
	var priority uint8
	switch eventType {
	case 1: // Accident
		priority = 8
	case 2: // Roadworks
		priority = 5
	case 3: // Weather
		priority = 6
	case 4: // Hazard
		priority = 7
	case 5: // Traffic
		priority = 4
	default:
		priority = 3
	}
	rsa.Priority = priority
	
	// Create timestamp based on the parsed Unix timestamp
	messageTime := time.Unix(int64(timestamp), 0)
	
	// Create a V2X message
	v2xMessage := &models.V2XMessage{
		Protocol:    models.ProtocolCV2XMode4,
		MessageType: models.MessageTypeDENM,
		RawData:     append([]byte{}, data...), // Make a copy to avoid data races
		Timestamp:   messageTime,
		ReceivedAt:  time.Now(),
		SourceID:    fmt.Sprintf("DENM-%d", eventID),
		Latitude:    latitude,
		Longitude:   longitude,
		RSSI:        int16(-70 + rand.Intn(20)), // Simulated RSSI
	}
	
	// Determine interface type string and PLMN info
	interfaceTypeStr := "PC5"
	plmnInfo := ""
	
	// Add PLMN info if it's a Uu (network) interface
	if interfaceType == InterfaceTypeUu {
		interfaceTypeStr = "Uu"
		plmnInfo = "310-410" // Example PLMN ID (simulated)
	}
	
	// Create CV2X specific info
	cv2xInfo := &models.CV2XMessage{
		InterfaceType: interfaceTypeStr,
		PLMNInfo:      plmnInfo,
	}
	
	return rsa, cv2xInfo, v2xMessage, nil
}

// ParseCAM parses a Cooperative Awareness Message (European equivalent to BSM)
func (p *CV2XParser) ParseCAM(data []byte, interfaceType InterfaceType) (*models.BasicSafetyMessage, *models.CV2XMessage, *models.V2XMessage, error) {
	// CAM is structurally similar to BSM but with European standards
	// For simplicity in this implementation, we'll adapt it to our BSM model
	
	if len(data) < 24 {
		return nil, nil, nil, fmt.Errorf("CAM data too short, expected at least 24 bytes, got %d", len(data))
	}
	
	// Skip the message type and interface type bytes
	data = data[2:]
	
	bsm := &models.BasicSafetyMessage{}
	
	// Parse station ID (4 bytes)
	bsm.TemporaryID = binary.BigEndian.Uint32(data[0:4])
	
	// Parse message count (1 byte)
	bsm.MessageCount = uint8(data[4])
	
	// Parse timestamp (4 bytes)
	timestamp := binary.BigEndian.Uint32(data[5:9])
	
	// Parse position (8 bytes each for latitude and longitude)
	latitude := math.Float64frombits(binary.BigEndian.Uint64(data[9:17]))
	longitude := math.Float64frombits(binary.BigEndian.Uint64(data[17:25]))
	
	// Parse speed and heading if available
	var speed float32 = 0
	var heading float32 = 0
	
	if len(data) >= 33 {
		speed = math.Float32frombits(binary.BigEndian.Uint32(data[25:29]))
		heading = math.Float32frombits(binary.BigEndian.Uint32(data[29:33]))
	}
	
	bsm.Speed = speed
	bsm.Heading = heading
	
	// Create timestamp based on the parsed Unix timestamp
	messageTime := time.Unix(int64(timestamp), 0)
	
	// Create a V2X message
	v2xMessage := &models.V2XMessage{
		Protocol:    models.ProtocolCV2XMode4,
		MessageType: models.MessageTypeCAM,
		RawData:     append([]byte{}, data...), // Make a copy to avoid data races
		Timestamp:   messageTime,
		ReceivedAt:  time.Now(),
		SourceID:    fmt.Sprintf("CAM-%08X", bsm.TemporaryID),
		Latitude:    latitude,
		Longitude:   longitude,
		RSSI:        int16(-72 + rand.Intn(25)), // Simulated RSSI
	}
	
	// Determine interface type string and PLMN info
	interfaceTypeStr := "PC5"
	plmnInfo := ""
	
	// Add PLMN info if it's a Uu (network) interface
	if interfaceType == InterfaceTypeUu {
		interfaceTypeStr = "Uu"
		plmnInfo = "310-410" // Example PLMN ID (simulated)
	}
	
	// Create CV2X specific info
	cv2xInfo := &models.CV2XMessage{
		InterfaceType: interfaceTypeStr,
		PLMNInfo:      plmnInfo,
	}
	
	return bsm, cv2xInfo, v2xMessage, nil
}

// ParseCPM parses a Collective Perception Message
func (p *CV2XParser) ParseCPM(data []byte, interfaceType InterfaceType) (*models.CV2XMessage, *models.V2XMessage, error) {
	if len(data) < 20 {
		return nil, nil, fmt.Errorf("CPM data too short, expected at least 20 bytes, got %d", len(data))
	}
	
	// Skip the message type and interface type bytes
	data = data[2:]
	
	// For CPM, we don't have a specific model, so we'll just extract basic info
	// and store it in CV2XMessage
	
	// Parse station ID (4 bytes)
	stationID := binary.BigEndian.Uint32(data[0:4])
	
	// Parse timestamp (4 bytes)
	timestamp := binary.BigEndian.Uint32(data[4:8])
	
	// Parse position (8 bytes each for latitude and longitude)
	latitude := math.Float64frombits(binary.BigEndian.Uint64(data[8:16]))
	longitude := math.Float64frombits(binary.BigEndian.Uint64(data[16:24]))
	
	// Create timestamp based on the parsed Unix timestamp
	messageTime := time.Unix(int64(timestamp), 0)
	
	// Create a V2X message
	v2xMessage := &models.V2XMessage{
		Protocol:    models.ProtocolCV2XMode4,
		MessageType: models.MessageTypeCPM,
		RawData:     append([]byte{}, data...), // Make a copy to avoid data races
		Timestamp:   messageTime,
		ReceivedAt:  time.Now(),
		SourceID:    fmt.Sprintf("CPM-%08X", stationID),
		Latitude:    latitude,
		Longitude:   longitude,
		RSSI:        int16(-68 + rand.Intn(15)), // Simulated RSSI
	}
	
	// Determine interface type string and PLMN info
	interfaceTypeStr := "PC5"
	plmnInfo := ""
	
	// Add PLMN info if it's a Uu (network) interface
	if interfaceType == InterfaceTypeUu {
		interfaceTypeStr = "Uu"
		plmnInfo = "310-410" // Example PLMN ID (simulated)
	}
	
	// Create CV2X specific info with additional info for CPM
	cv2xInfo := &models.CV2XMessage{
		InterfaceType: interfaceTypeStr,
		QoSInfo:       1, // Higher QoS class for perception data
		PLMNInfo:      plmnInfo,
	}
	
	return cv2xInfo, v2xMessage, nil
}