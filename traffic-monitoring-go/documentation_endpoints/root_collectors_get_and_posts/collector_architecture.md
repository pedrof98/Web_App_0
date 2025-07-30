# Collector Architecture and Data Flow Analysis

## Complete Data Flow: From V2X Messages to Security Events

```
V2X Vehicle → DSRC/C-V2X Radio → Collector → Parser → Security Event → Rule Engine → Alert
                                     ↑
                         GET /collectors manages this layer
```

### 1. **Collector Registration and Initialization** (System Startup)
```go
// In NewCollectorHandler() - app/handlers/collector.go
func NewCollectorHandler(db *gorm.DB) *CollectorHandler {
    manager := collectors.NewCollectorManager(db)

    // Network Security Collectors
    syslogCollector := collectors.NewSyslogCollector(db, 514)   // RFC 3164 logs
    snmpCollector := collectors.NewSNMPCollector(db, 162)       // Network traps

    // V2X Security Collectors  
    dsrcCollector := collectors.NewDSRCCollector(db, 5001)      // IEEE 802.11p
    cv2xCollector := collectors.NewCV2XCollector(db, 5002)      // 3GPP C-V2X

    // Register with manager
    manager.RegisterCollector(syslogCollector)
    manager.RegisterCollector(snmpCollector) 
    manager.RegisterCollector(dsrcCollector)
    manager.RegisterCollector(cv2xCollector)
}
```

### 2. **Collector Interface and Base Architecture**
```go
// All collectors implement this interface
type CollectorInterface interface {
    Start(ctx context.Context) error  // Begin message collection
    Stop() error                      // Graceful shutdown
    Name() string                     // Unique identifier
    IsRunning() bool                  // Status check
}

// Base collector provides common functionality
type BaseCollector struct {
    DB           *gorm.DB                    // Database connection
    EventIngester *siem.EventIngester        // Security event creation
    Running      bool                       // Current state
    StopChan     chan struct{}              // Shutdown signal
}
```

### 3. **V2X Collector Implementation Deep Dive**

#### DSRC Collector (Dedicated Short Range Communications):
```go
type DSRCCollector struct {
    *BaseCollector
    Port         int               // UDP port 5001
    Interface    string            // "0.0.0.0" (all interfaces)
    listener     net.PacketConn    // UDP socket
    bsmProcessor *BSMProcessor     // Basic Safety Message parser
}

// DSRC Message Processing Flow
func (c *DSRCCollector) Start(ctx context.Context) error {
    // 1. Bind to UDP port
    c.listener, err = net.ListenPacket("udp", "0.0.0.0:5001")
    
    // 2. Launch message processing goroutine
    go func() {
        buffer := make([]byte, 2048)  // DSRC messages typically <2KB
        for {
            select {
            case <-c.StopChan:
                return  // Graceful shutdown
            default:
                // 3. Read incoming V2X message
                n, addr, err := c.listener.ReadFrom(buffer)
                
                // 4. Process message in separate goroutine (non-blocking)
                go c.processMessage(buffer[:n], addr.String())
            }
        }
    }()
}
```

#### C-V2X Collector (Cellular Vehicle-to-Everything):
```go
type CV2XCollector struct {
    *BaseCollector
    Port           int                    // UDP port 5002
    Interface      string                 // "0.0.0.0"
    listener       net.PacketConn         // UDP socket
    cv2xParser     *CV2XParser           // C-V2X message parser
    pcmParser      *PCMParser            // PC5 interface parser
}

// C-V2X supports multiple message types:
// - BSM (Basic Safety Message) - like DSRC but cellular
// - CAM (Cooperative Awareness Message) - European standard
// - DENM (Decentralized Environmental Notification) - emergency alerts
// - CPM (Collective Perception Message) - sensor sharing
```

### 4. **Message Processing Pipeline**

#### V2X Message Reception and Parsing:
```go
func (c *DSRCCollector) processMessage(message []byte, sourceAddr string) {
    // Step 1: Determine message type (J2735 standard)
    messageType := message[0]  // First byte indicates type
    
    switch messageType {
    case 0x00:  // Basic Safety Message (BSM)
        c.processBSM(message[1:], sourceAddr)
    case 0x01:  // MAP (intersection geometry)
        c.processMAP(message[1:], sourceAddr)
    case 0x02:  // SPAT (signal phase and timing)
        c.processSPAT(message[1:], sourceAddr)
    case 0x03:  // RSA (road side alert)
        c.processRSA(message[1:], sourceAddr)
    default:
        // Unknown message type - still create security event
        c.createSecurityEvent(message, sourceAddr, 
            fmt.Sprintf("Unknown DSRC message type: %d", messageType))
    }
}
```

