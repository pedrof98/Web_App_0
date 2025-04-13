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

// EnhancedCV2XCollector collects events from Cellular V2X
type EnhancedCV2XCollector struct {
	*BaseCollector
	Port              int
	Interface         string
	listener          net.PacketConn
	cv2xParser        *CV2XParser
	securityVerifier  *v2x.V2XSecurityVerifier
	anomalyDetector   *v2x.V2XAnomalyDetector
	esService		  *elasticsearch.Service
}

// Ensure EnhancedCV2XCollector implements CollectorInterface
var _ CollectorInterface = (*EnhancedCV2XCollector)(nil)

// NewEnhancedCV2XCollector creates a new enhanced C-V2X collector
func NewEnhancedCV2XCollector(db *gorm.DB, port int, esService *elasticsearch.Service) *EnhancedCV2XCollector {
	return &EnhancedCV2XCollector{
		BaseCollector:     NewBaseCollector(db),
		Port:              port,
		Interface:         "0.0.0.0", // Listen on all interfaces
		cv2xParser:        NewCV2XParser(),
		securityVerifier:  v2x.NewV2XSecurityVerifier(db),
		anomalyDetector:   v2x.NewV2XAnomalyDetector(db),
		esService:         esService,
	}
}

// Name returns the collector's name
func (c *EnhancedCV2XCollector) Name() string {
	return "enhanced-cv2x"
}

// Start begins listening for C-V2X messages
func (c *EnhancedCV2XCollector) Start(ctx context.Context) error {
	if c.Running {
		return fmt.Errorf("enhanced C-V2X collector is already running")
	}

	var err error
	// In a real implementation, you'd interface with a C-V2X modem
	// For development, we're using UDP to simulate C-V2X messages
	c.listener, err = net.ListenPacket("udp", fmt.Sprintf("%s:%d", c.Interface, c.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on UDP port %d: %v", c.Port, err)
	}

	c.Running = true
	log.Printf("Enhanced C-V2X collector started on UDP port %d", c.Port)

	// Start processing in a goroutine
	go func() {
		buffer := make([]byte, 2048) // C-V2X messages are typically small
		for {
			select {
			case <-c.StopChan:
				log.Println("Enhanced C-V2X collector received stop signal")
				return
			case <-ctx.Done():
				log.Println("Enhanced C-V2X collector context canceled")
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
					log.Printf("Error reading C-V2X message: %v", err)
					continue
				}

				// Process the received message
				message := make([]byte, n) // Create a copy of the buffer to avoid data races
				copy(message, buffer[:n])
				log.Printf("Received C-V2X message: %d bytes from %s", n, addr.String())

				// Process the C-V2X message
				go c.processCV2XMessage(message, addr.String())
			}
		}
	}()

	return nil
}

// Stop ends the collection process
func (c *EnhancedCV2XCollector) Stop() error {
	if !c.Running {
		return fmt.Errorf("enhanced C-V2X collector is not running")
	}

	c.StopChan <- struct{}{}
	if c.listener != nil {
		c.listener.Close()
	}
	c.Running = false
	log.Println("Enhanced C-V2X collector stopped")
	return nil
}

