// File: internal/siem/v2x/message_parser.go
package v2x

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"
)

// BSMData represents the core data in a Basic Safety Message
type BSMData struct {
	VehicleID    string
	MessageCount uint32
	Timestamp    time.Time
	Position     Position
	Speed        float64 // km/h
	Heading      float64 // degrees (0-359)
	Acceleration float64 // m/s²
	BrakeStatus  bool
	VehicleSize  VehicleSize
	VehicleType  VehicleType
}

type VehicleSize struct {
	Width  float64 // meters
	Length float64 // meters
	Height float64 // meters
}

type VehicleType string

const (
	VehicleTypePassengerCar VehicleType = "passenger_car"
	VehicleTypeTruck        VehicleType = "truck"
	VehicleTypeBus          VehicleType = "bus"
	VehicleTypeMotorcycle   VehicleType = "motorcycle"
	VehicleTypeEmergency    VehicleType = "emergency"
	VehicleTypeUnknown      VehicleType = "unknown"
)

// CAMData represents Cooperative Awareness Message data
type CAMData struct {
	StationID      uint32
	GenerationTime time.Time
	Position       Position
	Speed          float64
	Heading        float64
	StationType    StationType
	VehicleRole    VehicleRole
	ExteriorLights ExteriorLights
}

type StationType string

const (
	StationTypePassengerCar StationType = "passenger_car"
	StationTypeCyclist      StationType = "cyclist"
	StationTypeMoped        StationType = "moped"
	StationTypeMotorcycle   StationType = "motorcycle"
	StationTypePedestrian   StationType = "pedestrian"
	StationTypeRoadSideUnit StationType = "road_side_unit"
)

type VehicleRole string

const (
	VehicleRoleDefault          VehicleRole = "default"
	VehicleRolePublicTransport  VehicleRole = "public_transport"
	VehicleRoleSpecialTransport VehicleRole = "special_transport"
	VehicleRoleDangerousGoods   VehicleRole = "dangerous_goods"
	VehicleRoleRoadWork         VehicleRole = "road_work"
	VehicleRoleRescue           VehicleRole = "rescue"
	VehicleRoleEmergency        VehicleRole = "emergency"
)

type ExteriorLights struct {
	LowBeamHeadlights    bool
	HighBeamHeadlights   bool
	LeftTurnSignal       bool
	RightTurnSignal      bool
	DaytimeRunningLights bool
	ReverseLights        bool
	FogLights            bool
	ParkingLights        bool
}

// MessageParser handles parsing of different V2X message formats
type MessageParser struct {
	// Configuration for parsing different message standards
	enableJ2735 bool // SAE J2735 (US standard)
	enableETSI  bool // ETSI EN 302 (European standard)
}

// ParsedMessage represents a successfully parsed V2X message
type ParsedMessage struct {
	MessageType V2XMessageType
	BSMData     *BSMData
	CAMData     *CAMData
	RawBytes    []byte
	ParseTime   time.Time
	Valid       bool
	Errors      []string
}

func NewMessageParser() *MessageParser {
	return &MessageParser{
		enableJ2735: true,
		enableETSI:  true,
	}
}

// ParseMessage attempts to parse raw message bytes into structured data
func (p *MessageParser) ParseMessage(rawData []byte) (*ParsedMessage, error) {
	parsed := &ParsedMessage{
		RawBytes:  rawData,
		ParseTime: time.Now(),
		Valid:     false,
		Errors:    make([]string, 0),
	}

	// Try to determine message type and parse accordingly
	msgType, err := p.detectMessageType(rawData)
	if err != nil {
		parsed.Errors = append(parsed.Errors, fmt.Sprintf("message type detection failed: %v", err))
		return parsed, err
	}

	parsed.MessageType = msgType

	switch msgType {
	case BSMMessage:
		bsmData, err := p.parseBSM(rawData)
		if err != nil {
			parsed.Errors = append(parsed.Errors, fmt.Sprintf("BSM parsing failed: %v", err))
			return parsed, err
		}
		parsed.BSMData = bsmData
		parsed.Valid = true

	case CAMMessage:
		camData, err := p.parseCAM(rawData)
		if err != nil {
			parsed.Errors = append(parsed.Errors, fmt.Sprintf("CAM parsing failed: %v", err))
			return parsed, err
		}
		parsed.CAMData = camData
		parsed.Valid = true

	default:
		err := fmt.Errorf("unsupported message type: %s", msgType)
		parsed.Errors = append(parsed.Errors, err.Error())
		return parsed, err
	}

	return parsed, nil
}

