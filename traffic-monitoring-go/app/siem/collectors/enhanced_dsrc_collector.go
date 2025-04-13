package collectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
	"traffic-monitoring-go/app/siem/v2x"
	"traffic-monitoring-go/app/siem/elasticsearch"
)

// EnhancedDSRCCollector collects events from DSRC (dedicated short range communications)
type EnhancedDSRCCollector struct {
	*BaseCollector
	Port              int
	Interface         string
	listener          net.PacketConn
	j2735Parser       *J2735Parser
	securityVerifier  *v2x.V2XSecurityVerifier
	anomalyDetector   *v2x.V2XAnomalyDetector
	esService		  *elasticsearch.Service
}

// Ensure EnhancedDSRCCollector implements CollectorInterface
var _ CollectorInterface = (*EnhancedDSRCCollector)(nil)

// NewEnhancedDSRCCollector creates a new enhanced DSRC collector
func NewEnhancedDSRCCollector(db *gorm.DB, port int, esService *elasticsearch.Service) *EnhancedDSRCCollector {
	return &EnhancedDSRCCollector{
		BaseCollector:     NewBaseCollector(db),
		Port:              port,
		Interface:         "0.0.0.0", // Listen on all interfaces
		j2735Parser:       NewJ2735Parser(),
		securityVerifier:  v2x.NewV2XSecurityVerifier(db),
		anomalyDetector:   v2x.NewV2XAnomalyDetector(db),
		esService:	  	   esService,
	}
}

// Name returns the collector's name
func (c *EnhancedDSRCCollector) Name() string {
	return "enhanced-dsrc"
}

// Start begins listening for DSRC messages
func (c *EnhancedDSRCCollector) Start(ctx context.Context) error {
	if c.Running {
		return fmt.Errorf("enhanced DSRC collector is already running")
	}

	var err error
	// In real implementations we would interface with a DSRC radio
	// In our demo we will use UDP to simulate DSRC messages
	c.listener, err = net.ListenPacket("udp", fmt.Sprintf("%s:%d", c.Interface, c.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on UDP port %d: %v", c.Port, err)
	}

	c.Running = true
	log.Printf("Enhanced DSRC collector started on UDP port %d", c.Port)
	
	// Start processing in a goroutine
	go func() {
		buffer := make([]byte, 2048) // DSRC messages are typically small
		for {
			select {
			case <-c.StopChan:
				log.Println("Enhanced DSRC collector received stop signal")
				return
			case <-ctx.Done():
				log.Println("Enhanced DSRC collector context canceled")
				return
			default:
				// Set a read deadline to allow checking for the stop signal
				if err := c.listener.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
					log.Printf("Error setting read deadline: %v", err)
					continue
				}

				// Read a packet
				n, addr, err := c.listener.ReadFrom(buffer)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						// Timeout is expected when no data is received
						continue
					}
					log.Printf("Error reading DSRC message: %v", err)
					continue
				}

				// Process the received message
				message := make([]byte, n) // Create a copy of the buffer to avoid data races
				copy(message, buffer[:n])
				log.Printf("Received DSRC message: %d bytes from %s", n, addr.String())

				// Process DSRC message
				go c.processDSRCMessage(message, addr.String())
			}
		}
	}()

	return nil
}

// Stop ends the collection process
func (c *EnhancedDSRCCollector) Stop() error {
	if !c.Running {
		return fmt.Errorf("enhanced DSRC collector is not running")
	}

	c.StopChan <- struct{}{}
	if c.listener != nil {
		c.listener.Close()
	}
	c.Running = false
	log.Println("Enhanced DSRC collector stopped")
	return nil
}

