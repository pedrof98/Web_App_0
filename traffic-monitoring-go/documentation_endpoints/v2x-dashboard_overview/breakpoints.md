# Debug Breakpoints Guide for GET /v2x-dashboard/overview Endpoint

## Critical Breakpoints (Set these in order of execution)

### 1. **Dashboard Handler Entry Point**
**File**: `app/handlers/v2x_dashboard_handler.go`  
**Line**: `func (h *V2XDashboardHandler) GetV2XDashboardOverview(c *gin.Context)`  
**Purpose**: See the dashboard request and time range parameter processing

### 2. **Time Range Parameter Extraction**
**File**: `app/handlers/v2x_dashboard_handler.go`  
**Line**: `timeRange := c.DefaultQuery("timeRange", "last_hour")`  
**Purpose**: Understand how time range filtering works

### 3. **V2X Summary Service Call**
**File**: `app/handlers/v2x_dashboard_handler.go`  
**Line**: `summary, err := h.V2XDashboardService.GetV2XSummary(timeRange)`  
**Purpose**: First major database aggregation call

### 4. **V2X Summary Implementation**
**File**: `app/siem/v2x_dashboard.go`  
**Line**: `func (s *V2XDashboardService) GetV2XSummary(timeRange string) (*V2XSummary, error)`  
**Purpose**: See complex database queries for message type counting

### 5. **Time Filter Construction**
**File**: `app/siem/v2x_dashboard.go`  
**Line**: `timeFilter := getTimeWindowFilter(timeRange)`  
**Purpose**: Understand how time windows are translated to SQL WHERE clauses

### 6. **Protocol Counting Queries**
**File**: `app/siem/v2x_dashboard.go`  
**Line**: `if err := query.Where("protocol = ?", models.ProtocolDSRC).Count(&summary.DSRC).Error; err != nil {`  
**Purpose**: See GORM aggregation queries for DSRC vs C-V2X counts

### 7. **Message Type Aggregation**
**File**: `app/siem/v2x_dashboard.go`  
**Line**: `if err := query.Where("message_type = ?", models.MessageTypeBSM).Count(&summary.BSM).Error; err != nil {`  
**Purpose**: Watch BSM, SPAT, RSA, CAM, DENM, CPM message type counting

### 8. **Security Summary Service Call**
**File**: `app/handlers/v2x_dashboard_handler.go`  
**Line**: `securitySummary, err := h.V2XDashboardService.GetV2XSecuritySummary(timeRange)`  
**Purpose**: Second major aggregation for security metrics

### 9. **Security Summary Implementation**
**File**: `app/siem/v2x_dashboard.go`  
**Line**: `func (s *V2XDashboardService) GetV2XSecuritySummary(timeRange string) (*V2XSecuritySummary, error)`  
**Purpose**: See security-focused database queries (signatures, trust levels)

### 10. **Signature Validation Aggregation**
**File**: `app/siem/v2x_dashboard.go`  
**Line**: `if err := securityQuery.Where("signature_valid = ?", true).Count(&summary.ValidSignature).Error; err != nil {`  
**Purpose**: Watch signature validation statistics

### 11. **Anomaly Summary Service Call**
**File**: `app/handlers/v2x_dashboard_handler.go`  
**Line**: `anomalySummary, err := h.V2XDashboardService.GetV2XAnomalySummary(timeRange)`  
**Purpose**: Third major aggregation for anomaly detection metrics

### 12. **Vehicle Locations Service Call**
**File**: `app/handlers/v2x_dashboard_handler.go`  
**Line**: `vehicleLocations, err := h.V2XDashboardService.GetRecentVehicleLocations(25)`  
**Purpose**: Geospatial data aggregation for vehicle mapping

### 13. **Active Alerts Service Call**
**File**: `app/handlers/v2x_dashboard_handler.go`  
**Line**: `activeAlerts, err := h.V2XDashboardService.GetActiveAlerts(timeRange)`  
**Purpose**: Real-time alert aggregation

### 14. **Response Aggregation**
**File**: `app/handlers/v2x_dashboard_handler.go`  
**Line**: `c.JSON(http.StatusOK, gin.H{`  
**Purpose**: Final dashboard data consolidation

## Inspection Variables

At each breakpoint, inspect these key variables:

- `timeRange` - Time window for data filtering ("last_hour", "last_day", etc.)
- `timeFilter` - SQL WHERE clause for time filtering
- `summary` - V2X message type and protocol counts
- `securitySummary` - Signature validation and trust level metrics
- `anomalySummary` - Anomaly type and confidence metrics
- `vehicleLocations` - Recent vehicle position data
- `activeAlerts` - Current security alerts
- `query` - GORM query objects for database operations
- Database result counts for each aggregation