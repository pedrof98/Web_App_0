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
	
)

// CV2XCollector collects events from Cellular V2X
type CV2XCollector struct {
	*BaseCollector
	Port          int
	Interface     string
	listener      net.PacketConn
}

// Ensure CV2XCollector implements CollectorInterface
var _ CollectorInterface = (*CV2XCollector)(nil)

// NewCV2XCollector creates a new CV2XCollector
func NewCV2XCollector(db *gorm.DB, port int) *CV2XCollector {
	return &CV2XCollector{
		BaseCollector: NewBaseCollector(db),
		Port:          port,
		Interface:     "0.0.0.0", // Listen on all interfaces
	}
}

// Name returns the collector's name
func (c *CV2XCollector) Name() string {
	return "cv2x"
}

// Start begins listening for C-V2X messages
func (c *CV2XCollector) Start(ctx context.Context) error {
	if c.Running {
		return fmt.Errorf("C-V2X collector is already running")
	}

	var err error
	// In a real implementation, you'd interface with a C-V2X modem
	// For development, we're using UDP to simulate C-V2X messages
	c.listener, err = net.ListenPacket("udp", fmt.Sprintf("%s:%d", c.Interface, c.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on UDP port %d: %v", c.Port, err)
	}

	c.Running = true
	log.Printf("C-V2X collector started on UDP port %d", c.Port)

	// Start processing in a goroutine
	go func() {
		buffer := make([]byte, 2048) // C-V2X messages are typically small
		for {
			select {
			case <-c.StopChan:
				log.Println("C-V2X collector received stop signal")
				return
			case <-ctx.Done():
				log.Println("C-V2X collector context canceled")
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
				message := buffer[:n]
				log.Printf("Received C-V2X message: %d bytes from %s", n, addr.String())

				// Process the C-V2X message
				go c.processCV2XMessage(message, addr.String())
			}
		}
	}()

	return nil
}

// Stop ends the collection process
func (c *CV2XCollector) Stop() error {
	if !c.Running {
		return fmt.Errorf("C-V2X collector is not running")
	}

	c.StopChan <- struct{}{}
	if c.listener != nil {
		c.listener.Close()
	}
	c.Running = false
	log.Println("C-V2X collector stopped")
	return nil
}

// processCV2XMessage processes a received C-V2X message
func (c *CV2XCollector) processCV2XMessage(message []byte, sourceAddr string) {
	if len(message) < 2 {
		log.Printf("Message too short to be a valid C-V2X message")
		return
	}

	// Extract message type from first byte (simplified protocol)
	messageType := message[0]
	interfaceType := "PC5" // Default to PC5 (direct)
	if message[1]&0x80 != 0 {
		interfaceType = "Uu" // Network-based
	}

	// Based on the message type, process accordingly
	switch messageType {
	case 1: // C-V2X BSM (similar to DSRC BSM but may have different format)
		c.processCV2XBSM(message[2:], sourceAddr, interfaceType)
	case 2: // CAM (Cooperative Awareness Message - European equivalent to BSM)
		c.processCAM(message[2:], sourceAddr, interfaceType)
	case 3: // DENM (Decentralized Environmental Notification Message)
		c.processDENM(message[2:], sourceAddr, interfaceType)
	case 4: // CPM (Collective Perception Message)
		c.processCPM(message[2:], sourceAddr, interfaceType)
	default:
		// Create a generic V2X security event
		c.createSecurityEvent(message, sourceAddr, fmt.Sprintf("Unknown C-V2X message type: %d", messageType), interfaceType)
	}
}

// processCV2XBSM processes a C-V2X Basic Safety Message
func (c *CV2XCollector) processCV2XBSM(message []byte, sourceAddr string, interfaceType string) {
	// Parse the source IP from the address
	srcIP, _, err := net.SplitHostPort(sourceAddr)
	if err != nil {
		srcIP = sourceAddr // fallback to using the full address
	}

	// In a real implementation, this would parse according to C-V2X standards
	// For now, we'll use a simplified structure

	// Save to database
	v2xMessage := models.V2XMessage{
		Protocol:    models.ProtocolCV2XMode4,
		MessageType: models.MessageTypeCV2XBSM,
		RawData:     message,
		Timestamp:   time.Now(),
		ReceivedAt:  time.Now(),
		SourceID:    fmt.Sprintf("CV2X-%s", srcIP),
		// In a real implementation, extract lat/lon from the message
	}

	if err := c.DB.Create(&v2xMessage).Error; err != nil {
		log.Printf("Error saving V2X message: %v", err)
		return
	}

	// Create C-V2X specific info
	cv2xInfo := models.CV2XMessage{
		V2XMessageID:  v2xMessage.ID,
		InterfaceType: interfaceType,
		// Other C-V2X specific fields
	}

	if err := c.DB.Create(&cv2xInfo).Error; err != nil {
		log.Printf("Error saving C-V2X info: %v", err)
		return
	}

	// Create a security event for SIEM
	c.createSecurityEvent(message, sourceAddr, fmt.Sprintf("C-V2X BSM received from %s via %s", srcIP, interfaceType), interfaceType)
}

// processCAM processes a Cooperative Awareness Message
func (c *CV2XCollector) processCAM(message []byte, sourceAddr string, interfaceType string) {
	// Implementation for CAM messages (European equivalent to BSM)
	c.createSecurityEvent(message, sourceAddr, "C-V2X CAM message received", interfaceType)
}

// processDENM processes a Decentralized Environmental Notification Message
func (c *CV2XCollector) processDENM(message []byte, sourceAddr string, interfaceType string) {
	// Implementation for DENM messages (European alert messages)
	c.createSecurityEvent(message, sourceAddr, "C-V2X DENM message received", interfaceType)
}

// processCPM processes a Collective Perception Message
func (c *CV2XCollector) processCPM(message []byte, sourceAddr string, interfaceType string) {
	// Implementation for CPM messages (perception sharing)
	c.createSecurityEvent(message, sourceAddr, "C-V2X CPM message received", interfaceType)
}

// createSecurityEvent creates a generic security event from a C-V2X message
func (c *CV2XCollector) createSecurityEvent(message []byte, sourceAddr string, eventMessage string, interfaceType string) {
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
		Severity:   string(models.SeverityInfo),
		Category:   string(models.CategoryV2X),
		Message:    eventMessage,
		Details: map[string]interface{}{
			"source_ip":       srcIP,
			"raw_message_len": len(message),
			"protocol":        "C-V2X",
			"interface_type":  interfaceType,
		},
	}

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

	log.Printf("Processed C-V2X message from %s", sourceAddr)
}