// processDSRCMessage processes a received DSRC message
func (c *EnhancedDSRCCollector) processDSRCMessage(message []byte, sourceAddr string) {
	if len(message) < 2 {
		log.Printf("Message too short to be a valid DSRC message")
		return
	}

	// Extract message type
	messageType, err := c.j2735Parser.ParseMessageType(message)
	if err != nil {
		log.Printf("Error parsing message type: %v", err)
		return
	}

	// Based on the message type, process accordingly
	var v2xMessage *models.V2XMessage
	var securityEvent map[string]interface{}

	switch messageType {
	case MessageTypeBSM:
		bsm, v2xMsg, err := c.j2735Parser.ParseBSM(message)
		if err != nil {
			log.Printf("Error parsing BSM: %v", err)
			return
		}
		v2xMessage = v2xMsg

		// Save BSM to database
		if err := c.DB.Create(v2xMessage).Error; err != nil {
			log.Printf("Error saving V2X message: %v", err)
			return
		}

		bsm.V2XMessageID = v2xMessage.ID
		if err := c.DB.Create(bsm).Error; err != nil {
			log.Printf("Error saving BSM: %v", err)
			return
		}

		// Create security event
		securityEvent = map[string]interface{}{
			"message_type": "bsm",
			"vehicle_id": fmt.Sprintf("%08X", bsm.TemporaryID),
			"position": map[string]interface{}{
				"latitude": v2xMessage.Latitude,
				"longitude": v2xMessage.Longitude,
			},
			"speed": bsm.Speed,
			"heading": bsm.Heading,
		}

	case MessageTypeSPAT:
		spat, v2xMsg, err := c.j2735Parser.ParseSPAT(message)
		if err != nil {
			log.Printf("Error parsing SPAT: %v", err)
			return
		}
		v2xMessage = v2xMsg

		// Save SPAT to database
		if err := c.DB.Create(v2xMessage).Error; err != nil {
			log.Printf("Error saving V2X message: %v", err)
			return
		}

		spat.V2XMessageID = v2xMessage.ID
		if err := c.DB.Create(spat).Error; err != nil {
			log.Printf("Error saving SPAT: %v", err)
			return
		}

		// Save phase states
		for i := range spat.PhaseStates {
			spat.PhaseStates[i].SPATMessageID = spat.ID
			if err := c.DB.Create(&spat.PhaseStates[i]).Error; err != nil {
				log.Printf("Error saving phase state: %v", err)
			}
		}

		// Create security event
		securityEvent = map[string]interface{}{
			"message_type": "spat",
			"intersection_id": spat.IntersectionID,
			"phase_count": len(spat.PhaseStates),
		}

	case MessageTypeRSA:
		rsa, v2xMsg, err := c.j2735Parser.ParseRSA(message)
		if err != nil {
			log.Printf("Error parsing RSA: %v", err)
			return
		}
		v2xMessage = v2xMsg

		// Save RSA to database
		if err := c.DB.Create(v2xMessage).Error; err != nil {
			log.Printf("Error saving V2X message: %v", err)
			return
		}

		rsa.V2XMessageID = v2xMessage.ID
		if err := c.DB.Create(rsa).Error; err != nil {
			log.Printf("Error saving RSA: %v", err)
			return
		}

		// Create security event
		securityEvent = map[string]interface{}{
			"message_type": "rsa",
			"alert_type": rsa.AlertType,
			"description": rsa.Description,
			"priority": rsa.Priority,
			"radius": rsa.Radius,
			"duration": rsa.Duration,
		}

	default:
		// For unknown message types, just log and create a generic security event
		log.Printf("Unknown DSRC message type: %d", messageType)
		
		// Create a generic V2X message record
		v2xMessage = &models.V2XMessage{
			Protocol:    models.ProtocolDSRC,
			MessageType: "unknown",
			RawData:     message,
			Timestamp:   time.Now(),
			ReceivedAt:  time.Now(),
			SourceID:    fmt.Sprintf("UNKNOWN-%d", messageType),
		}

		if err := c.DB.Create(v2xMessage).Error; err != nil {
			log.Printf("Error saving unknown V2X message: %v", err)
			return
		}

		// Create security event
		securityEvent = map[string]interface{}{
			"message_type": "unknown",
			"type_id": messageType,
		}
	}

	// Verify security if we have a valid V2X message
	if v2xMessage != nil && v2xMessage.ID > 0 {
		securityInfo, err := c.securityVerifier.VerifyMessage(v2xMessage)
		if err != nil {
			log.Printf("Error verifying message security: %v", err)
		} else if securityInfo != nil {
			// Add security info to the security event
			securityEvent["signature_valid"] = securityInfo.SignatureValid
			securityEvent["trust_level"] = securityInfo.TrustLevel
			securityEvent["certificate_id"] = securityInfo.CertificateID
			
			// Check for validation error
			if securityInfo.ValidationError != "" {
				securityEvent["validation_error"] = securityInfo.ValidationError
			}
		}
		
		// Check for anomalies
		anomalies, err := c.anomalyDetector.DetectAnomalies(v2xMessage)
		if err != nil {
			log.Printf("Error detecting anomalies: %v", err)
		} else if len(anomalies) > 0 {
			// Add anomaly info to the security event
			anomalyDetails := make([]map[string]interface{}, len(anomalies))
			for i, anomaly := range anomalies {
				anomalyDetails[i] = map[string]interface{}{
					"type": anomaly.AnomalyType,
					"confidence": anomaly.ConfidenceScore,
					"description": anomaly.Description,
				}
			}
			securityEvent["anomalies"] = anomalyDetails
			
			// For high-confidence anomalies, increase severity
			hasHighConfidenceAnomaly := false
			for _, anomaly := range anomalies {
				if anomaly.ConfidenceScore > 0.8 {
					hasHighConfidenceAnomaly = true
					break
				}
			}
			
			if hasHighConfidenceAnomaly {
				c.createSecurityEvent(message, sourceAddr, "High-confidence anomaly detected in DSRC message", 
					models.SeverityHigh, securityEvent)
				return
			}
		}
		
		// Create a normal security event with appropriate message and severity
		var eventMessage string
		var severity models.EventSeverity
		
		switch messageType {
		case MessageTypeBSM:
			eventMessage = fmt.Sprintf("DSRC BSM received from vehicle ID %s", securityEvent["vehicle_id"])
			severity = models.SeverityInfo
		case MessageTypeSPAT:
			eventMessage = fmt.Sprintf("DSRC SPAT message received from intersection %v", securityEvent["intersection_id"])
			severity = models.SeverityInfo
		case MessageTypeRSA:
			eventMessage = fmt.Sprintf("DSRC Roadside Alert received: %s", securityEvent["description"])
			priority := securityEvent["priority"].(uint8)
			if priority >= 7 {
				severity = models.SeverityHigh
			} else if priority >= 4 {
				severity = models.SeverityMedium
			} else {
				severity = models.SeverityLow
			}
		default:
			eventMessage = fmt.Sprintf("Unknown DSRC message type %d received", messageType)
			severity = models.SeverityInfo
		}
		
		c.createSecurityEvent(message, sourceAddr, eventMessage, severity, securityEvent)

		if c.esService != nil {
			go func(msg *models.V2XMessage) {
				if err := c.esService.IndexV2XMessage(msg); err != nil {
					log.Printf("Warning: Failed to index V2X message to Elasticsearch: %v", err)
				}
			}(v2xMessage)
		}
	}
}

