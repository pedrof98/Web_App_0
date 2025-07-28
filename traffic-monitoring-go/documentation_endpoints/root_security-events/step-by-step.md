# Step-by-Step Debug Process for GET /security-events

## Prerequisites
1. Database with security event data (run `POST /ingest` or `POST /test/v2x-rules` to generate events)
2. Understanding that security events are the foundation of the SIEM system
3. Knowledge that security events are created by ingestion and trigger rule evaluation

## Debug Session Steps

### Step 1: Generate Test Security Events (Recommended)
1. Run `POST /test/v2x-rules` to create security events
2. Or run `POST /ingest` with V2X payloads to create individual events
3. Verify events exist: `SELECT COUNT(*) FROM security_events;`
4. Check event distribution: `SELECT category, severity, COUNT(*) FROM security_events GROUP BY category, severity;`

### Step 2: Launch Debugger and Send Request
1. Use same VS Code debug configuration
2. Set breakpoints from the security events breakpoints guide
3. Start debugging
4. Send GET request to `http://localhost:8080/security-events`

### Step 3: Debug Flow Analysis

#### Breakpoint 1: Security Event Handler Entry
- **Variable to inspect**: `h`, `c`
- **What to check**: Handler initialization with DB and ESService
- **Expected**: Valid SecurityEventHandler with database connection and optional Elasticsearch service

#### Breakpoint 2: Pagination Parameter Extraction
- **Variable to inspect**: `page`, `pageSize`, `offset`
- **What to check**: Parameter parsing and defaults
- **Expected Values**:
  ```
  No params: page=1, pageSize=50, offset=0
  ?page=2&pageSize=20: page=2, pageSize=20, offset=20
  ?page=3: page=3, pageSize=50, offset=100
  ```

#### Breakpoint 3: Severity Filter Parameter Processing
- **Variable to inspect**: `severity`
- **What to check**: Query parameter extraction for severity filtering
- **Expected Values**:
  ```
  No filter: severity=""
  ?severity=critical: severity="critical"
  ?severity=high: severity="high"
  ?severity=v2x: severity="v2x"
  ```

#### Breakpoint 4: Category Filter Parameter Processing
- **Variable to inspect**: `category`
- **What to check**: Query parameter extraction for category filtering
- **Expected Values**:
  ```
  No filter: category=""
  ?category=v2x: category="v2x"
  ?category=network: category="network"
  ?category=authentication: category="authentication"
  ```

#### Breakpoint 5: Query Builder Creation
- **Variable to inspect**: `query` (GORM query object)
- **What to check**: Base query targeting SecurityEvent model
- **Expected**: Simple GORM query **WITHOUT** preloads
- **Key insight**: Unlike alerts, NO relationship preloading occurs here

#### Breakpoint 6: Severity Filter Application
- **Variable to inspect**: `query` after filter
- **What to check**: Conditional WHERE clause addition
- **Expected**: If severity provided, WHERE severity = ? clause added
- **Debug tip**: Print `query.Statement.SQL` to see generated SQL

#### Breakpoint 7: Category Filter Application
- **Variable to inspect**: `query` after both filters
- **What to check**: Multiple WHERE clauses chained
- **Expected**: Both severity AND category filters if provided

#### Breakpoint 8: Ordering and Sorting
- **Variable to inspect**: `query` with ORDER BY
- **What to check**: Sorting by timestamp descending
- **Expected**: Most recent events first (same as alerts)

#### Breakpoint 9: Total Count Query
- **Variable to inspect**: `total`
- **What to check**: Total count for pagination (before LIMIT/OFFSET)
- **Expected**: Number representing total events matching filters
- **Key insight**: Count executed BEFORE pagination for accurate total

#### Breakpoint 10: Paginated Results Query
- **Variable to inspect**: `events` array, `err`
- **What to check**: Final database query execution **without** relationships
- **Expected**: Array of SecurityEvent objects with NO LogSource data populated
- **Key insight**: Fast, simple query with no JOINs

