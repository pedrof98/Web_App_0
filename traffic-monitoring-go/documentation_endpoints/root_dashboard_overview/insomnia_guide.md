# Debug Breakpoints Guide for GET /dashboard/overview Endpoint

## Critical Breakpoints (Set these in order of execution)

### 1. **Dashboard Handler Entry Point**
**File**: `app/handlers/dashboard.go`  
**Line**: `func (h *DashboardHandler) GetDashboardOverview(c *gin.Context)`  
**Purpose**: See dashboard overview request and understand aggregation flow

### 2. **Time Range Parameter Extraction**
**File**: `app/handlers/dashboard.go`  
**Line**: `timeRange := c.DefaultQuery("timeRange", "last_30_days")`  
**Purpose**: Watch time range parameter processing and defaults

### 3. **Event Summary Service Call**
**File**: `app/handlers/dashboard.go`  
**Line**: `eventSummary, err := h.DashboardService.GetEventSummary(timeRange)`  
**Purpose**: See first aggregation service call

### 4. **Event Summary Implementation**
**File**: `app/siem/dashboard.go`  
**Line**: `func (s *DashboardService) GetEventSummary(timeRange string) (*EventCountSummary, error)`  
**Purpose**: Watch event count aggregation logic

### 5. **Time Filter Generation**
**File**: `app/siem/dashboard.go`  
**Line**: `timeFilter := getTimeFilter(timeRange)`  
**Purpose**: See how time range strings convert to SQL WHERE clauses

### 6. **Total Events Count Query**
**File**: `app/siem/dashboard.go`  
**Line**: `if err := query.Count(&summary.Total).Error; err != nil {`  
**Purpose**: Watch total event count calculation

### 7. **Severity-based Event Counts**
**File**: `app/siem/dashboard.go`  
**Line**: `if err := query.Where("severity = ?", models.SeverityCritical).Count(&summary.Critical).Error; err != nil {`  
**Purpose**: See severity breakdown queries

### 8. **Alert Summary Service Call**
**File**: `app/handlers/dashboard.go`  
**Line**: `alertSummary, err := h.DashboardService.GetAlertSummary(timeRange)`  
**Purpose**: Watch alert aggregation service call

### 9. **Alert Summary Implementation**
**File**: `app/siem/dashboard.go`  
**Line**: `func (s *DashboardService) GetAlertSummary(timeRange string) (*AlertSummary, error)`  
**Purpose**: See alert count and status aggregation

### 10. **Alert Status Counts**
**File**: `app/siem/dashboard.go`  
**Line**: `if err := query.Where("status = ?", models.AlertStatusOpen).Count(&summary.Open).Error; err != nil {`  
**Purpose**: Watch alert status breakdown queries

### 11. **Event Time Series Service Call**
**File**: `app/handlers/dashboard.go`  
**Line**: `eventTimeSeries, err := h.DashboardService.GetEventTimeSeries(timeRange, "day")`  
**Purpose**: See time series data generation

### 12. **Time Series Implementation**
**File**: `app/siem/dashboard.go`  
**Line**: `func (s *DashboardService) GetEventTimeSeries(timeRange string, groupBy string) (*TimeSeriesData, error)`  
**Purpose**: Watch time-based grouping and aggregation

### 13. **Time Format Selection**
**File**: `app/siem/dashboard.go`  
**Line**: `timeFormat = "date_format(timestamp, '%Y-%m-%d')"`  
**Purpose**: See SQL date formatting based on groupBy parameter

### 14. **Top Source IPs Service Call**
**File**: `app/handlers/dashboard.go`  
**Line**: `topSources, err := h.DashboardService.GetTopSourceIPs(timeRange, 5)`  
**Purpose**: Watch top source IP aggregation

### 15. **Top Rules Service Call**
**File**: `app/handlers/dashboard.go`  
**Line**: `topRules, err := h.DashboardService.GetTopTriggeredRules(timeRange, 5)`  
**Purpose**: See top triggered rules aggregation

### 16. **Response Aggregation**
**File**: `app/handlers/dashboard.go`  
**Line**: `c.JSON(http.StatusOK, gin.H{`  
**Purpose**: Watch final response construction with all aggregated data

## Inspection Variables

At each breakpoint, inspect these key variables:

