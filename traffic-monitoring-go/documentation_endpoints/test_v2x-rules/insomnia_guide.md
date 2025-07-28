# Insomnia Configuration for POST /test/v2x-rules

## Basic Setup
- **Method**: POST
- **URL**: `http://localhost:8080/test/v2x-rules`
- **Headers**: 
  - `Content-Type: application/json`
- **Authentication**: None required
- **Body**: Empty (endpoint generates its own test data)

## Request Body
**Important**: This endpoint doesn't require a request body! It generates predefined test scenarios internally.

```json
{}
```

## What This Endpoint Does

This endpoint automatically generates 6 predefined test scenarios:

### Test Scenario 1: Position Jump Attack
```json
{
  "source_name": "v2x",
  "source_type": "vehicle", 
  "severity": "high",
  "category": "v2x",
  "message": "Position jump anomaly detected in vehicle VEH-12345",
  "details": {
    "source_ip": "192.168.1.100",
    "protocol": "DSRC", 
    "message_type": "bsm",
    "vehicle_id": "VEH-12345",
    "anomalies": [
      {
        "type": "position_jump",
        "confidence": 0.95,
        "details": {
          "distance_jumped": 150.0,
          "time_difference": 0.5
        }
      }
    ]
  }
}
```

### Test Scenario 2: Speed Jump Anomaly
```json
{
  "details": {
    "anomalies": [
      {
        "type": "speed_jump",
        "confidence": 0.88,
        "previous_speed": 25.0,
        "current_speed": 5.0
      }
    ]
  }
}
```

### Test Scenario 3: Message Flooding Attack
```json
{
  "details": {
    "anomalies": [
      {
        "type": "high_frequency",
        "confidence": 0.92,
        "message_frequency": 25.0
      }
    ]
  }
}
```

### Test Scenario 4: Invalid Digital Signature
```json
{
  "severity": "critical",
  "message": "Invalid digital signature detected",
  "details": {
    "signature_valid": false,
    "trust_level": 0
  }
}
```

### Test Scenario 5: Emergency Vehicle Alert
```json
{
  "details": {
    "message_type": "emergency_vehicle",
    "priority": 9,
    "alert_type": "emergency_vehicle"
  }
}
```

### Test Scenario 6: High Priority DENM Alert
```json
{
  "details": {
    "message_type": "denm",
    "priority": 9,
    "alert_type": "emergency_vehicle"
  }
}
```

## Expected Response

### Success Response (200 OK):
```json
{
  "message": "V2X security rules test completed",
  "events_processed": 6,
  "total_alerts_created": 4,
  "results": [
    {
      "event_index": 0,
      "status": "success", 
      "event_id": 125,
      "alerts_created": 1,
      "event_type": "Position jump anomaly detected in vehicle VEH-12345"
    },
    {
      "event_index": 1,
      "status": "success",
      "event_id": 126, 
      "alerts_created": 1,
      "event_type": "Speed jump anomaly detected in vehicle VEH-67890"
    },
    // ... more test results
  ],
  "note": "Check /alerts endpoint to see triggered alerts"
}
```

### Error Response (500 Internal Server Error):
```json
{
  "error": "Database transaction failed"
}
```

## Testing Sequence

1. **Send Test Request**: POST to endpoint with empty body
2. **Check Response**: Verify all 6 events processed
3. **Verify Alerts**: Check `total_alerts_created` count
4. **Database Verification**: Query alerts table to see created alerts
5. **Rule Testing**: Verify which V2X rules were triggered