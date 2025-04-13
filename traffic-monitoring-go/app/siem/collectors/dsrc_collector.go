package collectors

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
	
	
)

// DSRCCollector collects events from DSRC (dedicated short range communications)D
type DSRCCollector struct {
	*BaseCollector
	Port		int
	Interface	string
	listener	net.PacketConn
	bsmProcessor	*BSMProcessor
}


// BSMProcessor processes Basic Safety Messages
type BSMProcessor struct {
	DB *gorm.DB
}

// NewBSMProcessor creates a new BSM processor
func NewBSMProcessor(db *gorm.DB) *BSMProcessor {
	return &BSMProcessor{
		DB:	db,
	}
}

// Ensure DSRCCollector implements CollectorInterface
var _ CollectorInterface = (*DSRCCollector)(nil)


// NewDSRCCollector creates a new DSRCCollector
func NewDSRCCollector(db *gorm.DB, port int) *DSRCCollector {
	return &DSRCCollector{
		BaseCollector: NewBaseCollector(db),
		Port:          port,
		Interface:     "0.0.0.0", // Listen on all interfaces
		bsmProcessor:  NewBSMProcessor(db),
	}
}


// Name returns the collector's name
func (c *DSRCCollector) Name() string {
	return "dsrc"
}


// Start begins listening for DSRC messages
func (c *DSRCCollector) Start(ctx context.Context) error {
	if c.Running {
		return fmt.Errorf("DSRC collector is already running")
	}

	var err error 
	//in real implementations we would interface with a DSRC radio
	// in our demo we will use UDP to simulate DSRC messages

	c.listener, err = net.ListenPacket("udp", fmt.Sprintf("%s:%d", c.Interface, c.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on UDP port %d: %v", c.Port, err)
	}

	c.Running = true
	log.Printf("DSRC collector started on UDP port %d: %v", c.Port, err)
	
	// start processing i a goroutine
	go func() {
		buffer := make([]byte, 2048) // DSRC messages are typically small
		for {
			select {
			case <-c.StopChan:
				log.Println("DSRC collector received stop signal")
				return
			case <-ctx.Done():
				log.Println("DSRC collector context cancelled")
				return
			default:
				// set a read deeadline to allow checking for the stop signal
				if err := c.listener.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
					log.Printf("Error setting read deadline: %v", err)
					continue
				}

				// read a packet
				n, addr, err := c.listener.ReadFrom(buffer)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						// timeout is expected when no data is received
						continue
					}
					log.Printf("Error reading DSRC messagE: %v", err)
					continue
				}

				// Process the received message
				message := buffer[:n]
				log.Printf("Received DSRC message: %d bytes from %s", n, addr.String())

				// Process DSRC message
				go c.processDSRCMessage(message, addr.String())
			}
		}
	}()

	return nil
}

// Stop ends the collection process
func (c *DSRCCollector) Stop() error {
	if !c.Running {
		return fmt.Errorf("DSRC collector is not running")
	}

	c.StopChan <- struct{}{}
	if c.listener != nil {
		c.listener.Close()
	}
	c.Running = false
	log.Println("DSRC collector stopped")
	return nil
}


// processDSRCMessage processes a received DSRC message
func (c *DSRCCollector) processDSRCMessage(message []byte, sourceAddr string) {
	if len(message) < 2 {
		log.Printf("Message too short to be a valid DSRC message")
		return
	}

	// Extract message type from first byte (simplified protocol)
	messageType := message[0]

	// Based on the message type, process accordingly
	switch messageType {
	case 20: // BSM message type per J2735
		c.processBSM(message[1:], sourceAddr)
	case 19: // MAP message
		c.processMAP(message[1:], sourceAddr)
	case 13: // SPAT message
		c.processSPAT(message[1:], sourceAddr)
	case 31: // Roadside Alert
		c.processRSA(message[1:], sourceAddr)
	default:
		// Create a generic V2X security event
		c.createSecurityEvent(message, sourceAddr, fmt.Sprintf("Unknown DSRC message type: %d", messageType))
	}
}

// processBSM processes a Basic Safety Message
func (c *DSRCCollector) processBSM(message []byte, sourceAddr string) {
	// In a real implementation, decode according to J2735 standard
	// This is a simplified version for demonstration

	// Parse the source IP from the address
	srcIP, _, err := net.SplitHostPort(sourceAddr)
	if err != nil {
		srcIP = sourceAddr // fallback to using the full address
	}

	// Simplified BSM parsing (in reality, use a proper J2735 decoder)
	var temporaryID uint32
	if len(message) >= 4 {
		temporaryID = binary.BigEndian.Uint32(message[0:4])
	}

	// Extract other fields as needed

	// Store in the database
	v2xMessage := models.V2XMessage{
		Protocol:    models.ProtocolDSRC,
		MessageType: models.MessageTypeBSM,
		RawData:     message,
		Timestamp:   time.Now(),
		ReceivedAt:  time.Now(),
		SourceID:    fmt.Sprintf("VEH-%08X", temporaryID),
		// In a real implementation, extract lat/lon from the message
	}

	if err := c.DB.Create(&v2xMessage).Error; err != nil {
		log.Printf("Error saving V2X message: %v", err)
		return
	}

	// Create BSM record
	bsm := models.BasicSafetyMessage{
		V2XMessageID:  v2xMessage.ID,
		TemporaryID:   temporaryID,
		// Populate other fields from parsed data
	}

	if err := c.DB.Create(&bsm).Error; err != nil {
		log.Printf("Error saving BSM: %v", err)
		return
	}

	// Create a security event for SIEM
	c.createSecurityEvent(message, sourceAddr, fmt.Sprintf("DSRC BSM received from vehicle ID %08X (IP: %s)", temporaryID, srcIP))
}

// processMAP processes a MAP message
func (c *DSRCCollector) processMAP(message []byte, sourceAddr string) {
	// Implementation for MAP messages
	c.createSecurityEvent(message, sourceAddr, "DSRC MAP message received")
}

// processSPAT processes a Signal Phase and Timing message
func (c *DSRCCollector) processSPAT(message []byte, sourceAddr string) {
	// Implementation for SPAT messages
	c.createSecurityEvent(message, sourceAddr, "DSRC SPAT message received")
}

// processRSA processes a Roadside Alert message
func (c *DSRCCollector) processRSA(message []byte, sourceAddr string) {
	// Implementation for RSA messages
	c.createSecurityEvent(message, sourceAddr, "DSRC Roadside Alert received")
}

// createSecurityEvent creates a generic security event from a DSRC message
func (c *DSRCCollector) createSecurityEvent(message []byte, sourceAddr string, eventMessage string) {
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
		Severity:   string(models.SeverityInfo),
		Category:   string(models.CategoryV2X),
		Message:    eventMessage,
		Details: map[string]interface{}{
			"source_ip":       srcIP,
			"raw_message_len": len(message),
			"protocol":        "DSRC",
		},
	}

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

	log.Printf("Processed DSRC message from %s", sourceAddr)
}





