### Handler Level:
- `h.DashboardService` - Service instance with database connection
- `timeRange` - Time range parameter (e.g., "last_30_days", "last_hour")
- `eventSummary` - Event count aggregation results
- `alertSummary` - Alert count and status aggregation
- `eventTimeSeries` - Time-based event trend data
- `topSources` - Most active source IPs
- `topRules` - Most triggered security rules

### Service Level:
- `s.DB` - Database connection for queries
- `query` - GORM query builder object
- `timeFilter` - SQL WHERE clause for time filtering
- `summary.Total`, `summary.Critical`, etc. - Individual count results
- `result` - Raw query results for time series and top lists

### Data Structures:
- `EventCountSummary` - Event counts by severity
- `AlertSummary` - Alert counts by status and severity
- `TimeSeriesData` - Labels and data arrays for charting
- Top source/rule results - Arrays of count data

## Time Range Values and Filters

### Supported Time Range Parameters:
- `"last_hour"` - Events from the last 60 minutes
- `"today"` - Events from start of current day
- `"yesterday"` - Events from previous day
- `"last_7_days"` - Events from the last week
- `"last_30_days"` - Events from the last month (default)
- `"this_month"` - Events from start of current month
- `"last_month"` - Events from previous month
- `"this_year"` - Events from start of current year

### Generated SQL Time Filters:
```sql
-- For "last_30_days"
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)

-- For "today"  
WHERE created_at >= DATE(NOW())

-- For "last_hour"
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR)
```

## Database Query Patterns

### Event Summary Queries:
```sql
-- Total events
SELECT COUNT(*) FROM security_events WHERE created_at >= ?

-- Events by severity
SELECT COUNT(*) FROM security_events WHERE severity = 'critical' AND created_at >= ?
SELECT COUNT(*) FROM security_events WHERE severity = 'high' AND created_at >= ?
-- ... etc for each severity level
```

### Alert Summary Queries:
```sql
-- Total alerts
SELECT COUNT(*) FROM alerts WHERE created_at >= ?

-- Alerts by status
SELECT COUNT(*) FROM alerts WHERE status = 'open' AND created_at >= ?
SELECT COUNT(*) FROM alerts WHERE status = 'in_progress' AND created_at >= ?
-- ... etc for each status
```

### Time Series Query:
```sql
-- Daily grouping example
SELECT DATE_FORMAT(timestamp, '%Y-%m-%d') as time_group, COUNT(*) as count
FROM security_events 
WHERE created_at >= ?
GROUP BY time_group 
ORDER BY time_group
```

### Top Sources Query:
```sql
-- Most active source IPs
SELECT source_ip, COUNT(*) as count
FROM security_events 
WHERE source_ip IS NOT NULL AND source_ip != '' AND created_at >= ?
GROUP BY source_ip 
ORDER BY count DESC 
LIMIT 5
```

## Response Data Structure Analysis

### Complete Response Format:
```json
{
  "event_summary": {
    "total": 15420,
    "critical": 23,
    "high": 156,
    "medium": 892,
    "low": 2341,
    "info": 12008
  },
  "alert_summary": {
    "total": 234,
    "open": 45,
    "in_progress": 12,
    "closed": 165,
    "false_positive": 12,
    "critical": 8,
    "high": 34,
    "medium": 78,
    "low": 114
  },
  "event_time_series": {
    "labels": ["2025-07-01", "2025-07-02", "2025-07-03"],
    "data": [423, 567, 389]
  },
  "top_sources": [
    {"source_ip": "192.168.1.100", "count": 1234},
    {"source_ip": "10.0.0.50", "count": 987}
  ],
  "top_rules": [
    {"rule_name": "V2X Position Jump Detection", "count": 45},
    {"rule_name": "Invalid Digital Signature", "count": 23}
  ]
}
```

## Error Handling Patterns

### Service-Level Error Propagation:
Each service call can fail independently, with specific error messages:
- `"Failed to get event summary: " + err.Error()`
- `"Failed to get alert summary: " + err.Error()`
- `"Failed to get event time series: " + err.Error()`
- `"Failed to get top sources: " + err.Error()`
- `"Failed to get top rules: " + err.Error()`

### Database Error Types:
- **Connection errors**: Database unavailable
- **Query errors**: Malformed SQL or schema issues
- **Data type errors**: Unexpected data formats
- **Timeout errors**: Long-running aggregation queries