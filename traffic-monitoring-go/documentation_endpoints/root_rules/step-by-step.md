# Step-by-Step Debug Process for GET /rules

## Prerequisites
1. Database with default V2X rules (automatically created on startup)
2. Understanding that rules are the "brain" of threat detection
3. Knowledge that rules evaluate security events to create alerts

## Debug Session Steps

### Step 1: Verify Default Rules Exist
Before debugging, confirm rules are in database:
```sql
SELECT id, name, category, status FROM rules WHERE category = 'v2x';
```
Should show 6+ V2X security rules created automatically.

### Step 2: Launch Debugger and Send Request
1. Use same VS Code debug configuration  
2. Set breakpoints from the rules breakpoints guide
3. Start debugging
4. Send GET request to `http://localhost:8080/rules`

### Step 3: Debug Flow Analysis

#### Breakpoint 1: Rule Handler Entry
- **Variable to inspect**: `h`, `c`
- **What to check**: Handler initialization with database connection
- **Expected**: Valid RuleHandler with DB connection

#### Breakpoint 2: Status Filter Parameter
- **Variable to inspect**: `status`
- **What to check**: Query parameter extraction for status filtering
- **Expected Values**:
  ```
  No param: status=""
  ?status=enabled: status="enabled"
  ?status=disabled: status="disabled"
  ?status=testing: status="testing"
  ```

#### Breakpoint 3: Category Filter Parameter
- **Variable to inspect**: `category`
- **What to check**: Query parameter for category filtering
- **Expected Values**:
  ```
  No param: category=""
  ?category=v2x: category="v2x"
  ?category=network: category="network"
  ```

#### Breakpoint 4: Query Builder Creation
- **Variable to inspect**: `query` (GORM query object)
- **What to check**: Base query initialization
- **Expected**: GORM query targeting rules table

#### Breakpoint 5: Status Filter Application
- **Variable to inspect**: `query` after status filter
- **What to check**: Conditional WHERE clause addition
- **Expected**: If status provided, WHERE status = ? clause added

#### Breakpoint 6: Category Filter Application
- **Variable to inspect**: `query` after category filter
- **What to check**: Second WHERE clause chaining
- **Expected**: Additional WHERE category = ? if provided

#### Breakpoint 7: Rule Ordering
- **Variable to inspect**: `query` with ORDER BY
- **What to check**: Alphabetical sorting by rule name
- **Expected**: ORDER BY name ASC clause

#### Breakpoint 8: Database Query Execution
- **Variable to inspect**: `rules` array, `err`
- **What to check**: Final rule retrieval from database
- **Expected**: Array of Rule objects matching filters
- **Key insight**: This is where the actual database query executes

#### Breakpoint 9: Response Construction
- **Variable to inspect**: Final JSON response
- **What to check**: Rule array being returned to client
- **Expected**: JSON array of rule objects with complete metadata

### Step 4: Rule Analysis and Understanding

#### V2X Rule Examination:
When you inspect the `rules` array, you should see these critical V2X rules:

1. **Position Jump Detection**:
   ```json
   {
     "name": "V2X Position Jump Detection",
     "condition": "category = v2x AND raw_data.anomalies contains position_jump",
     "severity": "high"
   }
   ```

2. **Invalid Signature Detection**:
   ```json
   {
     "name": "V2X Invalid Digital Signature", 
     "condition": "category = v2x AND raw_data.signature_valid = false",
     "severity": "critical"
   }
   ```

3. **Message Flooding Detection**:
   ```json
   {
     "name": "V2X Message Flooding Attack",
     "condition": "category = v2x AND raw_data.anomalies contains high_frequency",
     "severity": "high"
   }
   ```

#### Rule Condition Analysis:
Each rule's `condition` field contains the logic for threat detection:
- **Category Check**: `category = v2x` (must be V2X event)
- **Anomaly Detection**: `raw_data.anomalies contains position_jump`
- **Confidence Threshold**: `confidence > 0.7` (high confidence)
- **Direct Field Check**: `signature_valid = false` (explicit values)

### Step 5: Understanding Rule-to-Alert Flow