#### Breakpoint 11: Response Construction
- **Variable to inspect**: Final response structure
- **What to check**: JSON response with data and pagination metadata
- **Expected**:
  ```json
  {
    "data": [...], // SecurityEvent objects without relationships
    "pagination": {
      "page": 1,
      "pageSize": 50, 
      "total": 2567,
      "pages": 52
    }
  }
  ```

### Step 4: Understanding Security Event Data Model

#### Core SecurityEvent Fields:
```go
type SecurityEvent struct {
    ID              uint          // Primary key
    Timestamp       time.Time     // When event occurred
    SourceIP        string        // Source IP (if network event)
    SourcePort      *int          // Source port (optional)
    DestinationIP   string        // Destination IP (optional)
    DestinationPort *int          // Destination port (optional)
    Protocol        string        // Network protocol (DSRC, C-V2X, etc.)
    Action          string        // Action taken (allow, block, alert)
    Status          string        // Status (success, failure)
    DeviceID        string        // Device identifier
    LogSourceID     uint          // Foreign key to log_sources
    LogSource       LogSource     // NOT preloaded in this endpoint
    Severity        EventSeverity // critical, high, medium, low, info
    Category        EventCategory // v2x, network, authentication, etc.
    Message         string        // Human-readable description
    RawData         string        // Original JSON from ingestion
    CreatedAt       time.Time     // Database insertion time
}
```

#### Key Field Analysis:
- **Timestamp vs CreatedAt**: Timestamp = when event occurred, CreatedAt = when stored
- **RawData**: Contains complete original event context as JSON string
- **LogSourceID**: Foreign key present but LogSource not loaded (performance optimization)
- **Optional Fields**: Many fields can be null/empty depending on event type

### Step 5: Performance Analysis and Comparison

#### Generated SQL Analysis:
```sql
-- Count query (for pagination total)
SELECT COUNT(*) FROM security_events 
WHERE severity = 'critical' AND category = 'v2x';

-- Main query (simplified)
SELECT * FROM security_events
WHERE severity = 'critical' AND category = 'v2x'
ORDER BY timestamp DESC
LIMIT 50 OFFSET 0;
```

#### Performance Characteristics:
- **No JOINs**: Much faster than alerts endpoint
- **Simple WHERE clauses**: Efficient filtering
- **Index usage**: Leverages indexes on severity, category, timestamp
- **Large result sets**: Can handle thousands of events efficiently

#### vs GET /alerts Performance:
```
Security Events Query Time: ~2-5ms (no JOINs)
Alerts Query Time: ~10-25ms (multiple JOINs)
Throughput: 5-10x higher for security events
```

### Step 6: Raw Data Analysis

#### Understanding RawData Field:
The `raw_data` field contains the original JSON from ingestion:

```json
{
  "source_name": "v2x-simulator",
  "source_type": "v2x",
  "timestamp": "2025-07-28T10:30:15Z",
  "severity": "high",
  "category": "v2x",
  "message": "Position jump anomaly detected",
  "details": {
    "vehicle_id": "VEH-12345",
    "message_type": "bsm",
    "anomalies": [{
      "type": "position_jump",
      "confidence": 0.95,
      "details": {
        "distance_jumped": 150.0,
        "time_difference": 0.5,
        "previous_location": "37.774929,-122.419416",
        "current_location": "37.785929,-122.419416"
      }
    }]
  }
}
```

#### Raw Data Use Cases:
- **Complete context**: All original event details preserved
- **Forensic analysis**: Investigate attack details
- **Rule debugging**: Understand why rules triggered
- **Data export**: Extract for external analysis tools

### Step 7: Event Lifecycle Understanding

#### Event Flow in SIEM System:
```
1. V2X Message/Log → POST /ingest → SecurityEvent (this endpoint shows)
2. SecurityEvent → Rule Engine → Alert (GET /alerts shows)
3. Alert → Assignment/Resolution → Incident Response
```

