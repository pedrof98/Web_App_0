# Insomnia Configuration for GET /rules

## Basic Setup
- **Method**: GET
- **URL**: `http://localhost:8080/rules`
- **Headers**: None required (returns JSON by default)
- **Authentication**: None required
- **Body**: None required

## Query Parameter Variations

### 1. **Basic Request (All Rules)**
```
http://localhost:8080/rules
```
**Returns**: All rules in the system, sorted alphabetically by name

### 2. **Filter by Status**
```
http://localhost:8080/rules?status=enabled
http://localhost:8080/rules?status=disabled
http://localhost:8080/rules?status=testing
```

### 3. **Filter by Category**
```
http://localhost:8080/rules?category=v2x
http://localhost:8080/rules?category=network
http://localhost:8080/rules?category=authentication
http://localhost:8080/rules?category=system
```

### 4. **Combined Filtering**
```
http://localhost:8080/rules?status=enabled&category=v2x
http://localhost:8080/rules?status=disabled&category=network
```

## Expected Response Structure

### Success Response (200 OK):
```json
[
  {
    "id": 1,
    "name": "V2X Position Jump Detection",
    "description": "Detects vehicles with impossible position changes (>100m in <1 second)",
    "condition": "category = v2x AND raw_data.anomalies contains position_jump AND raw_data.anomalies[0].confidence > 0.7",
    "severity": "high",
    "category": "v2x",
    "status": "enabled",
    "created_by": 1,
    "created_at": "2025-07-20T10:00:00Z",
    "updated_at": "2025-07-20T10:00:00Z"
  },
  {
    "id": 2,
    "name": "V2X Speed Anomaly Detection",
    "description": "Detects unrealistic speed changes (>10 m/s difference between consecutive messages)",
    "condition": "category = v2x AND raw_data.anomalies contains speed_jump AND raw_data.anomalies[0].confidence > 0.8",
    "severity": "medium",
    "category": "v2x",
    "status": "enabled",
    "created_by": 1,
    "created_at": "2025-07-20T10:00:00Z",
    "updated_at": "2025-07-20T10:00:00Z"
  },
  {
    "id": 3,
    "name": "V2X Message Flooding Attack",
    "description": "Detects abnormally high message frequency (>10 messages per second)",
    "condition": "category = v2x AND raw_data.anomalies contains high_frequency AND raw_data.anomalies[0].confidence > 0.8",
    "severity": "high",
    "category": "v2x",
    "status": "enabled",
    "created_by": 1,
    "created_at": "2025-07-20T10:00:00Z",
    "updated_at": "2025-07-20T10:00:00Z"
  },
  {
    "id": 4,
    "name": "V2X Invalid Digital Signature",
    "description": "Detects messages with invalid security signatures",
    "condition": "category = v2x AND raw_data.signature_valid = false",
    "severity": "critical",
    "category": "v2x",
    "status": "enabled",
    "created_by": 1,
    "created_at": "2025-07-20T10:00:00Z",
    "updated_at": "2025-07-20T10:00:00Z"
  },
  {
    "id": 5,
    "name": "V2X DENM High Priority Alert",
    "description": "Escalates high-priority Decentralized Environmental Notification Messages",
    "condition": "category = v2x AND raw_data.message_type = denm AND raw_data.priority >= 8",
    "severity": "high",
    "category": "v2x",
    "status": "enabled",
    "created_by": 1,
    "created_at": "2025-07-20T10:00:00Z",
    "updated_at": "2025-07-20T10:00:00Z"
  },
  {
    "id": 6,
    "name": "V2X BSM Timing Violation",
    "description": "Detects Basic Safety Messages sent too frequently (violating 100ms standard)",
    "condition": "category = v2x AND raw_data.message_type = bsm AND raw_data.interval_ms < 50",
    "severity": "low",
    "category": "v2x",
    "status": "enabled",
    "created_by": 1,
    "created_at": "2025-07-20T10:00:00Z",
    "updated_at": "2025-07-20T10:00:00Z"
  }
]
```

### Empty Results Response (200 OK):
```json
[]
```

### Error Response (500 Internal Server Error):
```json
{
  "error": "database connection error"
}
```

## Rule Status Values
- `"enabled"` - Rule is active and evaluating events
- `"disabled"` - Rule is inactive, won't generate alerts
- `"testing"` - Rule is in test mode (may log but not alert)

## Rule Category Values
- `"v2x"` - Vehicle-to-Everything communication rules
- `"network"` - Network security rules
- `"authentication"` - Login/access control rules
- `"system"` - System-level security rules

## Rule Severity Values
- `"critical"` - Immediate response required
- `"high"` - Important security issue
- `"medium"` - Moderate concern
- `"low"` - Minor issue
- `"info"` - Informational only

## Understanding Rule Conditions

### V2X-Specific Rule Conditions:
```sql
-- Position Jump Detection
"category = v2x AND raw_data.anomalies contains position_jump AND raw_data.anomalies[0].confidence > 0.7"

-- Invalid Signature Detection  
"category = v2x AND raw_data.signature_valid = false"

-- Message Flooding Detection
"category = v2x AND raw_data.anomalies contains high_frequency AND raw_data.anomalies[0].confidence > 0.8"

-- Speed Anomaly Detection
"category = v2x AND raw_data.anomalies contains speed_jump AND raw_data.anomalies[0].confidence > 0.8"
```

## Testing Sequence

1. **Test Basic Functionality**: GET `/rules` (no parameters)
2. **Test Status Filtering**: Filter by enabled/disabled/testing
3. **Test Category Filtering**: Focus on V2X rules
4. **Test Combined Filters**: Status + category combinations
5. **Analyze Rule Conditions**: Understand detection logic
6. **Cross-reference with Alerts**: See which rules generate alerts