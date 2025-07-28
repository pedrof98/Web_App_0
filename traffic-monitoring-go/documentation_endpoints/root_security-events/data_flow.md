# Security Events Data Flow and SIEM Architecture

## Complete Data Flow: From V2X Message to Security Event

```
V2X Message → Ingestion → Security Event → Rule Evaluation → Alert
                               ↑
                   GET /security-events retrieves this
```

### 1. **Security Event Creation Process** (How events are born)
```go
// In ingestion engine (from POST /ingest)
securityEvent := models.SecurityEvent{
    Timestamp:   rawEvent.Timestamp,     // When event occurred
    LogSourceID: logSourceID,            // Source system ID
    Severity:    rawEvent.Severity,      // Event severity level
    Category:    rawEvent.Category,      // Event type (v2x, network, etc.)
    Message:     rawEvent.Message,       // Human-readable description
    RawData:     string(rawEventData),   // Complete original JSON
    SourceIP:    extractedFromDetails,   // Network info if applicable
    Protocol:    extractedFromDetails,   // Communication protocol
}
```

### 2. **Database Schema and Relationships**
```sql
-- Core table structure
security_events.id                    -- Primary key
security_events.timestamp             -- Event occurrence time (indexed)
security_events.log_source_id         -- FK to log_sources.id
security_events.severity              -- critical|high|medium|low|info (indexed)
security_events.category              -- v2x|network|authentication|system (indexed)
security_events.message               -- Human-readable description
security_events.raw_data              -- Original JSON (TEXT field)
security_events.source_ip             -- Source IP address (indexed)
security_events.created_at            -- Database insertion time

-- Foreign key relationships
security_events.log_source_id → log_sources.id (NOT preloaded for performance)

-- Reverse relationships (events trigger these)
alerts.security_event_id → security_events.id (events create alerts)
```

### 3. **GORM Query Analysis Deep Dive**

#### What this endpoint does vs. Alerts endpoint:
```go
// Security Events (this endpoint) - FAST
query := h.DB.Model(&models.SecurityEvent{})
// NO Preload() calls = No JOINs = High performance

// Alerts endpoint - COMPREHENSIVE  
query := h.DB.Model(&models.Alert{}).
    Preload("Rule").
    Preload("SecurityEvent").
    Preload("SecurityEvent.LogSource")
// Multiple Preload() calls = Multiple JOINs = More context
```

#### Generated SQL Comparison:
```sql
-- Security Events Query (this endpoint)
SELECT * FROM security_events 
WHERE severity = 'critical' AND category = 'v2x'
ORDER BY timestamp DESC 
LIMIT 50 OFFSET 0;

-- Alerts Query (for comparison)
SELECT alerts.*, rules.*, security_events.*, log_sources.*
FROM alerts
LEFT JOIN rules ON alerts.rule_id = rules.id  
LEFT JOIN security_events ON alerts.security_event_id = security_events.id
LEFT JOIN log_sources ON security_events.log_source_id = log_sources.id
WHERE alerts.severity = 'critical'
ORDER BY alerts.timestamp DESC
LIMIT 50 OFFSET 0;
```

### 4. **Performance Characteristics**

#### Query Performance Metrics:
```
Security Events (no JOINs):
- Simple query time: 2-5ms
- Large dataset (10k+ events): 5-15ms
- Throughput: 500+ requests/second
- Memory usage: Low (no relationship loading)

Alerts (with JOINs):
- Complex query time: 10-25ms  
- Large dataset: 20-50ms
- Throughput: 100-200 requests/second
- Memory usage: Higher (full relationship context)
```

#### When to Use Each Endpoint:
- **Security Events**: Bulk analysis, performance monitoring, event investigation
- **Alerts**: Alert management, incident response, rule debugging

### 5. **Event Categories and V2X Context**

#### V2X-Specific Event Categories:
```json
{
  "v2x": "Vehicle-to-Everything communication events",
  "vehicle": "General vehicle-related events", 
  "network": "Network security events",
  "authentication": "Login/access control",
  "system": "System-level security events"
}
```

#### V2X Event Severity Assignment:
```go
// In V2X collectors (dsrc_collector.go, cv2x_collector.go)
switch eventType {
case "invalid_signature":
    severity = models.SeverityCritical  // Active attack
case "position_jump":
    severity = models.SeverityHigh      // Likely spoofing
case "speed_anomaly":
    severity = models.SeverityMedium    // Suspicious behavior
case "normal_bsm":
    severity = models.SeverityInfo      // Regular traffic
}
```

### 6. **Raw Data Structure and Analysis**

#### Complete V2X Event Raw Data Example:
```json
{
  "source_name": "v2x-simulator",
  "source_type": "v2x",
  "event_type": "position_jump",
  "timestamp": "2025-07-28T10:30:00Z",
  "message": "Vehicle 'veh-123' reported an impossible position jump.",
  "severity": "high",
  "details": {
    "vehicle_id": "veh-123",
    "protocol": "DSRC",
    "previous_position": {
      "latitude": 34.0522,
      "longitude": -118.2437
    },
    "current_position": {
      "latitude": 34.1522,
      "longitude": -118.3437
    },
    "time_delta_ms": 100,
    "distance_meters": 15700
  }
}
```

#### Raw Data Field Analysis:
- **`source_name` / `source_type`**: Identifies the origin of the data, used to link the event to a `LogSource` in the database.
- **`event_type`**: A specific machine-readable identifier for the event (e.g., `position_jump`). This is used by collectors to assign a `severity`.
- **`timestamp`**: The precise time the event occurred at the source. This is crucial for chronological analysis and is the primary sorting key.
- **`message`**: A human-readable summary. While useful for dashboards, the `event_type` and `details` are used for automated rule processing.
- **`details`**: A nested JSON object containing the specific, contextual data of the event. This is where the core evidence for the event is stored. For a `position_jump`, it includes previous and current coordinates, allowing an analyst or automated system to verify the anomaly. The entire `details` object is stored in the `raw_data` field of the `security_events` table.