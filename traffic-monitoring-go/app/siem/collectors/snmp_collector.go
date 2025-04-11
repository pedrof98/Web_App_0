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

// SNMPCollector collects events from SNMP traps
type SNMPCollector struct {
	*BaseCollector
	Port     int
	listener net.PacketConn
}

// Ensure SNMPCollector implements CollectorInterface
var _ CollectorInterface = (*SNMPCollector)(nil)

// NewSNMPCollector creates a new SNMPCollector
func NewSNMPCollector(db *gorm.DB, port int) *SNMPCollector {
	return &SNMPCollector{
		BaseCollector: NewBaseCollector(db),
		Port:         port,
	}
}

// Name returns the collector's name
func (c *SNMPCollector) Name() string {
	return "snmp"
}

// Start begins listening for SNMP traps
func (c *SNMPCollector) Start(ctx context.Context) error {
	if c.Running {
		return fmt.Errorf("SNMP collector is already running")
	}

	var err error
	// Listen for UDP packets on the specified port (default SNMP trap port is 162)
	c.listener, err = net.ListenPacket("udp", fmt.Sprintf(":%d", c.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on UDP port %d: %v", c.Port, err)
	}

	c.Running = true
	log.Printf("SNMP collector started on UDP port %d", c.Port)

	// start processing in a goroutine
	go func() {
		buffer := make([]byte, 65536) // 64KB buffer for each trap
		for {
			select {
			case <-c.StopChan:
				log.Println("SNMP collector received stop signal")
				return
			case <-ctx.Done():
				log.Println("SNMP collector context canceled")
				return
			default:
				// set a read deadline to allow checking for the stop signal
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
					log.Printf("Error reading SNMP trap: %v", err)
					continue
				}

				// process the received trap
				trap := buffer[:n]
				log.Printf("Received SNMP trap: %d bytes from %s", n, addr.String())

				// Parse and process the SNMP trap
				go c.processSNMPTrap(trap, addr.String())
			}
		}
	}()

	return nil
}

// Stop ends the collection process
func (c *SNMPCollector) Stop() error {
	if !c.Running {
		return fmt.Errorf("SNMP collector is not running")
	}

	c.StopChan <- struct{}{}
	if c.listener != nil {
		c.listener.Close()
	}
	c.Running = false
	log.Println("SNMP collector stopped")
	return nil
}

// processSNMPTrap handles a received SNMP trap
func (c *SNMPCollector) processSNMPTrap(trap []byte, sourceAddr string) {
	// Parse the source IP from the address
	srcIP, _, err := net.SplitHostPort(sourceAddr)
	if err != nil {
		srcIP = sourceAddr // fallback to using the full address
	}

	// Create a raw event from the SNMP trap
	rawEvent := struct {
		SourceName string                 `json:"source_name"`
		SourceType string                 `json:"source_type"`
		Timestamp  time.Time              `json:"timestamp"`
		Severity   string                 `json:"severity"`
		Category   string                 `json:"category"`
		Message    string                 `json:"message"`
		Details    map[string]interface{} `json:"details"`
	}{
		SourceName: "snmp",
		SourceType: string(models.SourceTypeNetwork),
		Timestamp:  time.Now(),
		Severity:   string(models.SeverityInfo),
		Category:   string(models.CategoryNetwork),
		Message:    fmt.Sprintf("SNMP trap received from %s", sourceAddr),
		Details: map[string]interface{}{
			"source_ip":  srcIP,
			"raw_length": len(trap),
			"protocol":   "SNMP",
		},
	}

	// Convert to JSON for ingestion
	eventJSON, err := json.Marshal(rawEvent)
	if err != nil {
		log.Printf("Error marshaling SNMP event: %v", err)
		return
	}

	// Ingest the event
	err = c.EventIngester.IngestEvent(eventJSON)
	if err != nil {
		log.Printf("Error ingesting SNMP event: %v", err)
		return
	}

	log.Printf("Processed SNMP trap from %s", sourceAddr)
}