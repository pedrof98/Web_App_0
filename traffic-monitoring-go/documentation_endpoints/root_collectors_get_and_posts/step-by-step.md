# Step-by-Step Debug Process for Collector Management Endpoints

## Prerequisites
1. Understanding that collectors are the data ingestion layer of the SIEM
2. Knowledge that V2X collectors (DSRC/C-V2X) capture vehicle messages
3. Understanding the collector lifecycle: Registration → Start → Running → Stop
4. Network collectors (Syslog/SNMP) handle traditional security data sources

## Debug Session Steps

### Step 1: Understand Collector Architecture
Before debugging, understand the collector system:
```
CollectorManager (registry) → Individual Collectors → Network Listeners → Message Processing → Security Events
```

**Key Collectors Registered**:
- **syslog** (port 514): Traditional log collection
- **snmp** (port 162): Network device traps
- **dsrc** (port 5001): Vehicle DSRC messages
- **cv2x** (port 5002): Cellular V2X messages

### Step 2: Launch Debugger and Test Sequence
1. Use same VS Code debug configuration
2. Set breakpoints from the collectors breakpoints guide
3. Start debugging
4. Follow this test sequence:

#### Test Sequence:
1. **GET /collectors** (check initial status)
2. **POST /collectors/dsrc/start** (start V2X collector)
3. **GET /collectors** (verify status change)
4. **POST /collectors/dsrc/stop** (stop collector)
5. **POST /collectors/start-all** (bulk start)
6. **POST /collectors/stop-all** (bulk stop)

### Step 3: Debug Flow Analysis - GET /collectors

#### Breakpoint 1: Handler Entry
- **Variable to inspect**: `h`, `c`
- **What to check**: CollectorHandler with CollectorManager initialized
- **Expected**: Handler with manager containing 4 registered collectors

#### Breakpoint 2: Collector Names Retrieval
- **Variable to inspect**: `collectorNames`
- **What to check**: Array of registered collector names
- **Expected**: `["syslog", "snmp", "dsrc", "cv2x"]`

#### Breakpoint 3: Status Check Loop
- **Variable to inspect**: `name` (current collector in loop)
- **What to check**: Iteration through each registered collector
- **Expected**: Process each collector sequentially

#### Breakpoint 4: Individual Status Check
- **Variable to inspect**: `status`, `err`
- **What to check**: Running status for current collector
- **Expected**: `status` = true/false, `err` = nil for valid collectors

#### Breakpoint 5: Response Construction
- **Variable to inspect**: `collectors` array being built
- **What to check**: Final response structure
- **Expected**:
  ```json
  [
    {"name": "syslog", "running": false},
    {"name": "snmp", "running": false},
    {"name": "dsrc", "running": false},
    {"name": "cv2x", "running": false}
  ]
  ```

### Step 4: Debug Flow Analysis - POST /collectors/:name/start

#### Breakpoint 6: Start Handler Entry
- **Variable to inspect**: `h`, `c`
- **What to check**: Request to start specific collector
- **Expected**: Valid handler and context

#### Breakpoint 7: Parameter Extraction
- **Variable to inspect**: `name`
- **What to check**: Collector name from URL parameter
- **Expected**: One of: "syslog", "snmp", "dsrc", "cv2x"

#### Breakpoint 8: Manager Start Call
- **Variable to inspect**: `err` from StartCollector
- **What to check**: Manager delegation success/failure
- **Expected**: `err` = nil for successful start

#### Breakpoint 9: Manager Start Implementation
- **Variable to inspect**: `m.collectors`, `name`
- **What to check**: Collector lookup in registry
- **Expected**: Collector exists in map

#### Breakpoint 10: Collector Lookup
- **Variable to inspect**: `collector`, `exists`
- **What to check**: Registry contains requested collector
- **Expected**: `exists` = true, `collector` = valid interface

#### Breakpoint 11: Individual Collector Start
- **Variable to inspect**: `collector.Start(m.ctx)` result
- **What to check**: Actual collector startup
- **Expected**: Collector state changes to Running = true

### Step 5: Deep Collector Analysis

#### Understanding Collector State:
When debugging individual collectors, inspect these key states:

**DSRC Collector State**:
```go
type DSRCCollector struct {
    Running   bool           // false → true on start
    Port      int            // 5001
    Interface string         // "0.0.0.0"
    listener  net.PacketConn // nil → UDP listener
    StopChan  chan struct{}  // Channel for graceful shutdown
}
```

**Collector Lifecycle**:
1. **Registration**: Collector added to manager during handler creation
2. **Start**: UDP listener created, goroutine launched for message processing
3. **Running**: Continuously listens for messages, processes via goroutines
4. **Stop**: Stop signal sent, listener closed, goroutine terminates

### Step 6: Network Layer Analysis

#### Collector Network Behavior:
When collectors start, they bind to specific UDP ports:

```go
// DSRC Collector binds to UDP port 5001
c.listener, err = net.ListenPacket("udp", "0.0.0.0:5001")

// Processing loop in goroutine
go func() {
    buffer := make([]byte, 2048)
    for {
        select {
        case <-c.StopChan:
            return  // Graceful shutdown
        default:
            n, addr, err := c.listener.ReadFrom(buffer)
            // Process received message
        }
    }
}()
```

#### Key Debug Points:
- **Port Binding**: Verify no conflicts with existing processes
- **Goroutine Launch**: Confirm message processing loop starts
- **Message Reception**: Watch for incoming UDP packets
- **Event Generation**: Verify security events created from messages