// processCV2XMessage processes a received C-V2X message
func (c *EnhancedCV2XCollector) processCV2XMessage(message []byte, sourceAddr string) {
	if len(message) < 2 {
		log.Printf("Message too short to be a valid C-V2X message")
		return
	}

	// Extract message type and interface type
	messageType, interfaceType, err := c.cv2xParser.ParseMessageType(message)
	if err != nil {
		log.Printf("Error parsing message type: %v", err)
		return
	}

	// Based on the message type, process accordingly
	var v2xMessage *models.V2XMessage
	var cv2xInfo *models.CV2XMessage
	var securityEvent map[string]interface{}
	var interfaceTypeStr string

	// Convert interface type to string for logging
	if interfaceType == InterfaceTypePC5 {
		interfaceTypeStr = "PC5"
	} else {
		interfaceTypeStr = "Uu"
	}

	switch messageType {
	case CV2XMessageTypeBSM:
		bsm, cv2xSpecific, v2xMsg, err := c.cv2xParser.ParseCV2XBSM(message, interfaceType)
		if err != nil {
			log.Printf("Error parsing C-V2X BSM: %v", err)
			return
		}
		v2xMessage = v2xMsg
		cv2xInfo = cv2xSpecific

		// Save to database
		if err := c.DB.Create(v2xMessage).Error; err != nil {
			log.Printf("Error saving V2X message: %v", err)
			return
		}

		bsm.V2XMessageID = v2xMessage.ID
		if err := c.DB.Create(bsm).Error; err != nil {
			log.Printf("Error saving BSM: %v", err)
			return
		}

		cv2xInfo.V2XMessageID = v2xMessage.ID
		if err := c.DB.Create(cv2xInfo).Error; err != nil {
			log.Printf("Error saving C-V2X info: %v", err)
			return
		}

		// Create security event
		securityEvent = map[string]interface{}{
			"message_type": "cv2x_bsm",
			"vehicle_id": fmt.Sprintf("%08X", bsm.TemporaryID),
			"position": map[string]interface{}{
				"latitude": v2xMessage.Latitude,
				"longitude": v2xMessage.Longitude,
			},
			"speed": bsm.Speed,
			"heading": bsm.Heading,
			"interface_type": interfaceTypeStr,
		}

	case CV2XMessageTypeCAM:
		cam, cv2xSpecific, v2xMsg, err := c.cv2xParser.ParseCAM(message, interfaceType)
		if err != nil {
			log.Printf("Error parsing C-V2X CAM: %v", err)
			return
		}
		v2xMessage = v2xMsg
		cv2xInfo = cv2xSpecific

		// Save to database
		if err := c.DB.Create(v2xMessage).Error; err != nil {
			log.Printf("Error saving V2X message: %v", err)
			return
		}

		cam.V2XMessageID = v2xMessage.ID
		if err := c.DB.Create(cam).Error; err != nil {
			log.Printf("Error saving CAM (as BSM): %v", err)
			return
		}

		cv2xInfo.V2XMessageID = v2xMessage.ID
		if err := c.DB.Create(cv2xInfo).Error; err != nil {
			log.Printf("Error saving C-V2X info: %v", err)
			return
		}

		// Create security event
		securityEvent = map[string]interface{}{
			"message_type": "cam",
			"vehicle_id": fmt.Sprintf("%08X", cam.TemporaryID),
			"position": map[string]interface{}{
				"latitude": v2xMessage.Latitude,
				"longitude": v2xMessage.Longitude,
			},
			"speed": cam.Speed,
			"heading": cam.Heading,
			"interface_type": interfaceTypeStr,
		}

	case CV2XMessageTypeDENM:
		denm, cv2xSpecific, v2xMsg, err := c.cv2xParser.ParseDENM(message, interfaceType)
		if err != nil {
			log.Printf("Error parsing C-V2X DENM: %v", err)
			return
		}
		v2xMessage = v2xMsg
		cv2xInfo = cv2xSpecific

		// Save to database
		if err := c.DB.Create(v2xMessage).Error; err != nil {
			log.Printf("Error saving V2X message: %v", err)
			return
		}

		denm.V2XMessageID = v2xMessage.ID
		if err := c.DB.Create(denm).Error; err != nil {
			log.Printf("Error saving DENM (as RSA): %v", err)
			return
		}

		cv2xInfo.V2XMessageID = v2xMessage.ID
		if err := c.DB.Create(cv2xInfo).Error; err != nil {
			log.Printf("Error saving C-V2X info: %v", err)
			return
		}

		// Create security event
		securityEvent = map[string]interface{}{
			"message_type": "denm",
			"alert_type": denm.AlertType,
			"description": denm.Description,
			"priority": denm.Priority,
			"radius": denm.Radius,
			"duration": denm.Duration,
			"position": map[string]interface{}{
				"latitude": v2xMessage.Latitude,
				"longitude": v2xMessage.Longitude,
			},
			"interface_type": interfaceTypeStr,
		}

	case CV2XMessageTypeCPM:
		cv2xSpecific, v2xMsg, err := c.cv2xParser.ParseCPM(message, interfaceType)
		if err != nil {
			log.Printf("Error parsing C-V2X CPM: %v", err)
			return
		}
		v2xMessage = v2xMsg
		cv2xInfo = cv2xSpecific

		// Save to database
		if err := c.DB.Create(v2xMessage).Error; err != nil {
			log.Printf("Error saving V2X message: %v", err)
			return
		}

		cv2xInfo.V2XMessageID = v2xMessage.ID
		if err := c.DB.Create(cv2xInfo).Error; err != nil {
			log.Printf("Error saving C-V2X info: %v", err)
			return
		}

		// Create security event
		securityEvent = map[string]interface{}{
			"message_type": "cpm",
			"source_id": v2xMessage.SourceID,
			"position": map[string]interface{}{
				"latitude": v2xMessage.Latitude,
				"longitude": v2xMessage.Longitude,
			},
			"interface_type": interfaceTypeStr,
		}

	default:
		// Unknown message type
		log.Printf("Unknown C-V2X message type: %d", messageType)
		
		// Create a generic V2X message record
		v2xMessage = &models.V2XMessage{
			Protocol:    models.ProtocolCV2XMode4,
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

		// Create a generic C-V2X record
		cv2xInfo = &models.CV2XMessage{
			V2XMessageID:  v2xMessage.ID,
			InterfaceType: interfaceTypeStr,
			
		}

		if err := c.DB.Create(cv2xInfo).Error; err != nil {
			log.Printf("Error saving unknown C-V2X info: %v", err)
			return
		}

		// Create security event
		securityEvent = map[string]interface{}{
			"message_type": "unknown",
			"type_id": messageType,
			"interface_type": interfaceTypeStr,
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
				c.createSecurityEvent(message, sourceAddr, "High-confidence anomaly detected in C-V2X message", 
					models.SeverityHigh, securityEvent)
				return
			}
		}
		
		// Create a normal security event with appropriate message and severity
		var eventMessage string
		var severity models.EventSeverity
		
		switch messageType {
		case CV2XMessageTypeBSM:
			eventMessage = fmt.Sprintf("C-V2X BSM received from vehicle ID %s via %s interface", 
				securityEvent["vehicle_id"], interfaceTypeStr)
			severity = models.SeverityInfo
		case CV2XMessageTypeCAM:
			eventMessage = fmt.Sprintf("C-V2X CAM received from vehicle ID %s via %s interface", 
				securityEvent["vehicle_id"], interfaceTypeStr)
			severity = models.SeverityInfo
		case CV2XMessageTypeDENM:
			eventMessage = fmt.Sprintf("C-V2X DENM received: %s via %s interface", 
				securityEvent["description"], interfaceTypeStr)
			
			priority, ok := securityEvent["priority"].(uint8)
			if ok {
				if priority >= 7 {
					severity = models.SeverityHigh
				} else if priority >= 4 {
					severity = models.SeverityMedium
				} else {
					severity = models.SeverityLow
				}
			} else {
				severity = models.SeverityInfo
			}
		case CV2XMessageTypeCPM:
			eventMessage = fmt.Sprintf("C-V2X CPM received from %s via %s interface", 
				v2xMessage.SourceID, interfaceTypeStr)
			severity = models.SeverityInfo
		default:
			eventMessage = fmt.Sprintf("Unknown C-V2X message type %d received via %s interface", 
				messageType, interfaceTypeStr)
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

// createSecurityEvent creates a security event from a C-V2X message
func (c *EnhancedCV2XCollector) createSecurityEvent(message []byte, sourceAddr string, eventMessage string, severity models.EventSeverity, details map[string]interface{}) {
	// Parse the source IP from the address
	srcIP, _, err := net.SplitHostPort(sourceAddr)
	if err != nil {
		srcIP = sourceAddr // fallback to using the full address
	}

	// Create a raw event from the C-V2X message
	rawEvent := struct {
		SourceName string                 `json:"source_name"`
		SourceType string                 `json:"source_type"`
		Timestamp  time.Time              `json:"timestamp"`
		Severity   string                 `json:"severity"`
		Category   string                 `json:"category"`
		Message    string                 `json:"message"`
		Details    map[string]interface{} `json:"details"`
	}{
		SourceName: "cv2x",
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
	rawEvent.Details["protocol"] = "C-V2X"

	// Convert to JSON for ingestion
	eventJSON, err := json.Marshal(rawEvent)
	if err != nil {
		log.Printf("Error marshaling C-V2X event: %v", err)
		return
	}

	// Ingest the event
	err = c.EventIngester.IngestEvent(eventJSON)
	if err != nil {
		log.Printf("Error ingesting C-V2X event: %v", err)
		return
	}

	log.Printf("Processed C-V2X message from %s with severity %s", sourceAddr, severity)
}