// detectMessageType analyzes raw bytes to determine the V2X message type
func (p *MessageParser) detectMessageType(data []byte) (V2XMessageType, error) {
	if len(data) < 4 {
		return "", fmt.Errorf("message too short: %d bytes", len(data))
	}

	// Check for JSON format (simulation/testing)
	if data[0] == '{' {
		return p.detectJSONMessageType(data)
	}

	// Check for ASN.1 PER encoded messages (production)
	return p.detectASN1MessageType(data)
}

// detectJSONMessageType detects message type from JSON format (for testing/simulation)
func (p *MessageParser) detectJSONMessageType(data []byte) (V2XMessageType, error) {
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return "", fmt.Errorf("invalid JSON: %v", err)
	}

	if msgType, exists := temp["messageType"]; exists {
		if typeStr, ok := msgType.(string); ok {
			return V2XMessageType(typeStr), nil
		}
	}

	// Try to infer from content
	if _, exists := temp["coreData"]; exists {
		return BSMMessage, nil
	}
	if _, exists := temp["cam"]; exists {
		return CAMMessage, nil
	}

	return "", fmt.Errorf("could not determine message type from JSON")
}

// detectASN1MessageType detects message type from ASN.1 encoded data
func (p *MessageParser) detectASN1MessageType(data []byte) (V2XMessageType, error) {
	// Simplified ASN.1 detection based on message structure
	// In a real implementation, this would use proper ASN.1 parsing

	// Check message ID in first few bytes
	if len(data) >= 2 {
		msgID := binary.BigEndian.Uint16(data[0:2])
		switch msgID {
		case 0x0014: // BSM message ID in J2735
			return BSMMessage, nil
		case 0x0002: // CAM message ID in ETSI
			return CAMMessage, nil
		}
	}

	// Fallback: assume BSM for unknown format in testing
	return BSMMessage, nil
}

// parseBSM parses Basic Safety Message data
func (p *MessageParser) parseBSM(data []byte) (*BSMData, error) {
	// Try JSON format first (for simulation/testing)
	if data[0] == '{' {
		return p.parseBSMFromJSON(data)
	}

	// Parse ASN.1 format (for production)
	return p.parseBSMFromASN1(data)
}

// parseBSMFromJSON parses BSM from JSON format (simulation/testing)
func (p *MessageParser) parseBSMFromJSON(data []byte) (*BSMData, error) {
	var jsonMsg struct {
		VehicleID    string `json:"vehicleId"`
		MessageCount uint32 `json:"msgCnt"`
		Timestamp    int64  `json:"timestamp"`
		CoreData     struct {
			Position struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
				Elv float64 `json:"elevation"`
			} `json:"position"`
			Speed        float64 `json:"speed"`
			Heading      float64 `json:"heading"`
			Acceleration float64 `json:"accel"`
			BrakeStatus  bool    `json:"brakes"`
		} `json:"coreData"`
		VehicleType string `json:"vehicleType,omitempty"`
		Size        struct {
			Width  float64 `json:"width,omitempty"`
			Length float64 `json:"length,omitempty"`
			Height float64 `json:"height,omitempty"`
		} `json:"size,omitempty"`
	}

	if err := json.Unmarshal(data, &jsonMsg); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %v", err)
	}

	bsm := &BSMData{
		VehicleID:    jsonMsg.VehicleID,
		MessageCount: jsonMsg.MessageCount,
		Timestamp:    time.Unix(jsonMsg.Timestamp/1000, (jsonMsg.Timestamp%1000)*1000000),
		Position: Position{
			Latitude:  jsonMsg.CoreData.Position.Lat,
			Longitude: jsonMsg.CoreData.Position.Lon,
			Elevation: jsonMsg.CoreData.Position.Elv,
		},
		Speed:        jsonMsg.CoreData.Speed,
		Heading:      jsonMsg.CoreData.Heading,
		Acceleration: jsonMsg.CoreData.Acceleration,
		BrakeStatus:  jsonMsg.CoreData.BrakeStatus,
		VehicleSize: VehicleSize{
			Width:  jsonMsg.Size.Width,
			Length: jsonMsg.Size.Length,
			Height: jsonMsg.Size.Height,
		},
		VehicleType: VehicleType(jsonMsg.VehicleType),
	}

	return bsm, nil
}

