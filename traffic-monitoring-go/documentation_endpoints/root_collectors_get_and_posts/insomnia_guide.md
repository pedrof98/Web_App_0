# Insomnia Configuration for Collector Management Endpoints

## Endpoint Overview
- **GET /collectors** - List all collectors and their status
- **POST /collectors/:name/start** - Start a specific collector
- **POST /collectors/:name/stop** - Stop a specific collector
- **POST /collectors/start-all** - Start all collectors
- **POST /collectors/stop-all** - Stop all collectors

## GET /collectors Configuration

### Basic Setup
- **Method**: GET
- **URL**: `http://localhost:8080/collectors`
- **Headers**: None required
- **Authentication**: None required
- **Body**: None required

### Expected Response (200 OK):
```json
[
  {
    "name": "syslog",
    "running": false
  },
  {
    "name": "snmp", 
    "running": false
  },
  {
    "name": "dsrc",
    "running": true
  },
  {
    "name": "cv2x",
    "running": true
  }
]
```

### Error Response (500 Internal Server Error):
```json
{
  "error": "failed to get collector status: collector 'invalid-name' not found"
}
```

## POST /collectors/:name/start Configuration

### Basic Setup
- **Method**: POST
- **URL**: `http://localhost:8080/collectors/{collector_name}/start`
- **Headers**: None required
- **Authentication**: None required
- **Body**: None required

### Valid Collector Names:
- `syslog` - Syslog message collector (UDP port 514)
- `snmp` - SNMP trap collector (UDP port 162)
- `dsrc` - DSRC V2X collector (UDP port 5001)
- `cv2x` - C-V2X collector (UDP port 5002)

### URL Examples:
```
POST http://localhost:8080/collectors/dsrc/start
POST http://localhost:8080/collectors/cv2x/start
POST http://localhost:8080/collectors/syslog/start
POST http://localhost:8080/collectors/snmp/start
```

### Success Response (200 OK):
```json
{
  "message": "Collector started successfully"
}
```

### Error Responses:

#### Collector Not Found (500 Internal Server Error):
```json
{
  "error": "collector 'invalid-collector' not found"
}
```

#### Already Running (500 Internal Server Error):
```json
{
  "error": "failed to start collector 'dsrc': DSRC collector is already running"
}
```

#### Port Binding Error (500 Internal Server Error):
```json
{
  "error": "failed to start collector 'dsrc': failed to listen on UDP port 5001: address already in use"
}
```

## POST /collectors/:name/stop Configuration

### Basic Setup
- **Method**: POST
- **URL**: `http://localhost:8080/collectors/{collector_name}/stop`
- **Headers**: None required
- **Authentication**: None required
- **Body**: None required

### URL Examples:
```
POST http://localhost:8080/collectors/dsrc/stop
POST http://localhost:8080/collectors/cv2x/stop
POST http://localhost:8080/collectors/syslog/stop
POST http://localhost:8080/collectors/snmp/stop
```

### Success Response (200 OK):
```json
{
  "message": "Collector stopped successfully"
}
```

### Error Responses:

#### Collector Not Found (500 Internal Server Error):
```json
{
  "error": "collector 'invalid-collector' not found"
}
```

#### Not Running (500 Internal Server Error):
```json
{
  "error": "failed to stop collector 'dsrc': DSRC collector is not running"
}
```

## POST /collectors/start-all Configuration

### Basic Setup
- **Method**: POST
- **URL**: `http://localhost:8080/collectors/start-all`
- **Headers**: None required
- **Authentication**: None required
- **Body**: None required

### Success Response (200 OK):
```json
{
  "message": "All collectors started"
}
```

### Partial Success (500 Internal Server Error):
```json
{
  "error": "failed to start some collectors"
}
```

**Note**: The system continues starting other collectors even if some fail. Check logs for specific failures.

## POST /collectors/stop-all Configuration

### Basic Setup
- **Method**: POST
- **URL**: `http://localhost:8080/collectors/stop-all`
- **Headers**: None required
- **Authentication**: None required
- **Body**: None required

### Success Response (200 OK):
```json
{
  "message": "All collectors stopped"
}
```

