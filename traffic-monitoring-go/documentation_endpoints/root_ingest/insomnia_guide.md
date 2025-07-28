# Insomnia Configuration for POST /ingest

## Basic Setup
- **Method**: POST
- **URL**: `http://localhost:8080/ingest`
- **Headers**: 
  - `Content-Type: application/json`
- **Authentication**: None required

## Test Payload 1: Basic V2X Event

```json
{
  "source_name": "v2x-simulator",
  "source_type": "v2x",
  "timestamp": "2025-07-27T10:30:00Z",
  "severity": "info",
  "category": "v2x",
  "message": "V2X BSM message from vehicle VEH001",
  "details": {
    "vehicle_id": "VEH001",
    "message_type": "bsm",
    "protocol": "DSRC",
    "location": "37.774929,-122.419416",
    "speed": 15.5,
    "signature_valid": true,
    "trust_level": 8,
    "source_ip": "192.168.1.100"
  }
}
```

## Test Payload 2: Attack Scenario (Position Jump)

```json
{
  "source_name": "v2x-simulator",
  "source_type": "v2x", 
  "timestamp": "2025-07-27T10:30:01Z",
  "severity": "warning",
  "category": "v2x",
  "message": "V2X BSM message from vehicle VEH001 with position anomaly",
  "details": {
    "vehicle_id": "VEH001",
    "message_type": "bsm",
    "protocol": "DSRC",
    "location": "37.785929,-122.419416",
    "speed": 15.5,
    "signature_valid": true,
    "trust_level": 8,
    "source_ip": "192.168.1.100",
    "anomalies": [
      {
        "type": "position_jump",
        "confidence": 0.95,
        "details": {
          "distance_jumped": 150.0,
          "time_difference": 0.5,
          "previous_location": "37.774929,-122.419416"
        }
      }
    ]
  }
}
```

## Test Payload 3: Invalid Signature Attack

```json
{
  "source_name": "v2x-simulator",
  "source_type": "v2x",
  "timestamp": "2025-07-27T10:30:02Z", 
  "severity": "critical",
  "category": "v2x",
  "message": "V2X BSM message with invalid signature detected",
  "details": {
    "vehicle_id": "VEH002",
    "message_type": "bsm",
    "protocol": "C-V2X",
    "location": "37.774929,-122.419416",
    "speed": 12.0,
    "signature_valid": false,
    "trust_level": 0,
    "source_ip": "192.168.1.101",
    "security_error": "Digital signature verification failed"
  }
}
```

## Expected Responses

### Success Response (200 OK):
```json
{
  "message": "Event ingested and processed successfully",
  "event_id": 123,
  "alerts_created": 1
}
```

### Error Response (400 Bad Request):
```json
{
  "error": "Failed to read request body"
}
```

### Error Response (500 Internal Server Error):
```json
{
  "error": "Database connection error"
}
```

## Testing Sequence

1. **Test Basic Event**: Use payload 1 to ensure normal flow works
2. **Test Position Jump**: Use payload 2 to trigger position jump rule
3. **Test Invalid Signature**: Use payload 3 to trigger signature validation rule
4. **Test Malformed JSON**: Send invalid JSON to test error handling