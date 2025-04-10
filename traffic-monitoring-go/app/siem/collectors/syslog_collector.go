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

// SyslogCollector collects events from syslog
type SyslogCollector struct {
	*BaseCollector
	Port     int
	listener net.PacketConn
}

// Ensure SyslogCollector implements CollectorInterface
var _ CollectorInterface = (*SyslogCollector)(nil)

// NewSyslogCollector creates a new SyslogCollector
func NewSyslogCollector(db *gorm.DB, port int) *SyslogCollector {
	return &SyslogCollector{
		BaseCollector: NewBaseCollector(db),
		Port:         port,
	}
}

// Name returns the collector's name
func (c *SyslogCollector) Name() string {
	return "syslog"
}

// Start begins listening for syslog messages
func (c *SyslogCollector) Start(ctx context.Context) error {
	if c.Running {
		return fmt.Errorf("syslog collector is already running")
	}

	var err error
	// listen for UDP packets on the specified port
	c.listener, err = net.ListenPacket("udp", fmt.Sprintf(":%d", c.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on UDP port %d: %v", c.Port, err)
	}

	c.Running = true
	log.Printf("Syslog collector started on UDP port %d", c.Port)

	// start processing in a goroutine
	go func() {
		buffer := make([]byte, 65536) // 64KB buffer for each message
		for {
			select {
			case <-c.StopChan:
				log.Println("Syslog collector received stop signal")
				return
			case <-ctx.Done():
				log.Println("Syslog collector context canceled")
				return
			default:
				// set a read deadline to allow checking for the stop signal
				if err := c.listener.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
					log.Printf("Error setting read deadline: %v", err)
					continue
				}

				// Read a packet
				n, addr, err := c.listener.ReadFrom(buffer)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						//Timeout is expected when no data is received
						continue
					}
					log.Printf("Error reading syslog message: %v", err)
					continue
				}

				// process the received message
				message := buffer[:n]
				log.Printf("Received %d bytes from %s", n, addr.String())

				//parse and process the syslog message
				go c.processSyslogMessage(message, addr.String())
			}
		}
	}()

	return nil
}

// Stop ends the collection process
func (c *SyslogCollector) Stop() error {
	if !c.Running {
		return fmt.Errorf("syslog collector is not running")
	}

	c.StopChan <- struct{}{}
	if c.listener != nil {
		c.listener.Close()
	}
	c.Running = false
	log.Println("Syslog collector stopped")
	return nil
}

// processSyslogMessage handles a received syslog message
func (c *SyslogCollector) processSyslogMessage(message []byte, sourceAddr string) {
	// Parse the source IP from the address
	srcIP, _, err := net.SplitHostPort(sourceAddr)
	if err != nil {
		srcIP = sourceAddr // fallback to using the full address
	}

	// create a raw event from the syslog message
	rawEvent := struct {
		SourceName string                 `json:"source_name"`
		SourceType string                 `json:"source_type"`
		Timestamp  time.Time              `json:"timestamp"`
		Severity   string                 `json:"severity"`
		Category   string                 `json:"category"`
		Message    string                 `json:"message"`
		Details    map[string]interface{} `json:"details"`
	}{
		SourceName: "syslog",
		SourceType: string(models.SourceTypeSystem),
		Timestamp:  time.Now(), // in real implementation, parse from the message
		Severity:   string(models.SeverityInfo),
		Category:   string(models.CategorySystem),
		Message:    string(message),
		Details: map[string]interface{}{
			"source_ip":  srcIP,
			"raw_length": len(message),
		},
	}

	// convert to JSON for ingestion
	eventJSON, err := json.Marshal(rawEvent)
	if err != nil {
		log.Printf("Error marshaling syslog event: %v", err)
		return
	}

	// ingest the event
	err = c.EventIngester.IngestEvent(eventJSON)
	if err != nil {
		log.Printf("Error ingesting syslog event: %v", err)
		return
	}

	log.Printf("Processed syslog message from %s", sourceAddr)
}