**Note**: Stop-all operation doesn't return errors - it gracefully stops all running collectors.

## Collector Types and Purposes

### V2X Collectors (Vehicle Communication)
- **dsrc**: Captures DSRC (Dedicated Short Range Communications) messages
  - **Port**: 5001 (UDP)
  - **Protocol**: IEEE 802.11p / J2735
  - **Use Case**: Vehicle safety messages, traffic management
  
- **cv2x**: Captures C-V2X (Cellular Vehicle-to-Everything) messages
  - **Port**: 5002 (UDP)
  - **Protocol**: 3GPP Release 14+ / C-V2X
  - **Use Case**: Cellular-based vehicle communication

### Network Security Collectors
- **syslog**: Captures syslog messages from network devices
  - **Port**: 514 (UDP)
  - **Protocol**: RFC 3164/5424 Syslog
  - **Use Case**: System logs, security events from firewalls/routers
  
- **snmp**: Captures SNMP traps from network infrastructure
  - **Port**: 162 (UDP)
  - **Protocol**: SNMPv1/v2c/v3
  - **Use Case**: Network device alerts, performance monitoring

## Testing Sequence

### Basic Workflow Test:
1. **Check Initial Status**: `GET /collectors`
2. **Start V2X Collectors**: 
   - `POST /collectors/dsrc/start`
   - `POST /collectors/cv2x/start`
3. **Verify Running**: `GET /collectors`
4. **Test Individual Stop**: `POST /collectors/dsrc/stop`
5. **Verify Status Change**: `GET /collectors`
6. **Test Start All**: `POST /collectors/start-all`
7. **Test Stop All**: `POST /collectors/stop-all`

### Error Handling Tests:
1. **Invalid Collector Name**: `POST /collectors/invalid/start`
2. **Start Already Running**: Start the same collector twice
3. **Stop Not Running**: Stop a collector that's not running
4. **Port Conflicts**: Start collectors when ports are occupied

### V2X Integration Test:
1. **Start V2X Collectors**: Start dsrc and cv2x collectors
2. **Send Test Messages**: Use V2X simulator to send UDP messages
3. **Check Security Events**: Verify events created via `GET /security-events`
4. **Check Alerts**: Verify rule evaluation via `GET /alerts`

## Real-World Usage Scenarios

### Production Deployment:
```
1. Start all collectors: POST /collectors/start-all
2. Monitor status: GET /collectors (periodic health checks)
3. Handle failures: Restart individual collectors as needed
4. Graceful shutdown: POST /collectors/stop-all
```

### Development/Testing:
```
1. Start only needed collectors: POST /collectors/dsrc/start
2. Test specific protocols
3. Stop collectors when testing complete
```

### V2X Security Monitoring:
```
1. Start V2X collectors: dsrc and cv2x
2. Monitor for vehicle security events
3. Correlate with security rules and alerts
4. Real-time threat detection
```

## Integration with Other Endpoints

### Data Flow After Collector Start:
```
Collector Running → Captures Messages → Creates Security Events → Rule Evaluation → Alerts
     ↑                     ↓                      ↓                    ↓
GET /collectors    GET /security-events    GET /rules        GET /alerts
```

### Related Monitoring Endpoints:
- **GET /security-events**: View events captured by collectors
- **GET /alerts**: View alerts generated from collector events  
- **GET /dashboard/overview**: System-wide statistics including collector metrics
- **GET /benchmark/metrics**: Performance metrics for active collectors

## Troubleshooting Common Issues

### Port Binding Failures:
- **Issue**: `address already in use`
- **Solution**: Check for existing processes on collector ports
- **Commands**: `netstat -tulpn | grep :5001`

### Collector Won't Start:
- **Issue**: Permission denied or network restrictions
- **Solution**: Verify firewall rules and user permissions
- **Check**: Application logs for detailed error messages

### High CPU/Memory Usage:
- **Issue**: Collectors consuming excessive resources
- **Solution**: Monitor via `GET /benchmark/metrics`
- **Action**: Stop problematic collectors individually

### V2X Message Processing:
- **Issue**: Messages received but no security events created
- **Solution**: Check ingestion pipeline and rule engine
- **Debug**: Verify EventIngester and RuleEngine integration