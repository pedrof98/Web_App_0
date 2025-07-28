# Step-by-Step Debug Process for GET /alerts

## Prerequisites
1. Database with alert data (run `POST /test/v2x-rules` first to generate alerts)
2. Understanding that alerts are created by the rule engine
3. Knowledge of GORM preloading and relationships

## Debug Session Steps

### Step 1: Generate Test Alerts (Recommended)
1. Run `POST /test/v2x-rules` to create security events and alerts
2. Verify alerts exist: `SELECT COUNT(*) FROM alerts;`
3. Check alert relationships: `SELECT a.id, r.name, se.message FROM alerts a JOIN rules r ON a.rule_id = r.id JOIN security_events se ON a.security_event_id = se.id;`

### Step 2: Launch Debugger and Send Request
1. Use same VS Code debug configuration
2. Set breakpoints from the alerts breakpoints guide
3. Start debugging
4. Send GET request to `http://localhost:8080/alerts`

### Step 3: Debug Flow Analysis

#### Breakpoint 1: Alert Handler Entry
- **Variable to inspect**: `h`, `c`
- **What to check**: Handler initialization with DB, NotificationManager, ESService
- **Expected**: Valid AlertHandler with all services initialized

#### Breakpoint 2: Pagination Parameter Extraction
- **Variable to inspect**: `page`, `pageSize`, `offset`
- **What to check**: Parameter parsing and defaults
- **Expected Values**:
  ```
  No params: page=1, pageSize=50, offset=0
  ?page=2&pagesize=20: page=2, pageSize=20, offset=20
  ?page=3: page=3, pageSize=50, offset=100
  ```

#### Breakpoint 3: Filter Parameter Processing
- **Variable to inspect**: `severity`, `status`
- **What to check**: Query parameter extraction
- **Expected Values**:
  ```
  No filters: severity="", status=""
  ?severity=critical: severity="critical", status=""
  ?status=open: severity="", status="open"
  ?severity=high&status=open: severity="high", status="open"
  ```

#### Breakpoint 4: Query Builder Creation
- **Variable to inspect**: `query` (GORM query object)
- **What to check**: Base query with preloads
- **Expected**: Complex query with three preload relationships:
  ```
  - Preload("Rule") - Load associated rule
  - Preload("SecurityEvent") - Load associated security event  
  - Preload("SecurityEvent.LogSource") - Load nested log source
  ```

#### Breakpoint 5: Severity Filter Application
- **Variable to inspect**: `query` after filter
- **What to check**: Conditional WHERE clause addition
- **Debug tip**: Print `query.Statement.SQL` to see generated SQL
- **Expected**: If severity provided, WHERE clause added

#### Breakpoint 6: Status Filter Application
- **Variable to inspect**: `query` after both filters
- **What to check**: Multiple WHERE clauses chained
- **Expected**: Both severity AND status filters if provided

#### Breakpoint 7: Ordering and Sorting
- **Variable to inspect**: `query` with ORDER BY
- **What to check**: Sorting by timestamp descending
- **Expected**: Most recent alerts first

#### Breakpoint 8: Total Count Query
- **Variable to inspect**: `total`
- **What to check**: Total count for pagination (before LIMIT/OFFSET)
- **Expected**: Number representing total alerts matching filters
- **Key insight**: Count executed BEFORE pagination to get accurate total

#### Breakpoint 9: Paginated Results Query
- **Variable to inspect**: `alerts` array, `err`
- **What to check**: Final database query execution with all relationships loaded
- **Expected**: Array of Alert objects with Rule, SecurityEvent, and LogSource fully populated
- **Key insight**: This is where the complex JOIN happens

#### Breakpoint 10: Response Construction
- **Variable to inspect**: Final response structure
- **What to check**: JSON response with data and pagination metadata
- **Expected**:
  ```json
  {
    "data": [...], // Alert objects with all relationships
    "pagination": {
      "page": 1,
      "pageSize": 50, 
      "total": 87,
      "pages": 2
    }
  }
  ```

### Step 4: Deep Relationship Analysis

#### Understanding the Alert Data Model:
```
Alert
├── Rule (rule_id → rules.id)
├── SecurityEvent (security_event_id → security_events.id)
    └── LogSource (log_source_id → log_sources.id)
```