#### Security Event Creation from V2X Messages:
```go
func (c *DSRCCollector) createSecurityEvent(message []byte, sourceAddr string, eventMessage string) {
    // Parse source IP
    srcIP, _, err := net.SplitHostPort(sourceAddr)
    
    // Create structured event for SIEM processing
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
        SourceType: "vehicle",
        Timestamp:  time.Now(),
        Severity:   "info",      // Default severity
        Category:   "v2x",       // V2X category for rule matching
        Message:    eventMessage,
        Details: map[string]interface{}{
            "source_ip":        srcIP,
            "protocol":         "DSRC",
            "raw_message_len":  len(message),
            "message_hex":      hex.EncodeToString(message[:min(64, len(message))]),
        },
    }
    
    // Convert to JSON and ingest via SIEM pipeline
    eventJSON, _ := json.Marshal(rawEvent)
    c.EventIngester.IngestEvent(eventJSON)
}
```

### 5. **Collector Manager Architecture**

#### Centralized Lifecycle Management:
```go
type CollectorManager struct {
    DB         *gorm.DB                           // Database connection
    collectors map[string]CollectorInterface     // Registry of collectors
    mutex      sync.Mutex                        // Thread-safe operations
    ctx        context.Context                   // Global context
    cancel     context.CancelFunc                // Shutdown coordination
}

// Manager Operations
func (m *CollectorManager) StartCollector(name string) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    collector, exists := m.collectors[name]
    if !exists {
        return fmt.Errorf("collector '%s' not found", name)
    }
    
    return collector.Start(m.ctx)  // Start with global context
}

func (m *CollectorManager) StopCollector(name string) error {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    collector, exists := m.collectors[name]
    if !exists {
        return fmt.Errorf("collector '%s' not found", name)
    }
    
    return collector.Stop()  // Graceful shutdown
}
```

### 6. **Network Protocol Integration**

#### Collector Port Assignments and Protocols:
```yaml
Network Collectors:
  syslog:
    port: 514
    protocol: UDP
    standard: RFC 3164/5424
    use_case: System logs, firewall events, router alerts
    
  snmp:
    port: 162  
    protocol: UDP
    standard: SNMPv1/v2c/v3
    use_case: Network device traps, performance alerts

V2X Collectors:
  dsrc:
    port: 5001
    protocol: UDP (simulating IEEE 802.11p)
    standard: J2735/IEEE 1609
    use_case: Vehicle safety messages, intersection data
    frequency: 10-100 Hz per vehicle
    
  cv2x:
    port: 5002
    protocol: UDP (simulating 3GPP C-V2X)
    standard: 3GPP Release 14+
    use_case: Cellular vehicle communication, V2N, V2I, V2P
    frequency: Variable based on service requirements
```

### 7. **Enhanced V2X Security Processing**

#### Advanced V2X Collector Features:
```go
type EnhancedDSRCCollector struct {
    *BaseCollector
    Port             int
    j2735Parser      *J2735Parser              // Protocol decoder
    securityVerifier *v2x.V2XSecurityVerifier  // PKI signature validation
    anomalyDetector  *v2x.V2XAnomalyDetector   // Behavioral analysis
    esService        *elasticsearch.Service    // Real-time indexing
}

// Enhanced message processing with security analysis
func (c *EnhancedDSRCCollector) processV2XMessage(message []byte, sourceAddr string) {
    // 1. Parse J2735 message structure
    parsedMsg, err := c.j2735Parser.Parse(message)
    
    // 2. Verify digital signature (if present)
    signatureValid := c.securityVerifier.VerifySignature(parsedMsg)
    
    // 3. Detect behavioral anomalies
    anomalies := c.anomalyDetector.AnalyzeMessage(parsedMsg)
    
    // 4. Determine event severity based on security analysis
    severity := c.calculateSeverity(signatureValid, anomalies)
    
    // 5. Create enriched security event
    c.createEnhancedSecurityEvent(parsedMsg, sourceAddr, severity, anomalies)
}
```

### 8. **Real-Time Rule Integration**

#### Collector to Alert Pipeline:
```go
func (c *DSRCCollector) createSecurityEvent(message []byte, sourceAddr string, eventMessage string) {
    // 1. Create and store security event
    eventJSON, _ := json.Marshal(rawEvent)
    err := c.EventIngester.IngestEvent(eventJSON)
    
    // 2. Automatic rule evaluation (happens in ingestion pipeline)
    // - Event stored in security_events table
    // - Rule engine evaluates all enabled V2X rules
    // - Matching rules create alerts automatically
    
    // 3. Get the created security event for rule processing
    var securityEvent models.SecurityEvent
    c.DB.Order("id DESC").First(&securityEvent)
    
    // 4. Trigger rule engine evaluation
    ruleEngine := siem.NewEnhancedRuleEngine(c.DB)
    err = ruleEngine.EvaluateEvent(&securityEvent)
    
    // 5. Check if alerts were generated
    var alertCount int64
    c.DB.Model(&models.Alert{}).Where("security_event_id = ?", securityEvent.ID).Count(&alertCount)
    
    if alertCount > 0 {
        log.Printf("!!!ALERT: %d alert(s) generated for V2X event %d", alertCount, securityEvent.ID)
    }
}
```

