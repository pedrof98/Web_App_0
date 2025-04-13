package collectors

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	"traffic-monitoring-go/app/models"
)

//J2735Parser provides a parsing funcitonality for J2735 DSRC messages
type J2735Parser struct{}

// New Parser func to instantiate a new J2735 parser
func NewJ2735Parser() *J2735Parser {
	return &J2735Parser{}
}

// Constants for J2735 message types
const (
	MessageTypeBSM  byte = 20 // Basic Safety Message
	MessageTypeSPAT byte = 19 // Signal Phase and Timing
	MessageTypeMAP  byte = 18 // Map Data
	MessageTypeRSA  byte = 31 // Roadside Alert
	MessageTypeTIM  byte = 31 // Traveler Information Message
)

// ParseMessageType determines the type of J2735 message
func (p *J2735Parser) ParseMessageType(data []byte) (byte, error) {
	if len(data) < 1 {
		return 0, fmt.Errorf("message too short")
	}
	return data[0], nil
}

// ParseBSM parses a Basic Safety Message
func (p *J2735Parser) ParseBSM(data []byte) (*models.BasicSafetyMessage, *models.V2XMessage, error) {
	if len(data) < 20 {
		return nil, nil, fmt.Errorf("BSM data too short, expected at least 20 bytes, got %d", len(data))
	}

	// Skip the message type byte which should be already processed
	data = data[1:]

	bsm := &models.BasicSafetyMessage{}

	// Parse temporary ID (4 bytes)
	bsm.TemporaryID = binary.BigEndian.Uint32(data[0:4])

	// Parse message count (1 byte)
	bsm.MessageCount = uint8(data[4])

	// Parse time mark (2 bytes) - milliseconds within the minute
	secMark := binary.BigEndian.Uint16(data[5:7])
	bsm.SecMark = secMark

	// Parse position (latitude, longitude) (8 bytes)
	// Latitude and longitude are in 1/10 microdegrees
	latitudeRaw := int32(binary.BigEndian.Uint32(data[7:11]))
	longitudeRaw := int32(binary.BigEndian.Uint32(data[11:15]))

	// Convert to degrees
	latitude := float64(latitudeRaw) / 10000000.0
	longitude := float64(longitudeRaw) / 10000000.0

	// Parse speed (2 bytes) - in units of 0.02 m/s
	speedRaw := binary.BigEndian.Uint16(data[15:17])
	speed := float32(speedRaw) * 0.02

	// Parse heading (2 bytes) - in units of 0.0125 degrees
	headingRaw := binary.BigEndian.Uint16(data[17:19])
	heading := float32(headingRaw) * 0.0125

	// Set values
	bsm.Speed = speed
	bsm.Heading = heading

	// Parse acceleration, brake status, and vehicle size if available
	if len(data) >= 30 {
		// Acceleration (lateral, longitudinal, vertical) - in 0.01 m/s^2
		if len(data) >= 25 {
			lateralAccelRaw := int16(binary.BigEndian.Uint16(data[19:21]))
			longAccelRaw := int16(binary.BigEndian.Uint16(data[21:23]))
			bsm.LateralAccel = float32(lateralAccelRaw) * 0.01
			bsm.LongitudinalAccel = float32(longAccelRaw) * 0.01

			// Yaw rate - in 0.01 degrees per second
			if len(data) >= 25 {
				yawRateRaw := int16(binary.BigEndian.Uint16(data[23:25]))
				bsm.YawRate = float32(yawRateRaw) * 0.01
			}
		}

		// Brake status (1 byte bit flags)
		if len(data) >= 26 {
			brakeStatus := data[25]
			bsm.BrakeApplied = (brakeStatus & 0x01) != 0
			bsm.TractionControl = (brakeStatus & 0x02) != 0
			bsm.ABS = (brakeStatus & 0x04) != 0
			bsm.StabilityControl = (brakeStatus & 0x08) != 0
			bsm.BrakeBoost = (brakeStatus & 0x10) != 0
			bsm.AuxiliaryBrakes = (brakeStatus & 0x20) != 0
		}

		// Vehicle size (width, length) - in centimeters
		if len(data) >= 30 {
			widthRaw := binary.BigEndian.Uint16(data[26:28])
			lengthRaw := binary.BigEndian.Uint16(data[28:30])
			bsm.Width = float32(widthRaw) / 100.0  // Convert to meters
			bsm.Length = float32(lengthRaw) / 100.0 // Convert to meters
		}
	}

	// Create a timestamp based on current time, but adjust seconds to match secMark
	now := time.Now()
	seconds := secMark / 1000
	milliseconds := secMark % 1000

	// Create a v2x message with the parsed data
	v2xMessage := &models.V2XMessage{
		Protocol:    models.ProtocolDSRC,
		MessageType: models.MessageTypeBSM,
		RawData:     data,
		Timestamp:   time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), int(seconds), int(milliseconds)*1000000, now.Location()),
		ReceivedAt:  now,
		SourceID:    fmt.Sprintf("VEH-%08X", bsm.TemporaryID),
		Latitude:    latitude,
		Longitude:   longitude,
	}

	// TODO: In a real system, calculate RSSI from received signal
	v2xMessage.RSSI = int16(-75 + rand.Intn(30)) // Simulated RSSI between -75 and -45 dBm

	return bsm, v2xMessage, nil
}