// parseBSMFromASN1 parses BSM from ASN.1 PER encoded format
func (p *MessageParser) parseBSMFromASN1(data []byte) (*BSMData, error) {
	// Simplified ASN.1 parsing - in production, use proper ASN.1 library
	if len(data) < 20 {
		return nil, fmt.Errorf("BSM data too short: %d bytes", len(data))
	}

	// This is a simplified parser for demonstration
	// Real implementation would use libraries like asn1go or similar

	bsm := &BSMData{
		VehicleID:    fmt.Sprintf("vehicle_%d", binary.BigEndian.Uint32(data[2:6])),
		MessageCount: binary.BigEndian.Uint32(data[6:10]),
		Timestamp:    time.Now(), // Would extract from message
		Position: Position{
			Latitude:  float64(int32(binary.BigEndian.Uint32(data[10:14]))) / 10000000.0,
			Longitude: float64(int32(binary.BigEndian.Uint32(data[14:18]))) / 10000000.0,
			Elevation: 0, // Would extract if available
		},
		Speed:   float64(binary.BigEndian.Uint16(data[18:20])) * 0.02, // Scale factor
		Heading: 0,                                                    // Would extract from remaining bytes
	}

	return bsm, nil
}

// parseCAM parses Cooperative Awareness Message data
func (p *MessageParser) parseCAM(data []byte) (*CAMData, error) {
	// Try JSON format first
	if data[0] == '{' {
		return p.parseCAMFromJSON(data)
	}

	// Parse ASN.1 format
	return p.parseCAMFromASN1(data)
}

// parseCAMFromJSON parses CAM from JSON format
func (p *MessageParser) parseCAMFromJSON(data []byte) (*CAMData, error) {
	var jsonMsg struct {
		StationID      uint32 `json:"stationId"`
		GenerationTime int64  `json:"generationTime"`
		Position       struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"position"`
		Speed       float64 `json:"speed"`
		Heading     float64 `json:"heading"`
		StationType string  `json:"stationType"`
		VehicleRole string  `json:"vehicleRole"`
	}

	if err := json.Unmarshal(data, &jsonMsg); err != nil {
		return nil, fmt.Errorf("CAM JSON parsing failed: %v", err)
	}

	cam := &CAMData{
		StationID:      jsonMsg.StationID,
		GenerationTime: time.Unix(jsonMsg.GenerationTime/1000, 0),
		Position: Position{
			Latitude:  jsonMsg.Position.Lat,
			Longitude: jsonMsg.Position.Lon,
		},
		Speed:       jsonMsg.Speed,
		Heading:     jsonMsg.Heading,
		StationType: StationType(jsonMsg.StationType),
		VehicleRole: VehicleRole(jsonMsg.VehicleRole),
	}

	return cam, nil
}

// parseCAMFromASN1 parses CAM from ASN.1 format
func (p *MessageParser) parseCAMFromASN1(data []byte) (*CAMData, error) {
	// Simplified implementation for demonstration
	if len(data) < 16 {
		return nil, fmt.Errorf("CAM data too short: %d bytes", len(data))
	}

	cam := &CAMData{
		StationID:      binary.BigEndian.Uint32(data[2:6]),
		GenerationTime: time.Now(),
		Position: Position{
			Latitude:  float64(int32(binary.BigEndian.Uint32(data[6:10]))) / 10000000.0,
			Longitude: float64(int32(binary.BigEndian.Uint32(data[10:14]))) / 10000000.0,
		},
		Speed:       float64(binary.BigEndian.Uint16(data[14:16])) * 0.01,
		StationType: StationTypePassengerCar, // Would extract from message
	}

	return cam, nil
}

// ConvertToV2XMessage converts parsed message to detection engine format
func (p *ParsedMessage) ToV2XMessage() *V2XMessage {
	if !p.Valid {
		return nil
	}

	msg := &V2XMessage{
		MessageType: p.MessageType,
		Timestamp:   p.ParseTime,
		RawData:     p.RawBytes,
	}

	if p.BSMData != nil {
		msg.ID = fmt.Sprintf("bsm_%s_%d", p.BSMData.VehicleID, p.BSMData.MessageCount)
		msg.VehicleID = p.BSMData.VehicleID
		msg.Position = p.BSMData.Position
		msg.Speed = p.BSMData.Speed
		msg.Heading = p.BSMData.Heading
		msg.Acceleration = p.BSMData.Acceleration
		msg.MessageCount = p.BSMData.MessageCount
		msg.Timestamp = p.BSMData.Timestamp
	}

	if p.CAMData != nil {
		msg.ID = fmt.Sprintf("cam_%d_%d", p.CAMData.StationID, time.Now().Unix())
		msg.VehicleID = fmt.Sprintf("station_%d", p.CAMData.StationID)
		msg.Position = p.CAMData.Position
		msg.Speed = p.CAMData.Speed
		msg.Heading = p.CAMData.Heading
		msg.Timestamp = p.CAMData.GenerationTime
	}

	return msg
}
