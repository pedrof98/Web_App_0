# Step-by-Step Debug Process for POST /test/v2x-rules

## Prerequisites
1. Same as POST /ingest: PostgreSQL running, database exists
2. **Important**: Default V2X rules must exist in database
3. Understanding of V2X attack patterns helps

## Debug Session Steps

### Step 1: Launch Debugger
1. Use same VS Code debug configuration from previous setup
2. Set breakpoints from the V2X rules breakpoints guide
3. Start debugging and wait for server startup
4. Verify V2X rules exist: `SELECT * FROM rules WHERE category = 'v2x';`

### Step 2: Send Test Request  
1. Open Insomnia
2. Create new POST request to `http://localhost:8080/test/v2x-rules`
3. Add `Content-Type: application/json` header
4. **Body**: Leave empty `{}` or no body at all
5. Send the request

### Step 3: Debug Flow Analysis

#### Breakpoint 1: Test Handler Entry
- **Variable to inspect**: `h` handler
- **What to check**: Handler has DB connection and rule engine
- **Expected**: Non-nil DB, EventIngester, EnhancedRuleEngine

#### Breakpoint 2: Test Events Creation
- **Variable to inspect**: `testEvents` array
- **What to check**: 6 predefined test scenarios
- **Expected**: Array with 6 maps, each containing different attack scenarios
- **Key inspection**:
  ```
  testEvents[0] = Position jump attack
  testEvents[1] = Speed jump attack  
  testEvents[2] = Message flooding attack
  testEvents[3] = Invalid signature attack
  testEvents[4] = Emergency vehicle alert
  testEvents[5] = High priority DENM alert
  ```

#### Breakpoint 3: Test Event Processing Loop
- **Variable to inspect**: `i`, `eventData`
- **What to check**: Current test event being processed
- **Expected**: `i` increments from 0-5, `eventData` contains current test scenario

#### Breakpoint 4: JSON Marshaling
- **Variable to inspect**: `eventJSON`
- **What to check**: Test event converted to JSON bytes
- **Expected**: Valid JSON representing the test event

#### Breakpoint 5: Transaction Start
- **Variable to inspect**: `tx` database transaction
- **What to check**: Transaction-scoped database operations
- **Expected**: Valid transaction object for atomic test processing

#### Breakpoint 6: Event Ingestion within Test
- **Variable to inspect**: `ingester`
- **What to check**: Same ingestion logic as /ingest endpoint
- **Expected**: Event successfully ingested, security event created

#### Breakpoint 7: Rule Engine Evaluation
- **Variable to inspect**: `ruleEngine`, `securityEvent`
- **What to check**: Rule evaluation starts for test event
- **Expected**: Valid security event with category = "v2x"

#### Breakpoint 8: Enhanced Rule Engine Entry
- **Variable to inspect**: `event`, `rules` array
- **What to check**: V2X rules loaded from database
- **Expected**: Array of enabled V2X rules like:
  ```
  - "V2X Position Jump Detection"
  - "V2X Speed Anomaly Detection"  
  - "V2X Message Flooding Attack"
  - "V2X Invalid Digital Signature"
  ```

#### Breakpoint 9: Rule Condition Evaluation
- **Variable to inspect**: `rule.Condition`, `eventData`
- **What to check**: Specific rule condition being tested
- **Expected**: Conditions like:
  ```
  "category = v2x AND anomalies contains position_jump"
  "category = v2x AND signature_valid = false"
  ```

#### Breakpoint 10: V2X-Specific Condition Logic
- **Variable to inspect**: `condition`, `eventData["details"]`
- **What to check**: How V2X-specific logic evaluates
- **Key inspections**:
  - For position_jump: Check `eventData["details"]["anomalies"]` array
  - For signature: Check `eventData["details"]["signature_valid"]` = false
  - For message type: Check `eventData["details"]["message_type"]`

#### Breakpoint 11: Alert Creation
- **Variable to inspect**: `alert` object
- **What to check**: Alert being created for matched rule
- **Expected**: Alert with SecurityEventID, RuleID, and severity

#### Breakpoint 12: Alert Count Tracking
- **Variable to inspect**: `alertCount`, `results`
- **What to check**: How many alerts generated per test event
- **Expected**: 
  ```
  Position jump test -> 1 alert
  Speed jump test -> 1 alert
  Flooding test -> 1 alert
  Invalid signature test -> 1 alert
  Emergency/DENM tests -> 0-1 alerts (depends on rules)
  ```

### Step 4: Response Analysis
- **Check Insomnia response**: Should show test summary
- **Verify alerts created**: `total_alerts_created` should be > 0
- **Database verification**: Query alerts table for new records
- **Rule effectiveness**: See which rules triggered

## Key Learning Points

### Understanding V2X Rule Evaluation:
1. **Category Filtering**: All V2X rules require `category = "v2x"`
2. **Nested JSON Access**: Rules can access `raw_data.details.anomalies`
3. **Array Processing**: Anomalies are arrays that must be searched
4. **Boolean Logic**: Signature validation uses explicit false checking
5. **Complex Conditions**: Rules combine multiple conditions with AND

### V2X Attack Pattern Recognition:
1. **Position Jump**: Distance > threshold in time < threshold
2. **Speed Anomaly**: Speed difference > threshold between messages  
3. **Message Flooding**: Message frequency > threshold per time window
4. **Signature Failure**: signature_valid = false in message details
5. **Emergency Alerts**: Special message types with high priority

## Common Issues and Debugging

### Issue: No Alerts Generated
- **Check**: Default V2X rules exist in database
- **Verify**: Rules are enabled (`status = 'enabled'`)
- **Solution**: Run database migrations to create default rules

### Issue: Rule Condition Not Matching
- **Check**: Event data structure in `eventData["details"]`
- **Verify**: Anomaly types and field names match rule conditions
- **Debug**: Print rule condition and event data for comparison

### Issue: JSON Parsing Errors
- **Check**: Test event structure is valid
- **Verify**: Nested details object exists
- **Solution**: Ensure test events follow expected schema

### Issue: Transaction Failures
- **Check**: Database constraints and foreign keys
- **Verify**: Log source exists for test events
- **Solution**: Test events create valid security events first

## Advanced Debugging Tips

1. **Add Debug Prints**: Print rule conditions and event data matches
2. **Database Queries**: Query rules table to see active V2X rules
3. **Alert Inspection**: Check created alerts for rule ID and severity
4. **Event Timeline**: Trace event from creation → ingestion → rule evaluation → alert creation