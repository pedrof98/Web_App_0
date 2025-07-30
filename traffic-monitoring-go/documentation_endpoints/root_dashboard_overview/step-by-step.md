# Step-by-Step Debug Process for GET /dashboard/overview

## Prerequisites
1. Database with security events and alerts (run `POST /test/v2x-rules` to generate test data)
2. Understanding that the dashboard aggregates data from multiple SIEM components
3. Knowledge that this is a composite endpoint calling multiple service methods
4. Awareness that response time depends on database size and time range scope

## Debug Session Steps

### Step 1: Generate Test Data for Dashboard (Recommended)
1. Run `POST /test/v2x-rules` to create security events and alerts
2. Run `POST /ingest` with multiple V2X payloads for variety
3. Verify data exists:
   ```sql
   SELECT COUNT(*) FROM security_events;
   SELECT COUNT(*) FROM alerts;
   SELECT severity, COUNT(*) FROM security_events GROUP BY severity;
   SELECT status, COUNT(*) FROM alerts GROUP BY status;
   ```
4. Ensure data spans different time ranges for meaningful aggregation

### Step 2: Launch Debugger and Send Request
1. Use same VS Code debug configuration
2. Set breakpoints from the dashboard overview breakpoints guide
3. Start debugging
4. Send GET request to `http://localhost:8080/dashboard/overview`

### Step 3: Debug Flow Analysis - Handler Layer

#### Breakpoint 1: Dashboard Handler Entry
- **Variable to inspect**: `h`, `c`
- **What to check**: Handler initialization with DashboardService and optional ESService
- **Expected**: Valid DashboardHandler with database connection

#### Breakpoint 2: Time Range Parameter Processing
- **Variable to inspect**: `timeRange`
- **What to check**: Query parameter extraction and default assignment
- **Expected Values**:
  ```
  No param: timeRange="last_30_days"
  ?timeRange=today: timeRange="today"
  ?timeRange=last_hour: timeRange="last_hour"
  ?timeRange=invalid: timeRange="invalid" (will use default in service)
  ```

### Step 4: Debug Flow Analysis - Event Summary

#### Breakpoint 3: Event Summary Service Call
- **Variable to inspect**: `h.DashboardService.GetEventSummary(timeRange)` call
- **What to check**: Service delegation for event aggregation
- **Expected**: Service call with time range parameter

#### Breakpoint 4: Event Summary Implementation
- **Variable to inspect**: `s.DB`, `timeRange` in service
- **What to check**: Service receives correct parameters
- **Expected**: Valid GORM database connection and time range string

#### Breakpoint 5: Time Filter Generation
- **Variable to inspect**: `timeFilter` string
- **What to check**: Time range conversion to SQL WHERE clause
- **Expected Time Filters**:
  ```sql
  "last_30_days": "created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)"
  "today": "created_at >= DATE(NOW())"
  "last_hour": "created_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR)"
  "": "" (empty for no time filter)
  ```

#### Breakpoint 6: Total Events Count
- **Variable to inspect**: `summary.Total`, `err`
- **What to check**: First aggregation query execution
- **Expected**: Non-negative integer count, `err` = nil

#### Breakpoint 7: Severity Breakdown Queries
- **Variable to inspect**: `summary.Critical`, `summary.High`, etc.
- **What to check**: Individual severity count queries
- **Expected**: Count breakdown that sums to total
- **Key insight**: Multiple sequential database queries for each severity level

### Step 5: Debug Flow Analysis - Alert Summary

#### Breakpoint 8: Alert Summary Service Call
- **Variable to inspect**: `alertSummary`, `err` from service call
- **What to check**: Alert aggregation service execution
- **Expected**: Valid AlertSummary structure or error

#### Breakpoint 9: Alert Summary Implementation
- **Variable to inspect**: Query building in alert service
- **What to check**: Similar pattern to event summary but for alerts table
- **Expected**: GORM query targeting alerts table with time filter

#### Breakpoint 10: Alert Status Breakdown
- **Variable to inspect**: `summary.Open`, `summary.InProgress`, etc.
- **What to check**: Alert status count aggregation
- **Expected Alert Status Distribution**:
  ```
  summary.Open: New alerts needing attention
  summary.InProgress: Currently being investigated
  summary.Closed: Resolved alerts
  summary.FalsePositive: Benign alerts
  Total = Open + InProgress + Closed + FalsePositive
  ```

### Step 6: Debug Flow Analysis - Time Series Data

#### Breakpoint 11: Event Time Series Service Call
- **Variable to inspect**: `eventTimeSeries`, `err`
- **What to check**: Time-based aggregation service call
- **Expected**: TimeSeriesData structure with labels and data arrays

