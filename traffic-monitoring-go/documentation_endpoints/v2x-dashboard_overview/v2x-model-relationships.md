# V2X Data Model and Relationships

## Database Table Relationships

Understanding the data model helps debug the dashboard aggregations:

### Core Tables:

#### 1. **v2x_messages** (Primary table)
```go
type V2XMessage struct {
    ID          uint      `json:"id"`
    Protocol    Protocol  `json:"protocol"`    // "DSRC" or "CV2X"
    MessageType MessageType `json:"message_type"` // "BSM", "SPAT", "RSA", etc.
    Timestamp   time.Time `json:"timestamp"`
    SourceID    string    `json:"source_id"`   // Vehicle ID
    // ... other fields
}
```

#### 2. **v2x_security_info** (1:1 with v2x_messages)
```go
type V2XSecurityInfo struct {
    ID            uint `json:"id"`
    V2XMessageID  uint `json:"v2x_message_id"` // Foreign Key
    SignatureValid bool `json:"signature_valid"`
    TrustLevel    int  `json:"trust_level"`    // 0-10 scale
    // ... other fields
}
```

#### 3. **v2x_anomaly_detections** (1:Many with v2x_messages)
```go
type V2XAnomalyDetection struct {
    ID              uint    `json:"id"`
    V2XMessageID    uint    `json:"v2x_message_id"` // Foreign Key
    AnomalyType     string  `json:"anomaly_type"`   // "position_jump", "speed_jump", etc.
    ConfidenceScore float64 `json:"confidence_score"` // 0.0-1.0
    // ... other fields
}
```

#### 4. **cv2x_messages** (Optional, 1:1 with v2x_messages where protocol="CV2X")
```go
type CV2XMessage struct {
    ID            uint   `json:"id"`
    V2XMessageID  uint   `json:"v2x_message_id"` // Foreign Key
    InterfaceType string `json:"interface_type"`  // "PC5" or "Uu"
    // ... other fields
}
```

## How Dashboard Queries Work

### V2X Summary Query Flow:
1. **Filter by time**: `WHERE timestamp >= NOW() - INTERVAL '1 hour'`
2. **Count by protocol**: `WHERE protocol = 'DSRC'` vs `WHERE protocol LIKE 'cv2x%'`
3. **Count by message type**: `WHERE message_type = 'BSM'`, etc.
4. **Join for CV2X details**: Get interface types (PC5/Uu) from cv2x_messages table

### Security Summary Query Flow:
1. **Get message IDs** from v2x_messages within time range
2. **Join with security info**: `WHERE v2x_message_id IN (message_ids)`
3. **Aggregate by security fields**: Count valid/invalid signatures, trust levels

### Anomaly Summary Query Flow:
1. **Get message IDs** from v2x_messages within time range  
2. **Join with anomaly detections**: `WHERE v2x_message_id IN (message_ids)`
3. **Count by anomaly type**: Group by anomaly_type field

## Key Relationships to Debug:

### Foreign Key Relationships:
- `v2x_security_info.v2x_message_id` → `v2x_messages.id`
- `v2x_anomaly_detections.v2x_message_id` → `v2x_messages.id`
- `cv2x_messages.v2x_message_id` → `v2x_messages.id`

### Common Query Patterns:
```sql
-- Pattern 1: Direct aggregation on main table
SELECT protocol, COUNT(*) FROM v2x_messages 
WHERE timestamp >= ? GROUP BY protocol;

-- Pattern 2: JOIN aggregation for related data
SELECT signature_valid, COUNT(*) 
FROM v2x_messages m
JOIN v2x_security_info s ON m.id = s.v2x_message_id
WHERE m.timestamp >= ? 
GROUP BY signature_valid;

-- Pattern 3: Subquery with IN clause (GORM approach)
SELECT COUNT(*) FROM v2x_security_info 
WHERE v2x_message_id IN (
    SELECT id FROM v2x_messages WHERE timestamp >= ?
) AND signature_valid = true;
```

This understanding helps you debug:
- **Missing counts**: Check if foreign key relationships exist
- **Slow queries**: Verify indexes on join columns
- **Inconsistent data**: Look for orphaned records