#### Relationship to Other Endpoints:
- **POST /ingest**: Creates the events you see here
- **GET /alerts**: Shows alerts generated from these events
- **GET /rules**: Shows rules that evaluate these events
- **Dashboard endpoints**: Aggregate statistics from these events

### Step 8: Response Validation and Analysis

#### Data Integrity Checks:
- **Timestamp ordering**: Events should be in descending timestamp order
- **Foreign keys**: log_source_id should reference valid log sources
- **Category/Severity**: Should match expected enum values
- **Raw data**: Should be valid JSON string
- **Pagination math**: Verify pages calculation: `(total + pageSize - 1) / pageSize`

#### Real-world Event Examples:
```json
{
  "id": 1234,
  "timestamp": "2025-07-28T10:30:15Z",
  "source_ip": "192.168.1.100",
  "log_source_id": 2,
  "severity": "high",
  "category": "v2x",
  "message": "Position jump anomaly detected in vehicle VEH-12345",
  "raw_data": "{...complete original context...}"
}
```

## Key Learning Points

### SIEM Event Management:
1. **Event Foundation**: Security events are the raw data foundation
2. **No Relationships**: Optimized for bulk viewing without context
3. **Raw Data Preservation**: Complete original event context maintained
4. **High Performance**: Designed for processing large event volumes

### GORM Query Patterns:
1. **Simple Queries**: No preloading for performance
2. **Conditional Filtering**: Dynamic WHERE clause building
3. **Pagination**: Standard count + limit/offset pattern
4. **Index Utilization**: Efficient filtering on indexed columns

### V2X Security Context:
1. **Event Types**: BSM, DENM, position anomalies, signature failures
2. **Severity Assignment**: Based on threat level and impact
3. **Category Classification**: V2X vs network vs authentication events
4. **Performance Requirements**: Real-time processing for safety-critical systems

## Common Issues and Debugging

### Issue: No Events Returned
- **Check**: Events exist in database (`SELECT COUNT(*) FROM security_events`)
- **Verify**: Ingestion has occurred (`POST /ingest` or test data generation)
- **Solution**: Generate test events using `POST /test/v2x-rules`

### Issue: Missing Event Data
- **Check**: Database schema and migrations completed
- **Verify**: Required fields populated during ingestion
- **Solution**: Check ingestion logic in `siem/ingestion.go`

### Issue: Slow Performance
- **Check**: Database indexes on filtered columns (severity, category, timestamp)
- **Verify**: Query execution plan using EXPLAIN
- **Solution**: This endpoint is already optimized - compare with alerts endpoint

### Issue: Raw Data Corruption
- **Check**: JSON validity of raw_data field
- **Verify**: Ingestion process preserving original event structure
- **Solution**: Validate JSON during ingestion process

## Advanced Analysis Tips

1. **Event Pattern Analysis**: Look for event frequency and timing patterns
2. **Source Analysis**: Examine source_ip and log_source_id distributions
3. **Severity Trends**: Monitor severity distribution over time
4. **Category Analysis**: Compare V2X vs other event categories
5. **Raw Data Mining**: Extract specific details from raw_data JSON

## Event-to-Alert Correlation

Understanding the relationship between events and alerts:
```sql
-- Find alerts generated from specific events
SELECT se.message, a.severity, r.name as rule_name
FROM security_events se
LEFT JOIN alerts a ON a.security_event_id = se.id
LEFT JOIN rules r ON a.rule_id = r.id
WHERE se.category = 'v2x'
ORDER BY se.timestamp DESC;

-- Event-to-alert conversion rate
SELECT 
  COUNT(se.id) as total_events,
  COUNT(a.id) as total_alerts,
  (COUNT(a.id) * 100.0 / COUNT(se.id)) as conversion_rate
FROM security_events se
LEFT JOIN alerts a ON a.security_event_id = se.id
WHERE se.category = 'v2x';
```

This endpoint provides the **raw event data** that drives the entire V2X SIEM system - understanding it is crucial for effective security monitoring!