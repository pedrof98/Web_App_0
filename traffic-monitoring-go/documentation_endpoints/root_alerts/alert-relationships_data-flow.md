# Alert Relationships and SIEM Workflow

## Complete Data Flow: From Event to Alert

```
V2X Message → Security Event → Rule Evaluation → Alert → Management
```

### 1. **Alert Creation Process** (How alerts are born)
```go
// In rule engine (when rule matches)
alert := models.Alert{
    SecurityEventID: event.ID,        // Links to triggering event
    RuleID:          rule.ID,         // Links to triggered rule  
    Timestamp:       time.Now(),      // When alert was created
    Severity:        rule.Severity,   // Inherited from rule
    Status:          AlertStatusOpen, // Always starts as "open"
}
```

### 2. **Database Relationships**
```sql
-- Core relationships
alerts.rule_id → rules.id
alerts.security_event_id → security_events.id
security_events.log_source_id → log_sources.id

-- Optional relationships  
alerts.assigned_to → users.id (nullable)
```

### 3. **GORM Preloading Deep Dive**

#### What `Preload("Rule")` does:
```sql
-- Without preload (N+1 problem)
SELECT * FROM alerts WHERE ... -- 1 query
SELECT * FROM rules WHERE id = 1 -- N queries (one per alert)
SELECT * FROM rules WHERE id = 2
-- ...

-- With preload (2 queries total)
SELECT * FROM alerts WHERE ... -- 1 query  
SELECT * FROM rules WHERE id IN (1,2,3,4,5...) -- 1 query for all rules
```

#### What `Preload("SecurityEvent.LogSource")` does:
```sql
-- Nested preloading
SELECT * FROM security_events WHERE id IN (...) -- Load events
SELECT * FROM log_sources WHERE id IN (...) -- Load log sources
```

### 4. **Alert Status Workflow**

#### Status Transitions:
```
open (initial) → in_progress (being investigated)
               → closed (resolved)
               → false_positive (not a real threat)
```

#### Real-world Usage:
- **open**: New alerts requiring attention
- **in_progress**: Analyst actively investigating  
- **closed**: Issue resolved or mitigated
- **false_positive**: Rule triggered incorrectly

### 5. **Severity Impact on Operations**

#### Severity Levels (highest to lowest):
1. **critical** - Immediate response required (active attacks)
2. **high** - Important security issues (suspicious behavior)
3. **medium** - Moderate concerns (policy violations)
4. **low** - Minor issues (informational warnings)
5. **info** - Informational only (normal events flagged)

#### V2X-Specific Severity Examples:
```json
{
  "critical": "Invalid digital signatures (active attack)",
  "high": "Position jump anomalies (GPS spoofing)",
  "medium": "Message flooding (potential DoS)",
  "low": "Trust level degradation (aging certificates)",
  "info": "Normal V2X communication patterns"
}
```

### 6. **Why This Endpoint is Critical for SIEM**

#### Security Operations Center (SOC) Workflow:
1. **Alert Queue**: `GET /alerts?status=open` - What needs attention
2. **Priority Triage**: `GET /alerts?severity=critical&status=open` - Urgent items
3. **Assignment**: `PUT /alerts/:id` - Assign to analyst
4. **Investigation**: Full context via preloaded relationships
5. **Resolution**: Update status to closed/false_positive

#### Key SIEM Capabilities Demonstrated:
- **Contextual Information**: Full rule + event + source context
- **Workflow Management**: Status-based filtering and assignment
- **Scalability**: Pagination for large alert volumes
- **Performance**: Efficient querying with relationship preloading

### 7. **Integration with Other Endpoints**

#### Alert Lifecycle Endpoints:
- `GET /alerts` - List and filter alerts (this endpoint)
- `GET /alerts/:id` - Get single alert details
- `PUT /alerts/:id` - Update status, assignment, resolution
- `POST /alerts/:id/notify` - Send notifications

#### Data Source Endpoints:
- `GET /rules` - Manage detection rules
- `GET /security-events` - View source events
- `POST /ingest` - Creates the security events that become alerts

#### Analytics Endpoints:
- `GET /dashboard/overview` - Alert summary statistics
- `GET /v2x-dashboard/overview` - V2X-specific alert analytics

This makes `GET /alerts` the central hub of SIEM operations - it's where security analysis begins and threat response is coordinated.