// ParseSPAT parses a Signal Phase and Timing message
func (p *J2735Parser) ParseSPAT(data []byte) (*models.SignalPhaseAndTiming, *models.V2XMessage, error) {
	if len(data) < 8 {
		return nil, nil, fmt.Errorf("SPAT data too short, expected at least 8 bytes, got %d", len(data))
	}

	// Skip the message type byte which should be already processed
	data = data[1:]

	spat := &models.SignalPhaseAndTiming{}

	// Parse intersection ID (4 bytes)
	spat.IntersectionID = binary.BigEndian.Uint32(data[0:4])

	// Parse message count (1 byte)
	spat.MsgCount = uint8(data[4])

	// Parse number of phases (1 byte)
	phaseCount := uint8(data[5])

	// Create a timestamp for the message
	now := time.Now()

	// Create a v2x message with the parsed data
	v2xMessage := &models.V2XMessage{
		Protocol:    models.ProtocolDSRC,
		MessageType: models.MessageTypeSPAT,
		RawData:     data,
		Timestamp:   now,
		ReceivedAt:  now,
		SourceID:    fmt.Sprintf("RSU-%d", spat.IntersectionID),
	}

	// TODO: In a real system, extract RSU location from the message or lookup
	// For now we'll set a default or simulated location
	v2xMessage.Latitude = 37.7749 + (float64(spat.IntersectionID % 100) * 0.001)
	v2xMessage.Longitude = -122.4194 + (float64(spat.IntersectionID % 100) * 0.001)
	v2xMessage.RSSI = int16(-70 + rand.Intn(20)) // Simulated RSSI between -70 and -50 dBm

	// Parse phase states
	phaseStates := make([]models.PhaseState, 0, phaseCount)

	// Each phase has a phase ID (1 byte), light state (1 byte), and 3 timing values (2 bytes each)
	// Total 7 bytes per phase
	expectedDataLength := 6 + (int(phaseCount) * 7)
	if len(data) < expectedDataLength {
		return nil, v2xMessage, fmt.Errorf("SPAT data too short for specified number of phases")
	}

	offset := 6 // Start parsing phases after the header
	for i := uint8(0); i < phaseCount; i++ {
		phaseID := uint8(data[offset])
		lightState := uint8(data[offset+1])
		startTime := binary.BigEndian.Uint16(data[offset+2:offset+4])
		minEndTime := binary.BigEndian.Uint16(data[offset+4:offset+6])
		maxEndTime := binary.BigEndian.Uint16(data[offset+6:offset+8])

		// Convert light state number to string
		lightStateStr := ""
		switch lightState {
		case 0:
			lightStateStr = "red"
		case 1:
			lightStateStr = "yellow"
		case 2:
			lightStateStr = "green"
		default:
			lightStateStr = "unknown"
		}

		phaseState := models.PhaseState{
			PhaseID:     phaseID,
			LightState:  lightStateStr,
			StartTime:   startTime,
			MinEndTime:  minEndTime,
			MaxEndTime:  maxEndTime,
		}

		phaseStates = append(phaseStates, phaseState)
		offset += 7
	}

	spat.PhaseStates = phaseStates

	return spat, v2xMessage, nil
}

// ParseRSA parses a Roadside Alert message
func (p *J2735Parser) ParseRSA(data []byte) (*models.RoadsideAlert, *models.V2XMessage, error) {
	if len(data) < 15 {
		return nil, nil, fmt.Errorf("RSA data too short, expected at least 15 bytes, got %d", len(data))
	}

	// Skip the message type byte which should be already processed
	data = data[1:]

	rsa := &models.RoadsideAlert{}

	// Parse alert type (2 bytes)
	rsa.AlertType = binary.BigEndian.Uint16(data[0:2])

	// Parse description if available (simulated by next few bytes)
	description := ""
	switch rsa.AlertType {
	case 1:
		description = "Accident ahead"
	case 2:
		description = "Road work"
	case 3:
		description = "Weather condition"
	case 4:
		description = "Road hazard"
	default:
		description = fmt.Sprintf("Alert type %d", rsa.AlertType)
	}
	rsa.Description = description

	// Parse priority (1 byte)
	rsa.Priority = uint8(data[2])

	// Parse location (8 bytes - latitude and longitude)
	latitudeRaw := int32(binary.BigEndian.Uint32(data[3:7]))
	longitudeRaw := int32(binary.BigEndian.Uint32(data[7:11]))

	// Convert to degrees
	latitude := float64(latitudeRaw) / 10000000.0
	longitude := float64(longitudeRaw) / 10000000.0

	// Parse radius and duration
	rsa.Radius = binary.BigEndian.Uint16(data[11:13])
	rsa.Duration = binary.BigEndian.Uint16(data[13:15])

	// Create a timestamp for the message
	now := time.Now()

	// Create a v2x message with the parsed data
	v2xMessage := &models.V2XMessage{
		Protocol:    models.ProtocolDSRC,
		MessageType: models.MessageTypeRSA,
		RawData:     data,
		Timestamp:   now,
		ReceivedAt:  now,
		SourceID:    fmt.Sprintf("RSA-%d", rsa.AlertType),
		Latitude:    latitude,
		Longitude:   longitude,
		RSSI:        int16(-65 + rand.Intn(20)), // Simulated RSSI between -65 and -45 dBm
	}

	return rsa, v2xMessage, nil
}