#### Breakpoint 12: Time Series Implementation
- **Variable to inspect**: `groupBy`, `timeFormat`, `result`
- **What to check**: SQL date formatting and grouping logic
- **Expected**: 
  ```go
  groupBy = "day" (hardcoded in handler)
  timeFormat = "date_format(timestamp, '%Y-%m-%d')" for daily grouping
  result = array of {TimeGroup, Count} structs
  ```

#### Breakpoint 13: Time Format Selection
- **Variable to inspect**: `timeFormat` variable
- **What to check**: SQL date format string based on groupBy parameter
- **Expected Formats**:
  ```sql
  "hour": "date_format(timestamp, '%Y-%m-%d %H:00')"
  "day": "date_format(timestamp, '%Y-%m-%d')"
  "week": "date_format(date_sub(timestamp, interval weekday(timestamp) day), '%Y-%m-%d')"
  "month": "date_format(timestamp, '%Y-%m')"
  ```

### Step 7: Debug Flow Analysis - Top Sources and Rules

#### Breakpoint 14: Top Source IPs Service Call
- **Variable to inspect**: `topSources`, `err`
- **What to check**: Source IP aggregation with limit=5
- **Expected**: Array of {source_ip, count} objects, max 5 entries

#### Breakpoint 15: Top Rules Service Call
- **Variable to inspect**: `topRules`, `err`
- **What to check**: Rule trigger frequency aggregation
- **Expected**: Array of {rule_name, count} objects for most triggered rules

### Step 8: Understanding Aggregation Queries

#### Event Summary SQL Pattern:
```sql
-- Base query with time filter
SELECT COUNT(*) FROM security_events 
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY);

-- Severity breakdown (executed 5 times)
SELECT COUNT(*) FROM security_events 
WHERE severity = 'critical' AND created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY);
```

#### Time Series SQL Pattern:
```sql
-- Daily grouping example
SELECT DATE_FORMAT(timestamp, '%Y-%m-%d') as time_group, COUNT(*) as count
FROM security_events 
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY time_group 
ORDER BY time_group;
```

#### Top Sources SQL Pattern:
```sql
-- Most active sources
SELECT source_ip, COUNT(*) as count
FROM security_events 
WHERE source_ip IS NOT NULL 
  AND source_ip != '' 
  AND created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY source_ip 
ORDER BY count DESC 
LIMIT 5;
```

### Step 9: Response Construction and Validation

#### Breakpoint 16: Final Response Assembly
- **Variable to inspect**: Final `gin.H` response object
- **What to check**: All five data sections properly assembled
- **Expected Structure**:
  ```json
  {
    "event_summary": {...},     // EventCountSummary
    "alert_summary": {...},     // AlertSummary  
    "event_time_series": {...}, // TimeSeriesData
    "top_sources": [...],       // Array of source objects
    "top_rules": [...]          // Array of rule objects
  }
  ```

#### Data Consistency Validation:
- **Event Totals**: Severity counts should sum to total
- **Alert Totals**: Status counts should sum to total
- **Time Series**: Labels and data arrays should have equal length
- **Top Lists**: Should be sorted by count in descending order

### Step 10: Performance Analysis

#### Query Performance Monitoring:
Monitor execution time for each service call:
```go
// Watch for performance bottlenecks
start := time.Now()
eventSummary, err := h.DashboardService.GetEventSummary(timeRange)
eventSummaryTime := time.Since(start)

start = time.Now()
alertSummary, err := h.DashboardService.GetAlertSummary(timeRange)
alertSummaryTime := time.Since(start)
// ... continue for other service calls
```

#### Expected Performance Characteristics:
```
Event Summary: 10-50ms (depends on table size)
Alert Summary: 5-25ms (typically smaller than events)
Time Series: 20-100ms (complex GROUP BY query)
Top Sources: 15-75ms (GROUP BY + ORDER BY + LIMIT)
Top Rules: 20-80ms (JOIN with rules table)
Total Response Time: 70-330ms
```

### Step 11: Error Handling Analysis

#### Service-Level Error Propagation:
Each service call can fail independently:
```go
// Individual error handling pattern
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": "Failed to get event summary: " + err.Error()
    })
    return
}
```

#### Common Error Scenarios:
- **Database Connection Errors**: Service unavailable
- **Table Missing Errors**: Schema migration issues
- **Query Timeout Errors**: Large dataset aggregations
- **Data Type Errors**: Unexpected NULL values or data formats
- **Time Range Parsing Errors**: Invalid date calculations

### Step 12: Time Range Impact Analysis

#### Understanding Time Range Effects on Data:

**Recent Time Ranges (`last_hour`, `today`)**:
- **Expected Volume**: Lower event counts, focused on immediate activity
- **Performance**: Faster queries due to smaller dataset
- **Use Case**: Real-time monitoring, immediate threat detection
- **Debug Focus**: Verify recent events are being captured

