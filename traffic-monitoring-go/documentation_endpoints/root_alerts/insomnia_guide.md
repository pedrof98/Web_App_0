# Insomnia Configuration for GET /alerts

## Basic Setup
- **Method**: GET
- **URL**: `http://localhost:8080/alerts`
- **Headers**: None required (returns JSON by default)
- **Authentication**: None required
- **Body**: None required

## Query Parameter Variations

### 1. **Basic Request (Default)**
```
http://localhost:8080/alerts
```
**Returns**: First 50 alerts, most recent first

### 2. **Pagination Tests**
```
http://localhost:8080/alerts?page=1&pagesize=10
http://localhost:8080/alerts?page=2&pagesize=20
http://localhost:8080/alerts?page=1&pagesize=100
```

### 3. **Severity Filtering**
```
http://localhost:8080/alerts?severity=critical
http://localhost:8080/alerts?severity=high
http://localhost:8080/alerts?severity=medium
http://localhost:8080/alerts?severity=low
http://localhost:8080/alerts?severity=info
```

### 4. **Status Filtering**
```
http://localhost:8080/alerts?status=open
http://localhost:8080/alerts?status=closed
http://localhost:8080/alerts?status=in_progress
http://localhost:8080/alerts?status=false_positive
```

### 5. **Combined Filtering**
```
http://localhost:8080/alerts?severity=critical&status=open
http://localhost:8080/alerts?severity=high&status=in_progress&page=1&pagesize=25
http://localhost:8080/alerts?status=open&pagesize=100
```

## Expected Response Structure

### Success Response (200 OK):
```json
{
  "data": [
    {
      "id": 123,
      "rule_id": 1,
      "rule": {
        "id": 1,
        "name": "V2X Position Jump Detection",
        "description": "Detects vehicles with unrealistic position changes",
        "condition": "category = v2x AND anomalies contains position_jump",
        "severity": "high",
        "category": "v2x",
        "status": "enabled",
        "created_at": "2025-07-20T10:00:00Z"
      },
      "security_event_id": 456,
      "security_event": {
        "id": 456,
        "timestamp": "2025-07-28T10:30:15Z",
        "source_ip": "192.168.1.100",
        "log_source_id": 2,
        "log_source": {
          "id": 2,
          "name": "v2x-simulator",
          "type": "vehicle",
          "enabled": true
        },
        "severity": "high",
        "category": "v2x",
        "message": "Position jump anomaly detected in vehicle VEH-12345",
        "created_at": "2025-07-28T10:30:15Z"
      },
      "timestamp": "2025-07-28T10:30:15Z",
      "severity": "high",
      "status": "open",
      "assigned_to": null,
      "assigned_user": null,
      "resolution": "",
      "created_at": "2025-07-28T10:30:15Z",
      "updated_at": "2025-07-28T10:30:15Z"
    },
    {
      "id": 124,
      "rule_id": 4,
      "rule": {
        "id": 4,
        "name": "V2X Invalid Digital Signature",
        "description": "Detects messages with invalid digital signatures",
        "condition": "category = v2x AND signature_valid = false",
        "severity": "critical",
        "category": "v2x",
        "status": "enabled"
      },
      "security_event_id": 457,
      "security_event": {
        "id": 457,
        "timestamp": "2025-07-28T10:29:45Z",
        "source_ip": "192.168.1.101",
        "severity": "critical",
        "category": "v2x",
        "message": "Invalid digital signature detected",
        "log_source": {
          "id": 2,
          "name": "v2x-simulator",
          "type": "vehicle"
        }
      },
      "timestamp": "2025-07-28T10:29:45Z",
      "severity": "critical",
      "status": "open",
      "assigned_to": null,
      "resolution": "",
      "created_at": "2025-07-28T10:29:45Z"
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 50,
    "total": 87,
    "pages": 2
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

## Alert Status Values
- `"open"` - Newly created, needs attention
- `"in_progress"` - Being investigated
- `"closed"` - Resolved/completed
- `"false_positive"` - Determined to be benign

## Alert Severity Values
- `"critical"` - Immediate action required
- `"high"` - Important security issue
- `"medium"` - Moderate concern
- `"low"` - Minor issue
- `"info"` - Informational only

## Testing Sequence

1. **Test Basic Functionality**: GET `/alerts` (no parameters)
2. **Test Pagination**: Different page sizes and page numbers
3. **Test Severity Filtering**: Try each severity level
4. **Test Status Filtering**: Try each status type
5. **Test Combined Filters**: Multiple parameters together
6. **Test Edge Cases**: Invalid page numbers, empty results
7. **Test Large Page Sizes**: Performance with pagesize=100