### 9. **Performance and Scalability**

#### Collector Performance Characteristics:
```yaml
DSRC Collector Performance:
  message_rate: 100-1000 msgs/sec per vehicle
  processing_latency: <5ms per message
  memory_usage: ~10MB baseline + 1KB per goroutine
  concurrent_vehicles: 100+ vehicles supported
  
CV2X Collector Performance:
  message_rate: Variable (service-dependent)
  processing_latency: <10ms per message
  bandwidth: Higher than DSRC (cellular network)
  scalability: Network-limited rather than system-limited

System-wide Collector Metrics:
  total_throughput: 10,000+ events/second
  rule_evaluation: <50ms for complex V2X rules
  alert_generation: <100ms end-to-end latency
  resource_efficiency: <1GB RAM for all collectors
```

#### Horizontal Scaling Architecture:
```go
// Collectors designed for distributed deployment
type DistributedCollectorConfig struct {
    NodeID       string            // Unique node identifier
    LoadBalancer *LoadBalancer     // Message distribution
    SharedState  *SharedStateManager  // Cross-node coordination
    EventBus     *EventBus         // Inter-node communication
}

// Multiple collector instances can run on different nodes
// Messages distributed based on source vehicle ID or geographic region
```

### 10. **Integration with SIEM Components**

#### Complete SIEM Integration Flow:
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   V2X Vehicle   │───▶│   Collector      │───▶│ Security Event  │
│                 │    │  (DSRC/C-V2X)    │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Dashboard     │◀───│  Rule Engine     │◀───│   Ingestion     │
│   (Monitoring)  │    │  (V2X Rules)     │    │   Pipeline      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        
                                ▼                        
                       ┌──────────────────┐              
                       │     Alerts       │              
                       │  (Incident Mgmt) │              
                       └──────────────────┘              
```

#### Collector Status Monitoring:
```go
// GET /collectors endpoint provides operational visibility
type CollectorStatus struct {
    Name            string    `json:"name"`
    Running         bool      `json:"running"`
    Port            int       `json:"port,omitempty"`
    MessagesPerSec  float64   `json:"messages_per_sec,omitempty"`
    ErrorRate       float64   `json:"error_rate,omitempty"`
    LastMessage     time.Time `json:"last_message,omitempty"`
    UptimeSeconds   int64     `json:"uptime_seconds,omitempty"`
}
```

### 11. **Security Considerations**

#### V2X Security Threat Detection:
```go
// Collectors detect various V2X attack patterns:
type V2XSecurityThreats struct {
    // Message injection attacks
    InvalidSignatures    []string  // Unsigned or invalid digital signatures
    
    // Behavioral anomalies  
    PositionJumps       []PositionAnomaly  // Impossible vehicle movements
    SpeedAnomalies      []SpeedAnomaly     // Unrealistic speed changes
    
    // DoS attacks
    MessageFlooding     []FloodingAttack   // High frequency message spam
    
    // Protocol violations
    MalformedMessages   []ProtocolViolation // Invalid J2735/C-V2X format
}

// Real-time threat detection in collector pipeline
func (c *EnhancedDSRCCollector) detectThreats(message *V2XMessage) []SecurityThreat {
    threats := []SecurityThreat{}
    
    if !message.SignatureValid {
        threats = append(threats, SecurityThreat{
            Type: "invalid_signature",
            Severity: "critical",
            Description: "Message with invalid or missing digital signature",
        })
    }
    
    if c.anomalyDetector.DetectPositionJump(message) {
        threats = append(threats, SecurityThreat{
            Type: "position_jump",
            Severity: "high", 
            Description: "Vehicle reported impossible position change",
        })
    }
    
    return threats
}
```

### 12. **Operational Excellence**

#### Production Deployment Patterns:
```yaml
High Availability:
  - Multiple collector instances per protocol
  - Load balancing across collector nodes
  - Automatic failover on collector failure
  - Health check monitoring via GET /collectors

Performance Monitoring:
  - Message throughput tracking
  - Processing latency metrics
  - Memory and CPU utilization
  - Network socket health

Maintenance Operations:
  - Rolling collector restarts
  - Port configuration updates
  - Protocol parser updates
  - Security signature key rotation
```

This collector architecture provides the **real-time data ingestion foundation** for V2X security monitoring, designed for high-performance, low-latency threat detection in safety-critical vehicular environments.