**Medium Time Ranges (`last_7_days`, `last_30_days`)**:
- **Expected Volume**: Moderate to high event counts
- **Performance**: Moderate query times, most common use case
- **Use Case**: Weekly/monthly security reviews
- **Debug Focus**: Check time series trends and pattern analysis

**Historical Time Ranges (`this_year`, `last_month`)**:
- **Expected Volume**: Large datasets, comprehensive analysis
- **Performance**: Slower queries, potential timeout issues
- **Use Case**: Compliance reporting, long-term trend analysis
- **Debug Focus**: Monitor query performance and memory usage

### Step 13: V2X-Specific Data Analysis

#### V2X Event Patterns in Dashboard:
When debugging with V2X data, expect these patterns:

**Event Summary Distribution**:
```json
{
  "total": 15420,
  "info": 12008,    // Normal V2X BSM/DENM messages (majority)
  "low": 2341,      // BSM timing violations, minor issues
  "medium": 892,    // Speed anomalies, message frequency issues
  "high": 156,      // Position jumps, potential GPS spoofing
  "critical": 23    // Invalid signatures, active attacks
}
```

**Alert Summary for V2X**:
```json
{
  "total": 234,
  "open": 45,           // New V2X threats needing investigation
  "in_progress": 12,    // Active V2X incident response
  "closed": 165,        // Resolved V2X security issues
  "false_positive": 12  // Benign V2X anomalies
}
```

**Time Series Patterns**:
- **Peak Hours**: Higher V2X activity during commute times
- **Weekday vs Weekend**: Different traffic patterns
- **Attack Patterns**: Concentrated bursts during specific time periods

### Step 14: Database Query Optimization

#### Query Performance Debugging:

**Slow Event Summary Queries**:
```sql
-- Check if indexes exist for performance
SHOW INDEXES FROM security_events;

-- Verify time-based index usage
EXPLAIN SELECT COUNT(*) FROM security_events 
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY);

-- Check severity index usage  
EXPLAIN SELECT COUNT(*) FROM security_events 
WHERE severity = 'critical' AND created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY);
```

**Time Series Query Optimization**:
```sql
-- Monitor GROUP BY performance
EXPLAIN SELECT DATE_FORMAT(timestamp, '%Y-%m-%d') as time_group, COUNT(*) 
FROM security_events 
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
GROUP BY time_group;
```

**Top Sources Query Performance**:
```sql
-- Check source_ip index and NULL handling
EXPLAIN SELECT source_ip, COUNT(*) as count
FROM security_events 
WHERE source_ip IS NOT NULL AND source_ip != ''
GROUP BY source_ip ORDER BY count DESC LIMIT 5;
```

### Step 15: Data Consistency Validation

#### Cross-Service Data Validation:
Debug data consistency across different dashboard components:

**Event-Alert Correlation**:
```sql
-- Verify alert generation rate
SELECT 
  (SELECT COUNT(*) FROM alerts WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)) as alerts,
  (SELECT COUNT(*) FROM security_events WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)) as events,
  ROUND(
    (SELECT COUNT(*) FROM alerts WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)) * 100.0 / 
    (SELECT COUNT(*) FROM security_events WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)), 2
  ) as alert_rate_percent;
```

**Time Series Data Validation**:
```sql
-- Verify time series totals match summary totals
SELECT SUM(daily_count) as time_series_total
FROM (
  SELECT DATE_FORMAT(timestamp, '%Y-%m-%d') as day, COUNT(*) as daily_count
  FROM security_events 
  WHERE created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
  GROUP BY day
) as daily_counts;
```

### Step 16: Advanced Debugging Techniques

#### Service Call Timing Analysis:
```go
// Add timing instrumentation to debug performance
func (h *DashboardHandler) GetDashboardOverview(c *gin.Context) {
    start := time.Now()
    defer func() {
        total := time.Since(start)
        log.Printf("Dashboard overview total time: %v", total)
    }()
    
    timeRange := c.DefaultQuery("timeRange", "last_30_days")
    
    // Event summary timing
    eventStart := time.Now()
    eventSummary, err := h.DashboardService.GetEventSummary(timeRange)
    log.Printf("Event summary time: %v", time.Since(eventStart))
    
    // Alert summary timing  
    alertStart := time.Now()
    alertSummary, err := h.DashboardService.GetAlertSummary(timeRange)
    log.Printf("Alert summary time: %v", time.Since(alertStart))
    
    // Continue for other service calls...
}
```

#### Memory Usage Monitoring:
```go
// Monitor memory usage for large aggregations
func (s *DashboardService) GetEventTimeSeries(timeRange string, groupBy string) (*TimeSeriesData, error) {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    startMem := m.Alloc
    
    defer func() {
        runtime.ReadMemStats(&m)
        memUsed := m.Alloc - startMem
        log.Printf("Time series memory used: %d KB", memUsed/1024)
    }()
    
    // Query execution...
}
```

