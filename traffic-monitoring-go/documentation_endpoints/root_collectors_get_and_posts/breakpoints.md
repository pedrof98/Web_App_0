# Debug Breakpoints Guide for Collector Management Endpoints

## Covered Endpoints
- **GET /collectors** - List all collectors and their status
- **POST /collectors/:name/start** - Start a specific collector
- **POST /collectors/:name/stop** - Stop a specific collector
- **POST /collectors/start-all** - Start all collectors
- **POST /collectors/stop-all** - Stop all collectors

## Critical Breakpoints (Set these in order of execution)

### GET /collectors Endpoint

#### 1. **Collector Handler Entry Point**
**File**: `app/handlers/collector.go`  
**Line**: `func (h *CollectorHandler) GetCollectors(c *gin.Context)`  
**Purpose**: See collector status retrieval request

#### 2. **Collector Names Retrieval**
**File**: `app/handlers/collector.go`  
**Line**: `collectorNames := h.CollectorManager.GetCollectorNames()`  
**Purpose**: Watch how registered collector names are fetched

#### 3. **Status Check Loop**
**File**: `app/handlers/collector.go`  
**Line**: `for _, name := range collectorNames {`  
**Purpose**: See iteration through each collector

#### 4. **Individual Status Check**
**File**: `app/handlers/collector.go`  
**Line**: `status, err := h.CollectorManager.GetCollectorStatus(name)`  
**Purpose**: Watch status retrieval for each collector

#### 5. **Response Construction**
**File**: `app/handlers/collector.go`  
**Line**: `collectors = append(collectors, map[string]interface{}{`  
**Purpose**: See how collector data is assembled for response

### POST /collectors/:name/start Endpoint

#### 6. **Start Collector Handler Entry**
**File**: `app/handlers/collector.go`  
**Line**: `func (h *CollectorHandler) StartCollector(c *gin.Context)`  
**Purpose**: See collector start request processing

#### 7. **Collector Name Parameter Extraction**
**File**: `app/handlers/collector.go`  
**Line**: `name := c.Param("name")`  
**Purpose**: Watch URL parameter extraction

#### 8. **Manager Start Call**
**File**: `app/handlers/collector.go`  
**Line**: `err := h.CollectorManager.StartCollector(name)`  
**Purpose**: See manager delegation for starting

#### 9. **Manager Start Implementation**
**File**: `app/siem/collectors/manager.go`  
**Line**: `func (m *CollectorManager) StartCollector(name string) error {`  
**Purpose**: Watch actual collector startup logic

#### 10. **Collector Lookup**
**File**: `app/siem/collectors/manager.go`  
**Line**: `collector, exists := m.collectors[name]`  
**Purpose**: See how collectors are looked up from registry

#### 11. **Individual Collector Start**
**File**: `app/siem/collectors/manager.go`  
**Line**: `err := collector.Start(m.ctx)`  
**Purpose**: Watch actual collector start method call

### POST /collectors/:name/stop Endpoint

#### 12. **Stop Collector Handler Entry**
**File**: `app/handlers/collector.go`  
**Line**: `func (h *CollectorHandler) StopCollector(c *gin.Context)`  
**Purpose**: See collector stop request processing

#### 13. **Manager Stop Call**
**File**: `app/handlers/collector.go`  
**Line**: `err := h.CollectorManager.StopCollector(name)`  
**Purpose**: See manager delegation for stopping

#### 14. **Manager Stop Implementation**
**File**: `app/siem/collectors/manager.go`  
**Line**: `func (m *CollectorManager) StopCollector(name string) error {`  
**Purpose**: Watch actual collector stop logic

#### 15. **Individual Collector Stop**
**File**: `app/siem/collectors/manager.go`  
**Line**: `err := collector.Stop()`  
**Purpose**: Watch actual collector stop method call

### Collector-Specific Implementation Breakpoints

#### 16. **DSRC Collector Start**
**File**: `app/siem/collectors/dsrc_collector.go`  
**Line**: `func (c *DSRCCollector) Start(ctx context.Context) error {`  
**Purpose**: See V2X DSRC collector startup

#### 17. **CV2X Collector Start**
**File**: `app/siem/collectors/cv2x_collector.go`  
**Line**: `func (c *CV2XCollector) Start(ctx context.Context) error {`  
**Purpose**: See V2X C-V2X collector startup

#### 18. **SNMP Collector Start**
**File**: `app/siem/collectors/snmp_collector.go`  
**Line**: `func (c *SNMPCollector) Start(ctx context.Context) error {`  
**Purpose**: See SNMP trap collector startup

#### 19. **Syslog Collector Start**
**File**: `app/siem/collectors/syslog_collector.go`  
**Line**: `func (c *SyslogCollector) Start(ctx context.Context) error {`  
**Purpose**: See syslog collector startup

## Inspection Variables

At each breakpoint, inspect these key variables:

### For GET /collectors:
- `h.CollectorManager` - Manager instance with registered collectors
- `collectorNames` - Array of registered collector names
- `name` - Individual collector name being checked
- `status` - Running status (true/false) of each collector
- `collectors` - Final response array

### For Start/Stop operations:
- `name` - Collector name from URL parameter
- `m.collectors` - Map of registered collectors
- `collector` - Individual collector instance
- `m.ctx` - Context used for collector operations
- `err` - Error status from operations

### For Individual Collectors:
- `c.Running` - Boolean running state
- `c.Port` - Port number for network collectors
- `c.listener` - Network listener for UDP/TCP collectors
- `c.StopChan` - Channel used for graceful shutdown

## Registered Collectors

### Default Collectors Registered:
1. **"syslog"** - Syslog message collector (port 514)
2. **"snmp"** - SNMP trap collector (port 162)  
3. **"dsrc"** - DSRC V2X message collector (port 5001)
4. **"cv2x"** - C-V2X message collector (port 5002)

### Collector Interface Methods:
```go
type CollectorInterface interface {
    Start(ctx context.Context) error  // Start collection
    Stop() error                      // Stop collection  
    Name() string                     // Get collector name
    IsRunning() bool                  // Check if running
}
```

## Key Architecture Points

### Manager Pattern:
- **CollectorManager**: Central registry and lifecycle management
- **CollectorHandler**: HTTP API layer
- **Individual Collectors**: Specific protocol implementations

### Concurrency Model:
- Each collector runs in its own goroutine
- Graceful shutdown via context cancellation and stop channels
- Thread-safe status checking with mutex protection

### V2X Integration:
- DSRC and C-V2X collectors capture vehicle messages
- Automatic security event generation via EventIngester
- Real-time rule evaluation and alert generation