// createSecurityEvent creates a security event from a DSRC message
func (c *EnhancedDSRCCollector) createSecurityEvent(message []byte, sourceAddr string, eventMessage string, severity models.EventSeverity, details map[string]interface{}) {
	// Parse the source IP from the address
	srcIP, _, err := net.SplitHostPort(sourceAddr)
	if err != nil {
		srcIP = sourceAddr // fallback to using the full address
	}

	// Create a raw event from the DSRC message
	rawEvent := struct {
		SourceName string                 `json:"source_name"`
		SourceType string                 `json:"source_type"`
		Timestamp  time.Time              `json:"timestamp"`
		Severity   string                 `json:"severity"`
		Category   string                 `json:"category"`
		Message    string                 `json:"message"`
		Details    map[string]interface{} `json:"details"`
	}{
		SourceName: "dsrc",
		SourceType: string(models.SourceTypeVehicle),
		Timestamp:  time.Now(),
		Severity:   string(severity),
		Category:   string(models.CategoryV2X),
		Message:    eventMessage,
		Details:    details,
	}

	// Add common fields
	rawEvent.Details["source_ip"] = srcIP
	rawEvent.Details["raw_message_len"] = len(message)
	rawEvent.Details["protocol"] = "DSRC"

	// Convert to JSON for ingestion
	eventJSON, err := json.Marshal(rawEvent)
	if err != nil {
		log.Printf("Error marshaling DSRC event: %v", err)
		return
	}

	// Ingest the event
	err = c.EventIngester.IngestEvent(eventJSON)
	if err != nil {
		log.Printf("Error ingesting DSRC event: %v", err)
		return
	}

	log.Printf("Processed DSRC message from %s with severity %s", sourceAddr, severity)
}