#### Key Relationship Debugging:
- **Check Foreign Keys**: Verify alert.rule_id matches actual rule IDs
- **Verify Preloading**: Ensure Rule object is fully populated
- **Nested Preloading**: SecurityEvent.LogSource should be loaded
- **Null Handling**: assigned_to and assigned_user can be null

### Step 5: Query Performance Analysis

#### Generated SQL Analysis:
```sql
-- Count query (for pagination total)
SELECT COUNT(*) FROM alerts 
WHERE severity = 'critical' AND status = 'open';

-- Main query with JOINs (simplified)
SELECT alerts.*, rules.*, security_events.*, log_sources.*
FROM alerts
LEFT JOIN rules ON alerts.rule_id = rules.id
LEFT JOIN security_events ON alerts.security_event_id = security_events.id  
LEFT JOIN log_sources ON security_events.log_source_id = log_sources.id
WHERE severity = 'critical' AND status = 'open'
ORDER BY timestamp DESC
LIMIT 50 OFFSET 0;
```

#### Performance Considerations:
- **Index Usage**: Check if indexes on severity, status, timestamp are used
- **JOIN Performance**: Monitor query execution time for complex joins
- **Preloading vs N+1**: GORM preloading prevents N+1 query problem

### Step 6: Response Validation

#### Data Integrity Checks:
- **Relationship Consistency**: Alert severity matches rule severity
- **Timestamp Logic**: Alert timestamp matches security event timestamp  
- **Status Validity**: Only valid AlertStatus values present
- **Pagination Math**: Verify pages calculation: `(total + pageSize - 1) / pageSize`

#### Real-world Alert Examples:
```json
{
  "id": 123,
  "rule": {"name": "V2X Position Jump Detection"},
  "security_event": {
    "message": "Position jump anomaly detected in vehicle VEH-12345",
    "log_source": {"name": "v2x-simulator"}
  },
  "severity": "high",
  "status": "open"
}
```

## Key Learning Points

### SIEM Alert Management:
1. **Alert Lifecycle**: open → in_progress → closed/false_positive
2. **Severity Prioritization**: critical > high > medium > low > info
3. **Rule-Event Relationship**: Each alert links to triggering rule and event
4. **Assignment Workflow**: Alerts can be assigned to security analysts

### Advanced GORM Usage:
1. **Complex Preloading**: Multiple nested relationships in single query
2. **Conditional Queries**: Dynamic WHERE clause building
3. **Pagination Pattern**: Count + Limit/Offset combination
4. **Query Chaining**: Building queries step by step

### Security Operations:
1. **Alert Triage**: Filtering by severity and status for workflow management
2. **Investigation Context**: Full event and rule context for each alert
3. **Performance at Scale**: Pagination for handling large alert volumes

## Common Issues and Debugging

### Issue: No Alerts Returned
- **Check**: Alerts exist in database (`SELECT COUNT(*) FROM alerts`)
- **Verify**: Rule engine has been triggered (`POST /test/v2x-rules`)
- **Solution**: Generate test data first

### Issue: Missing Relationships (Rule/SecurityEvent null)
- **Check**: Foreign key constraints and data integrity
- **Verify**: Preload statements are correct
- **Solution**: Check for orphaned records or constraint violations

### Issue: Slow Query Performance
- **Check**: Database indexes on filtered columns (severity, status, timestamp)
- **Verify**: Query execution plan using EXPLAIN
- **Solution**: Add indexes or optimize query structure

### Issue: Pagination Inconsistencies
- **Check**: Total count calculation vs actual results
- **Verify**: Offset calculation: `(page - 1) * pageSize`
- **Solution**: Verify pagination math and parameter validation

## Advanced Debugging Tips

1. **SQL Logging**: Enable GORM SQL logging to see generated queries
2. **Relationship Verification**: Use database tools to verify foreign key relationships
3. **Performance Monitoring**: Time query execution and identify bottlenecks
4. **Data Validation**: Cross-reference alert data with source security events
5. **Edge Case Testing**: Test with empty results, large page sizes, invalid parameters

## Alert Status Transitions

Understanding alert workflow for debugging:
```
open → in_progress → closed
     → false_positive
```

This helps debug assignment logic and status update workflows in related endpoints like `PUT /alerts/:id`.