### Step 7: V2X Message Processing Flow

#### When V2X Collector Receives Message:
```
1. UDP Message Received → DSRC/C-V2X Collector
2. Protocol Parsing → J2735/C-V2X Parser
3. Security Event Creation → EventIngester
4. Rule Evaluation → EnhancedRuleEngine
5. Alert Generation → Alert System
```

#### Debug Message Processing:
```go
// Watch for this flow in DSRC collector
func (c *DSRCCollector) processMessage(message []byte, sourceAddr string) {
    // 1. Parse message type and content
    messageType := message[0]
    
    // 2. Create security event
    c.createSecurityEvent(message, sourceAddr, eventMessage)
    
    // 3. Events automatically trigger rule evaluation
}
```

### Step 8: Error Handling Analysis

#### Common Error Scenarios:

**Port Already in Use**:
- **Debug Point**: `net.ListenPacket("udp", ...)` call
- **Error**: `address already in use`
- **Solution**: Check for existing processes on port

**Collector Already Running**:
- **Debug Point**: `if c.Running` check in Start method
- **Error**: `collector is already running`
- **Solution**: Check collector state before start

**Invalid Collector Name**:
- **Debug Point**: `collector, exists := m.collectors[name]`
- **Error**: `collector 'xyz' not found`
- **Solution**: Use valid collector names

### Step 9: Integration Testing

#### V2X Integration Test:
1. **Start V2X Collectors**: 
   ```
   POST /collectors/dsrc/start
   POST /collectors/cv2x/start
   ```

2. **Send Test V2X Message** (using netcat or similar):
   ```bash
   echo -n "\x01\x00test_bsm_message" | nc -u localhost 5001
   ```

3. **Verify Security Event Created**:
   ```
   GET /security-events?category=v2x
   ```

4. **Check for Generated Alerts**:
   ```
   GET /alerts?category=v2x
   ```

#### Full System Integration:
```
Collector Start → Message Reception → Security Event → Rule Evaluation → Alert → Dashboard Update
```

### Step 10: Performance Monitoring

#### Collector Performance Metrics:
Monitor these aspects during debugging:

**Resource Usage**:
- **Memory**: Each collector uses goroutines and buffers
- **CPU**: Message processing overhead
- **Network**: UDP socket binding and traffic

**Message Throughput**:
- **DSRC**: Designed for high-frequency vehicle messages (10-100 Hz)
- **C-V2X**: Supports higher bandwidth than DSRC
- **Syslog/SNMP**: Lower frequency but larger message sizes

#### Debug Performance Issues:
```go
// Watch for performance bottlenecks
func (c *DSRCCollector) processMessage(message []byte, sourceAddr string) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        if duration > 10*time.Millisecond {
            log.Printf("Slow message processing: %v", duration)
        }
    }()
    // Message processing logic
}
```

## Key Learning Points

### SIEM Collector Architecture:
1. **Centralized Management**: CollectorManager provides unified lifecycle control
2. **Protocol Specialization**: Each collector handles specific communication protocols
3. **Concurrent Processing**: Goroutines enable parallel message handling
4. **Graceful Shutdown**: Stop channels ensure clean collector termination

### V2X Security Context:
1. **Real-time Processing**: Vehicle messages require <50ms processing for safety
2. **High Volume**: DSRC can generate 100+ messages/second per vehicle
3. **Protocol Complexity**: J2735 and C-V2X require specialized parsing
4. **Security Critical**: Invalid messages can represent active attacks

### Network Integration:
1. **UDP-Based**: Most collectors use UDP for performance and simplicity
2. **Port Management**: Each collector needs dedicated port binding
3. **Error Resilience**: Network errors shouldn't crash collectors
4. **Resource Efficiency**: Designed for 24/7 operation

## Common Issues and Debugging

### Issue: Collector Won't Start
- **Check**: Port availability (`netstat -tulpn | grep :5001`)
- **Verify**: No firewall blocking collector ports
- **Solution**: Kill conflicting processes or change ports

### Issue: Messages Received but No Events
- **Check**: EventIngester integration in collector
- **Verify**: Database connectivity for event storage
- **Solution**: Check ingestion pipeline and database logs

### Issue: High Resource Usage
- **Check**: Message processing efficiency
- **Verify**: Goroutine leak prevention
- **Solution**: Monitor via `/benchmark/metrics`

### Issue: V2X Messages Not Triggering Rules
- **Check**: Security event category matches rule conditions
- **Verify**: Rule engine evaluation after event creation
- **Solution**: Debug rule condition matching

## Advanced Debugging Techniques

### Network Packet Analysis:
```bash
# Monitor collector ports for incoming traffic
sudo tcpdump -i any -n port 5001
sudo tcpdump -i any -n port 5002
```

### Collector State Inspection:
```go
// Add debug endpoint to check collector internals
func (h *CollectorHandler) GetCollectorDetails(c *gin.Context) {
    name := c.Param("name")
    collector := h.CollectorManager.GetCollector(name)
    // Return detailed collector state
}
```

### Message Flow Tracing:
```go
// Add tracing to follow message from collector to alert
func (c *DSRCCollector) processMessage(message []byte, sourceAddr string) {
    traceID := generateTraceID()
    log.Printf("TRACE %s: Message received from %s", traceID, sourceAddr)
    // Continue tracing through pipeline
}
```

This collector system provides the **data ingestion foundation** for the entire V2X SIEM - understanding its operation is crucial for effective security monitoring!