#### Database Connection Pool Monitoring:
```go
// Check database connection health
func (h *DashboardHandler) checkDBHealth() {
    db, err := h.DB.DB()
    if err != nil {
        log.Printf("DB error: %v", err)
        return
    }
    
    stats := db.Stats()
    log.Printf("DB Stats - Open: %d, InUse: %d, Idle: %d", 
        stats.OpenConnections, stats.InUse, stats.Idle)
}
```

### Step 17: Real-World Debugging Scenarios

#### Scenario 1: Dashboard Loading Slowly
**Debug Steps**:
1. Check time range scope (shorter ranges = faster queries)
2. Monitor individual service call times
3. Verify database indexes on filtered columns
4. Check for database connection pool exhaustion
5. Consider data volume and query complexity

#### Scenario 2: Inconsistent Data Between Calls
**Debug Steps**:
1. Verify time range parameter consistency
2. Check for database writes between service calls
3. Validate transaction isolation levels
4. Ensure consistent time filtering across services
5. Check for clock synchronization issues

#### Scenario 3: Missing V2X Data in Dashboard
**Debug Steps**:
1. Verify V2X collectors are running (`GET /collectors`)
2. Check recent V2X events exist (`GET /security-events?category=v2x`)
3. Validate time range covers V2X activity period
4. Ensure V2X rules are generating alerts
5. Check source IP filtering for V2X sources

#### Scenario 4: Memory Issues with Large Datasets
**Debug Steps**:
1. Monitor query result set sizes
2. Check for efficient aggregation vs. full table scans
3. Implement query result streaming for large datasets
4. Consider pagination for time series data
5. Optimize GROUP BY and ORDER BY operations

### Step 18: Integration Testing

#### Dashboard Data Flow Validation:
```
Collectors → Security Events → Rules → Alerts → Dashboard Aggregation
    ↓              ↓             ↓        ↓              ↓
  Running      Event Count   Rule Count Alert Count  Dashboard Stats
```

#### End-to-End Testing Sequence:
1. **Start Collectors**: `POST /collectors/start-all`
2. **Generate Events**: Send V2X messages to collectors
3. **Verify Events**: `GET /security-events` (check recent events)
4. **Verify Alerts**: `GET /alerts` (check rule-generated alerts)
5. **Test Dashboard**: `GET /dashboard/overview` (verify aggregated data)
6. **Validate Consistency**: Cross-check counts across endpoints

#### Performance Benchmarking:
```bash
# Test dashboard performance under load
for i in {1..10}; do
  curl -w "Time: %{time_total}s\n" \
       -o /dev/null -s \
       "http://localhost:8080/dashboard/overview?timeRange=last_30_days"
done
```

## Key Learning Points

### SIEM Dashboard Architecture:
1. **Composite Endpoint**: Aggregates data from multiple service methods
2. **Service-Oriented Design**: Each component (events, alerts, time series) handled separately
3. **Error Isolation**: Individual service failures don't crash entire dashboard
4. **Performance Considerations**: Multiple database queries require optimization

### Aggregation Patterns:
1. **Count Summaries**: Total and categorical breakdowns
2. **Time Series Analysis**: Temporal trend visualization
3. **Top Lists**: Ranked analysis of sources and rules
4. **Cross-Service Correlation**: Events-to-alerts conversion tracking

### V2X Security Context:
1. **Real-Time Monitoring**: Dashboard shows current V2X threat landscape
2. **Attack Detection**: Critical/high events indicate active V2X attacks
3. **Traffic Analysis**: Event volume patterns reveal V2X communication health
4. **Rule Effectiveness**: Top rules show which V2X threats are most common

## Common Issues and Debugging

### Issue: Dashboard Returns Empty Data
- **Check**: Security events and alerts exist in database
- **Verify**: Time range covers period with actual data
- **Solution**: Generate test data with `POST /test/v2x-rules`

### Issue: Slow Dashboard Response
- **Check**: Database indexes on timestamp, severity, status columns
- **Verify**: Query execution plans using EXPLAIN
- **Solution**: Optimize queries or reduce time range scope

### Issue: Inconsistent Totals
- **Check**: Time range parameter consistency across service calls
- **Verify**: Data integrity between events and alerts tables
- **Solution**: Ensure referential integrity and consistent time filtering

### Issue: Time Series Data Gaps
- **Check**: Continuous data ingestion and event generation
- **Verify**: Date formatting and grouping logic
- **Solution**: Validate collector uptime and event creation timestamps

This dashboard endpoint provides the **central monitoring view** for the entire V2X SIEM system, aggregating security intelligence from all system components into actionable insights for security operations teams.