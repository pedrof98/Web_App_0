# Debug Breakpoints Guide for GET /security-events Endpoint

## Critical Breakpoints (Set these in order of execution)

### 1. **Security Event Handler Entry Point**
**File**: `app/handlers/security_event.go`  
**Line**: `func (h *SecurityEventHandler) GetSecurityEvents(c *gin.Context)`  
**Purpose**: See security event retrieval request and parameter processing

### 2. **Pagination Parameter Extraction**
**File**: `app/handlers/security_event.go`  
**Line**: `page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))`  
**Purpose**: Understand pagination logic and parameter defaults

### 3. **Filter Parameter Processing**
**File**: `app/handlers/security_event.go`  
**Line**: `severity := c.Query("severity")`  
**Purpose**: See how filtering parameters are extracted and processed

### 4. **Category Filter Processing**
**File**: `app/handlers/security_event.go`  
**Line**: `category := c.Query("category")`  
**Purpose**: Watch category-based filtering for event types

### 5. **Query Builder Creation**
**File**: `app/handlers/security_event.go`  
**Line**: `query := h.DB.Model(&models.SecurityEvent{})`  
**Purpose**: Watch GORM query construction (note: NO preloads in this endpoint)

### 6. **Severity Filter Application**
**File**: `app/handlers/security_event.go`  
**Line**: `if severity != "" { query = query.Where("severity = ?", severity) }`  
**Purpose**: See conditional filter application

### 7. **Category Filter Application**
**File**: `app/handlers/security_event.go`  
**Line**: `if category != "" { query = query.Where("category = ?", category) }`  
**Purpose**: Watch multiple filter chaining

### 8. **Ordering and Sorting**
**File**: `app/handlers/security_event.go`  
**Line**: `query = query.Order("timestamp DESC")`  
**Purpose**: Understand default sorting (most recent first)

### 9. **Total Count Query**
**File**: `app/handlers/security_event.go`  
**Line**: `query.Count(&total)`  
**Purpose**: See how pagination total is calculated

### 10. **Paginated Results Query**
**File**: `app/handlers/security_event.go`  
**Line**: `if err := query.Offset(offset).Limit(pageSize).Find(&events).Error; err != nil {`  
**Purpose**: Watch paginated data retrieval without relationship preloading

### 11. **Response Construction**
**File**: `app/handlers/security_event.go`  
**Line**: `c.JSON(http.StatusOK, gin.H{`  
**Purpose**: See final response with data and pagination metadata

## Inspection Variables

At each breakpoint, inspect these key variables:

- `page`, `pageSize`, `offset` - Pagination parameters
- `severity`, `category` - Filter parameters  
- `query` - GORM query object (watch it build up)
- `total` - Total count for pagination
- `events` - Final security event array (NO relationships loaded)
- Response structure with data and pagination

## Key Differences from GET /alerts

### No Relationship Preloading:
Unlike alerts, security events are returned **without** preloaded relationships:
- **Alerts**: `Preload("Rule").Preload("SecurityEvent").Preload("SecurityEvent.LogSource")`
- **Security Events**: No preloads - just the event data

### Database Schema Context:
```sql
-- SecurityEvent model fields that will be populated
security_events.id
security_events.timestamp  
security_events.source_ip
security_events.source_port
security_events.destination_ip
security_events.destination_port
security_events.protocol
security_events.action
security_events.status
security_events.device_id
security_events.log_source_id  -- Foreign key (but LogSource not preloaded)
security_events.severity
security_events.category
security_events.message
security_events.raw_data       -- Original JSON from ingestion
security_events.created_at
```

### Performance Implications:
- **Faster queries** - No JOIN operations
- **Smaller payloads** - Only core event data
- **Higher throughput** - Suitable for bulk event viewing
- **Less context** - LogSource details not included