# Insomnia Configuration for GET /security-events

## Basic Setup
- **Method**: GET
- **URL**: `http://localhost:8080/security-events`
- **Headers**: None required (returns JSON by default)
- **Authentication**: None required
- **Body**: None required

## Query Parameter Variations

### 1. **Basic Request (Default)**
```
http://localhost:8080/security-events
```
**Returns**: First 50 security events, most recent first

### 2. **Pagination Tests**
```
http://localhost:8080/security-events?page=1&pageSize=10
http://localhost:8080/security-events?page=2&pageSize=20
http://localhost:8080/security-events?page=1&pageSize=100
```

### 3. **Severity Filtering**
```
http://localhost:8080/security-events?severity=critical
http://localhost:8080/security-events?severity=high
http://localhost:8080/security-events?severity=medium
http://localhost:8080/security-events?severity=low
http://localhost:8080/security-events?severity=info
```

### 4. **Category Filtering**
```
http://localhost:8080/security-events?category=v2x
http://localhost:8080/security-events?category=network
http://localhost:8080/security-events?category=authentication
http://localhost:8080/security-events?category=system
http://localhost:8080/security-events?category=vehicle
```

### 5. **Combined Filtering**
```
http://localhost:8080/security-events?severity=critical&category=v2x
http://localhost:8080/security-events?severity=high&category=network&page=1&pageSize=25
http://localhost:8080/security-events?category=v2x&pageSize=100
```

## Expected Response Structure

### Success Response (200 OK):
```json
{
  "data": [
    {
      "id": 1234,
      "timestamp": "2025-07-28T10:30:15Z",
      "source_ip": "192.168.1.100",
      "source_port": null,
      "destination_ip": "",
      "destination_port": null,
      "protocol": "DSRC",
      "action": "",
      "status": "",
      "device_id": "",
      "log_source_id": 2,
      "severity": "high",
      "category": "v2x",
      "message": "Position jump anomaly detected in vehicle VEH-12345",
      "raw_data": "{\"source_name\":\"v2x-simulator\",\"source_type\":\"v2x\",\"timestamp\":\"2025-07-28T10:30:15Z\",\"severity\":\"high\",\"category\":\"v2x\",\"message\":\"Position jump anomaly detected in vehicle VEH-12345\",\"details\":{\"vehicle_id\":\"VEH-12345\",\"anomalies\":[{\"type\":\"position_jump\",\"confidence\":0.95}]}}",
      "created_at": "2025-07-28T10:30:15Z"
    },
    {
      "id": 1235,
      "timestamp": "2025-07-28T10:29:45Z",
      "source_ip": "192.168.1.101",
      "source_port": null,
      "destination_ip": "",
      "destination_port": null,
      "protocol": "C-V2X",
      "action": "",
      "status": "",
      "device_id": "",
      "log_source_id": 2,
      "severity": "critical",
      "category": "v2x",
      "message": "Invalid digital signature detected",
      "raw_data": "{\"source_name\":\"v2x-simulator\",\"source_type\":\"v2x\",\"timestamp\":\"2025-07-28T10:29:45Z\",\"severity\":\"critical\",\"category\":\"v2x\",\"message\":\"Invalid digital signature detected\",\"details\":{\"vehicle_id\":\"VEH-12346\",\"signature_valid\":false}}",
      "created_at": "2025-07-28T10:29:45Z"
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 50,
    "total": 2567,
    "pages": 52
  }
}
```

### Empty Results Response (200 OK):
```json
{
  "data": [],
  "pagination": {
    "page": 1,
    "pageSize": 50,
    "total": 0,
    "pages": 0
  }
}
```

### Error Response (500 Internal Server Error):
```json
{
  "error": "database connection error"
}
```

## Event Severity Values
- `"critical"` - Active attacks, immediate threat
- `"high"` - Important security issues requiring attention
- `"medium"` - Moderate concerns, potential issues
- `"low"` - Minor issues or informational warnings
- `"info"` - Normal operational events

## Event Category Values
- `"v2x"` - Vehicle-to-Everything communication events
- `"vehicle"` - General vehicle-related events
- `"network"` - Network security events
- `"authentication"` - Login/access control events
- `"authorization"` - Permission/privilege events
- `"system"` - System-level events
- `"malware"` - Malware detection events

## Key Field Explanations

### Core Event Fields:
- **`id`**: Unique database identifier
- **`timestamp`**: When the original event occurred
- **`source_ip`**: Source IP address (if network-related)
- **`log_source_id`**: Foreign key to log_sources table (not preloaded)
- **`severity`**: Event severity level
- **`category`**: Event category/type
- **`message`**: Human-readable event description
- **`raw_data`**: Original JSON from ingestion (contains full context)
- **`created_at`**: When event was stored in database

### V2X-Specific Raw Data Examples:
```json
// Position Jump Detection
{
  "source_name": "v2x-simulator",
  "details": {
    "vehicle_id": "VEH-12345",
    "anomalies": [{
      "type": "position_jump",
      "confidence": 0.95,
      "details": {
        "distance_jumped": 150.0,
        "time_difference": 0.5
      }
    }]
  }
}

// Invalid Signature
{
  "source_name": "dsrc-collector",
  "details": {
    "vehicle_id": "VEH-12346",
    "signature_valid": false,
    "security_error": "Digital signature verification failed"
  }
}
```

## Testing Sequence

1. **Test Basic Functionality**: GET `/security-events` (no parameters)
2. **Test Pagination**: Different page sizes and page numbers
3. **Test Severity Filtering**: Try each severity level
4. **Test Category Filtering**: Focus on V2X and network events
5. **Test Combined Filters**: Multiple parameters together
6. **Test Large Datasets**: Use pageSize=100 to test performance
7. **Test Edge Cases**: Invalid page numbers, empty results
8. **Analyze Raw Data**: Examine `raw_data` field for event context

## Comparison with Other Endpoints

### vs GET /alerts:
- **Security Events**: Raw event data, no relationships
- **Alerts**: Processed events with Rule and SecurityEvent context

### vs POST /ingest:
- **Ingest**: Creates new security events
- **Security Events**: Retrieves existing events

### Data Flow Context:
```
V2X Message → POST /ingest → Security Event → Rule Engine → Alert
                                ↑
                        GET /security-events
                        (retrieves this data)
```

## Performance Considerations

### Query Performance:
- **No JOINs**: Faster than alerts endpoint
- **Simple filters**: Only severity and category
- **Indexed fields**: timestamp, severity, category, source_ip
- **Large datasets**: Can handle thousands of events efficiently

### Use Cases:
- **Event investigation**: Examine raw event details
- **Bulk analysis**: Process large volumes of events
- **Performance monitoring**: High-throughput event viewing
- **Data export**: Extract event data for external analysis