# Insomnia Configuration for GET /v2x-dashboard/overview

## Basic Setup
- **Method**: GET
- **URL**: `http://localhost:8080/v2x-dashboard/overview`
- **Headers**: None required (returns JSON by default)
- **Authentication**: None required
- **Body**: None required

## Query Parameters (Optional)

### Test URLs with Different Time Ranges:

#### Default (Last Hour):
```
http://localhost:8080/v2x-dashboard/overview
```

#### Last Day:
```
http://localhost:8080/v2x-dashboard/overview?timeRange=last_day
```

#### Last Week:
```
http://localhost:8080/v2x-dashboard/overview?timeRange=last_week
```

#### Last Month:
```
http://localhost:8080/v2x-dashboard/overview?timeRange=last_month
```

## Expected Response Structure

### Success Response (200 OK):
```json
{
  "summary": {
    "total": 1250,
    "dsrc": 800,
    "cv2x": 450,
    "bsm": 900,
    "spat": 150,
    "rsa": 75,
    "cam": 80,
    "denm": 25,
    "cpm": 20,
    "pc5_interface": 300,
    "uu_interface": 150
  },
  "security_summary": {
    "total": 1250,
    "valid_signature": 1180,
    "invalid_signature": 70,
    "high_trust_level": 1100,
    "low_trust_level": 25,
    "detected_anomalies": 45,
    "high_confidence_anomaly": 15
  },
  "anomaly_summary": {
    "total": 45,
    "position_jump": 12,
    "speed_jump": 8,
    "heading_jump": 3,
    "high_frequency": 15,
    "conflicting_alerts": 2,
    "timing_anomaly": 3,
    "other": 2
  },
  "vehicle_locations": [
    {
      "vehicle_id": "VEH001",
      "latitude": 37.774929,
      "longitude": -122.419416,
      "protocol": "DSRC",
      "message_type": "bsm",
      "timestamp": "2025-07-28T10:30:00Z",
      "rssi": -45,
      "has_anomaly": false
    },
    {
      "vehicle_id": "VEH002", 
      "latitude": 37.775929,
      "longitude": -122.420416,
      "protocol": "C-V2X",
      "message_type": "cam",
      "timestamp": "2025-07-28T10:29:55Z",
      "rssi": -52,
      "has_anomaly": true
    }
    // ... up to 25 recent vehicles
  ],
  "active_alerts": [
    {
      "id": 123,
      "alert_type": "position_jump",
      "description": "Vehicle VEH003 experienced suspicious position jump",
      "priority": 8,
      "latitude": 37.776929,
      "longitude": -122.421416,
      "radius": 100,
      "timestamp": "2025-07-28T10:28:30Z"
    },
    {
      "id": 124,
      "alert_type": "invalid_signature",
      "description": "Message with invalid digital signature detected",
      "priority": 9,
      "latitude": 37.777929,
      "longitude": -122.422416,
      "radius": 50,
      "timestamp": "2025-07-28T10:27:15Z"
    }
    // ... more active alerts
  ]
}
```

### Empty Data Response (200 OK):
```json
{
  "summary": {
    "total": 0,
    "dsrc": 0,
    "cv2x": 0,
    "bsm": 0,
    "spat": 0,
    "rsa": 0,
    "cam": 0,
    "denm": 0,
    "cpm": 0,
    "pc5_interface": 0,
    "uu_interface": 0
  },
  "security_summary": {
    "total": 0,
    "valid_signature": 0,
    "invalid_signature": 0,
    "high_trust_level": 0,
    "low_trust_level": 0,
    "detected_anomalies": 0,
    "high_confidence_anomaly": 0
  },
  "anomaly_summary": {
    "total": 0,
    "position_jump": 0,
    "speed_jump": 0,
    "heading_jump": 0,
    "high_frequency": 0,
    "conflicting_alerts": 0,
    "timing_anomaly": 0,
    "other": 0
  },
  "vehicle_locations": [],
  "active_alerts": []
}
```

### Error Response (500 Internal Server Error):
```json
{
  "error": "Failed to get V2X summary: database connection error"
}
```

## Testing Sequence

1. **Test with Default Time Range**: Basic functionality test
2. **Test with Different Time Ranges**: Verify time filtering works
3. **Test After Generating Data**: Use POST /test/v2x-rules first to create test data
4. **Test Error Handling**: Verify graceful error responses