# Step-by-Step Debug Process for POST /ingest

## Prerequisites
1. Ensure PostgreSQL is running on localhost:5432
2. Ensure Elasticsearch is running on localhost:9200 (optional)
3. Database `v2x_siem` exists with proper schema

## Debug Session Steps

### Step 1: Launch Debugger
1. Open VS Code in your project root
2. Set breakpoints as listed in the breakpoints guide
3. Press F5 or go to Run > Start Debugging
4. Select "Debug V2X SIEM App" configuration
5. Wait for "Starting SIEM server on port 8080..." message

### Step 2: Send Test Request
1. Open Insomnia
2. Create new POST request to `http://localhost:8080/ingest`
3. Add `Content-Type: application/json` header
4. Paste Test Payload 1 (basic V2X event) into body
5. Send the request

### Step 3: Debug Flow Analysis

#### Breakpoint 1: Entry Point
- **Variable to inspect**: `c.Request`
- **What to check**: Request method, headers, content-type
- **Expected**: Method=POST, Content-Type=application/json

#### Breakpoint 2: Body Reading  
- **Variable to inspect**: `body`
- **What to check**: Raw bytes of JSON payload
- **Expected**: JSON string matching your sent payload

#### Breakpoint 3: Transaction Start
- **Variable to inspect**: `h.DB`
- **What to check**: Database connection is valid
- **Expected**: No nil values, connection established

#### Breakpoint 4: Event Ingestion Entry
- **Variable to inspect**: `rawEventData`
- **What to check**: Same as body from step 2
- **Expected**: Byte array with JSON content

#### Breakpoint 5: JSON Unmarshaling
- **Variable to inspect**: `rawEvent`
- **What to check**: All fields properly parsed
- **Expected**: 
  ```
  rawEvent.SourceName = "v2x-simulator"
  rawEvent.Category = "v2x"
  rawEvent.Details["vehicle_id"] = "VEH001"
  ```

#### Breakpoint 6: Log Source Lookup
- **Variable to inspect**: `logSourceID`
- **What to check**: Database ID found or created
- **Expected**: Non-zero integer value

#### Breakpoint 7: Security Event Creation
- **Variable to inspect**: `securityEvent`
- **What to check**: All fields populated correctly
- **Expected**:
  ```
  securityEvent.LogSourceID = [from step 6]
  securityEvent.Category = "v2x"
  securityEvent.RawData = [original JSON]
  ```

#### Breakpoint 8: Rule Engine Call
- **Variable to inspect**: `ruleEngine`
- **What to check**: Rules engine initialized
- **Expected**: Non-nil ruleEngine with DB connection

#### Breakpoint 9: Rule Engine Processing
- **Variable to inspect**: `rules` array
- **What to check**: V2X rules loaded from database
- **Expected**: Array containing position_jump, speed_jump, etc. rules

#### Breakpoint 10: Elasticsearch Indexing
- **Variable to inspect**: `event`, `alertList`
- **What to check**: Final processed data
- **Expected**: Complete security event and any generated alerts

### Step 4: Response Analysis
- **Check Insomnia response**: Should receive 200 OK with event_id
- **Check database**: Query `security_events` table for new record
- **Check alerts**: Query `alerts` table for any generated alerts

## Common Issues and Debugging

### Issue: Database Connection Failed
- **Check**: Environment variables in launch.json
- **Verify**: PostgreSQL service is running
- **Solution**: Update DB credentials in launch configuration

### Issue: JSON Parsing Error
- **Check**: Request body format in Insomnia
- **Verify**: Valid JSON syntax
- **Solution**: Use JSON validator before sending

### Issue: No Rules Triggered
- **Check**: Default rules exist in database
- **Verify**: Rule conditions match event data  
- **Solution**: Run database migrations or create default rules

### Issue: Elasticsearch Errors
- **Check**: Elasticsearch service status
- **Verify**: URL configuration correct
- **Solution**: System continues without ES, logs warnings only