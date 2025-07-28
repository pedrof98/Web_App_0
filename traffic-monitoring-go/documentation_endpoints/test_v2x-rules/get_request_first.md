# Bonus: GET /test/v2x-rules/examples Endpoint

## Quick Test Before Running POST
**URL**: `http://localhost:8080/test/v2x-rules/examples`  
**Method**: GET  
**Body**: None required

### Purpose
This endpoint shows you exactly what V2X security rules exist and provides concrete examples of what triggers each rule.

### Sample Response:
```json
{
  "message": "Concrete examples of V2X security detection rules",
  "total_rules": 4,
  "rules": [
    {
      "id": 1,
      "name": "V2X Position Jump Detection", 
      "description": "Detects vehicles with unrealistic position changes",
      "condition": "category = v2x AND anomalies contains position_jump AND confidence > 0.7",
      "severity": "high",
      "status": "enabled",
      "example": {
        "trigger_condition": "Vehicle moves >100m in <1 second",
        "example_scenario": "Vehicle reports position change from (37.7749, -122.4194) to (37.7759, -122.4194) in 0.5 seconds = ~111m movement",
        "detection_logic": "Distance calculation using haversine formula, time difference analysis"
      }
    },
    {
      "id": 2,
      "name": "V2X Speed Anomaly Detection",
      "condition": "category = v2x AND anomalies contains speed_jump AND confidence > 0.8",
      "example": {
        "trigger_condition": "Speed difference >10 m/s between messages",
        "example_scenario": "Vehicle reports 15 m/s then 30 m/s in next message (15 m/s jump)",
        "detection_logic": "Compare speed values in consecutive BSM messages"
      }
    }
  ],
  "note": "Use POST /test/v2x-rules to trigger these rules with test data"
}
```

### Why Use This First
1. **Verify Rules Exist**: Confirm V2X rules are in database
2. **Understand Conditions**: See exact rule conditions before testing
3. **Learn Examples**: Understand what triggers each rule
4. **Debugging Prep**: Know what to expect when POST test runs