#### How Rules Generate Alerts:
1. **Security Event Created** → (`POST /ingest`)
2. **Rule Engine Evaluates** → (Enhanced Rule Engine)
3. **Condition Matching** → (Each rule's condition tested)
4. **Alert Generation** → (If rule matches, alert created)
5. **Alert Management** → (`GET /alerts`)

#### Rule Engine Connection:
The rules you see here are the exact same ones used by:
- `EnhancedRuleEngine.EvaluateEvent()` 
- `POST /test/v2x-rules` (rule testing)
- Alert generation pipeline

### Step 6: Cross-Reference with Other Endpoints

#### Related Endpoint Testing:
1. **Compare with Rule Examples**: `GET /test/v2x-rules/examples`
2. **Test Rule Execution**: `POST /test/v2x-rules`
3. **View Generated Alerts**: `GET /alerts` 
4. **See Rule Performance**: `GET /dashboard/alerts/top-rules`

#### Database Validation:
```sql
-- Verify rule exists
SELECT * FROM rules WHERE name = 'V2X Position Jump Detection';

-- Check rule effectiveness
SELECT r.name, COUNT(a.id) as alert_count 
FROM rules r 
LEFT JOIN alerts a ON r.id = a.rule_id 
WHERE r.category = 'v2x' 
GROUP BY r.id, r.name;

-- Find most triggered rules
SELECT r.name, r.severity, COUNT(a.id) as alerts
FROM rules r
JOIN alerts a ON r.id = a.rule_id  
GROUP BY r.id
ORDER BY alerts DESC;
```

## Key Learning Points

### SIEM Rule Management:
1. **Rule as Code**: Conditions written in structured query language
2. **Dynamic Filtering**: Status and category-based rule organization
3. **Hierarchical Severity**: Critical > High > Medium > Low > Info
4. **Category Separation**: V2X rules separate from network/system rules

### V2X-Specific Detection Logic:
1. **Anomaly-Based Detection**: Rules check for specific anomaly types
2. **Confidence Thresholds**: Rules require minimum confidence levels
3. **Multi-Factor Conditions**: Rules combine category + specific checks
4. **Real-World Threats**: Each rule targets actual V2X attack vectors

### Rule Engine Architecture:
1. **Condition Parsing**: Complex condition syntax evaluation
2. **Event Matching**: Rules tested against incoming security events
3. **Alert Generation**: Matching rules create alerts automatically
4. **Performance Optimization**: Only enabled rules are evaluated

## Common Issues and Debugging

### Issue: No Rules Returned
- **Check**: Default rules created during startup
- **Verify**: Database migration completed successfully
- **Solution**: Check application startup logs for rule creation

### Issue: Wrong Rules for Category
- **Check**: Category parameter spelling and case
- **Verify**: Available categories in database
- **Solution**: Use exact category names: "v2x", "network", etc.

### Issue: Rules Not Generating Alerts
- **Check**: Rule status is "enabled" not "disabled"
- **Verify**: Rule conditions match incoming event data
- **Solution**: Test rules with `POST /test/v2x-rules`

### Issue: Rule Conditions Unclear
- **Check**: Rule condition syntax and logical operators
- **Verify**: Event data structure matches rule expectations
- **Solution**: Examine raw_data in security events for structure

## Advanced Rule Analysis

### Rule Condition Syntax Patterns:
```sql
-- Basic field comparison
"category = v2x"

-- Nested JSON field access  
"raw_data.signature_valid = false"

-- Array content checking
"raw_data.anomalies contains position_jump"

-- Numeric comparisons
"raw_data.anomalies[0].confidence > 0.7"

-- Complex conditions
"category = v2x AND raw_data.signature_valid = false"
```

### Rule Effectiveness Metrics:
- **True Positive Rate**: How often rule correctly identifies threats
- **False Positive Rate**: How often rule triggers on benign events  
- **Alert Volume**: Total alerts generated by each rule
- **Response Time**: How quickly rules detect attacks

This endpoint gives you complete visibility into the **detection logic** that powers the entire V2X SIEM system!