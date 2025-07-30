# Dashboard Overview Architecture and Data Aggregation Analysis

## Complete Data Flow: From Raw Events to Dashboard Insights

```
V2X Messages → Collectors → Security Events → Rules → Alerts → Dashboard Aggregation
                                     ↓                      ↓              ↓
                              Event Summaries         Alert Summaries   Combined Overview
```

### 1. **Dashboard Service Architecture** (Multi-Service Aggregation)

```go
// Dashboard handler orchestrates multiple service calls
func (h *DashboardHandler) GetDashboardOverview(c *gin.Context) {
    timeRange := c.DefaultQuery("timeRange", "last_30_days")
    
    // Five parallel data aggregation streams
    eventSummary, err := h.DashboardService.GetEventSummary(timeRange)      // Security events analysis
    alertSummary, err := h.DashboardService.GetAlertSummary(timeRange)      // Alert status breakdown  
    eventTimeSeries, err := h.DashboardService.GetEventTimeSeries(timeRange, "day") // Temporal trends
    topSources, err := h.DashboardService.GetTopSourceIPs(timeRange, 5)     // Threat source analysis
    topRules, err := h.DashboardService.GetTopTriggeredRules(timeRange, 5)  // Detection effectiveness
    
    // Composite response with all security intelligence
    return CombinedDashboardData{...}
}
```

### 2. **Time Range Processing and SQL Generation**

```go
// Dynamic time filter generation for flexible analysis periods
func getTimeFilter(timeRange string) string {
    switch timeRange {
    case "last_hour":
        return "created_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR)"
    case "today":
        return "created_at >= DATE(NOW())"
    case "yesterday":
        return "created_at >= DATE_SUB(DATE(NOW()), INTERVAL 1 DAY) AND created_at < DATE(NOW())"
    case "last_7_days":
        return "created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)"
    case "last_30_days":
        return "created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)"
    case "this_month":
        return "created_at >= DATE_FORMAT(NOW(), '%Y-%m-01')"
    case "last_month":
        return "created_at >= DATE_SUB(DATE_FORMAT(NOW(), '%Y-%m-01'), INTERVAL 1 MONTH) AND created_at < DATE_FORMAT(NOW(), '%Y-%m-01')"
    case "this_year":
        return "created_at >= DATE_FORMAT(NOW(), '%Y-01-01')"
    default:
        return "created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)" // Default fallback
    }
}
```

### 3. **Event Summary Aggregation Deep Dive**

```go
// Multi-dimensional event analysis with severity breakdown
type EventCountSummary struct {
    Total     int64 `json:"total"`     // Overall security event volume
    Critical  int64 `json:"critical"`  // Active attacks requiring immediate response
    High      int64 `json:"high"`      // Significant threats needing prompt attention
    Medium    int64 `json:"medium"`    // Moderate security concerns
    Low       int64 `json:"low"`       // Minor issues or policy violations
    Info      int64 `json:"info"`      // Normal operational events
}

// Implementation with optimized counting queries
func (s *DashboardService) GetEventSummary(timeRange string) (*EventCountSummary, error) {
    var summary EventCountSummary
    
    // Base query with time filtering
    query := s.DB.Model(&models.SecurityEvent{})
    if timeFilter := getTimeFilter(timeRange); timeFilter != "" {
        query = query.Where(timeFilter)
    }
    
    // Efficient counting with individual severity queries
    // (Alternative: single GROUP BY query for better performance)
    query.Count(&summary.Total)
    query.Where("severity = ?", models.SeverityCritical).Count(&summary.Critical)
    query.Where("severity = ?", models.SeverityHigh).Count(&summary.High)
    query.Where("severity = ?", models.SeverityMedium).Count(&summary.Medium)
    query.Where("severity = ?", models.SeverityLow).Count(&summary.Low)
    query.Where("severity = ?", models.SeverityInfo).Count(&summary.Info)
    
    return &summary, nil
}
```

### 4. **Alert Summary with Status and Severity Analysis**

```go
// Comprehensive alert tracking for SOC workflow management
type AlertSummary struct {
    // Workflow status tracking
    Total         int64 `json:"total"`
    Open          int64 `json:"open"`           // New alerts requiring triage
    InProgress    int64 `json:"in_progress"`    // Active investigations
    Closed        int64 `json:"closed"`         // Resolved incidents
    FalsePositive int64 `json:"false_positive"` // Benign detections
    
    // Threat severity distribution
    Critical      int64 `json:"critical"`       // Immediate response required
    High          int64 `json:"high"`           // Important security issues
    Medium        int64 `json:"medium"`         // Moderate concerns
    Low           int64 `json:"low"`            // Minor issues
}

// Alert aggregation with dual-axis analysis (status + severity)
func (s *DashboardService) GetAlertSummary(timeRange string) (*AlertSummary, error) {
    // Similar pattern to event summary but targeting alerts table
    // Provides insights into:
    // 1. SOC workload (open vs closed alerts)
    // 2. Investigation efficiency (in_progress alerts)
    // 3. Detection accuracy (false_positive rate)
    // 4. Threat landscape (severity distribution)
}
```

### 5. **Time Series Analysis for Trend Detection**

```go
// Temporal pattern analysis for threat trend identification
type TimeSeriesData struct {
    Labels []string `json:"labels"` // Time period identifiers (dates/hours)
    Data   []int64  `json:"data"`   // Event counts for each period
}

// Advanced time-based aggregation with flexible grouping
func (s *DashboardService) GetEventTimeSeries(timeRange string, groupBy string) (*TimeSeriesData, error) {
    // Dynamic SQL date formatting based on analysis granularity
    var timeFormat string
    switch groupBy {
    case "hour":   // Real-time monitoring (last 24 hours)
        timeFormat = "date_format(timestamp, '%Y-%m-%d %H:00')"
    case "day":    // Daily trends (weeks/months)
        timeFormat = "date_format(timestamp, '%Y-%m-%d')"
    case "week":   // Weekly patterns (monthly/quarterly analysis)
        timeFormat = "date_format(date_sub(timestamp, interval weekday(timestamp) day), '%Y-%m-%d')"
    case "month":  // Monthly trends (yearly analysis)
        timeFormat = "date_format(timestamp, '%Y-%m')"
    }
    
    // Aggregation query with temporal grouping
    var result []struct {
        TimeGroup string
        Count     int64
    }
    
    s.DB.Model(&models.SecurityEvent{}).
        Select(timeFormat + " as time_group, count(*) as count").
        Where(getTimeFilter(timeRange)).
        Group("time_group").
        Order("time_group").
        Find(&result)
    
    // Convert to chart-ready format
    data := &TimeSeriesData{
        Labels: make([]string, len(result)),
        Data:   make([]int64, len(result)),
    }
    
    for i, r := range result {
        data.Labels[i] = r.TimeGroup    // X-axis labels
        data.Data[i] = r.Count          // Y-axis values
    }
    
    return data, nil
}
```

### 6. **Top Source IP Analysis for Threat Hunting**

```go
// Source-based threat analysis for attack attribution
func (s *DashboardService) GetTopSourceIPs(timeRange string, limit int) ([]map[string]interface{}, error) {
    var result []struct {
        SourceIP string
        Count    int64
    }
    
    // Aggregation query with ranking
    s.DB.Model(&models.SecurityEvent{}).
        Select("source_ip, count(*) as count").
        Where("source_ip IS NOT NULL AND source_ip != ''").  // Filter valid IPs
        Where(getTimeFilter(timeRange)).
        Group("source_ip").
        Order("count DESC").                                // Rank by activity volume
        Limit(limit).                                       // Top N results
        Find(&result)
    
    // Transform to generic map format for JSON response
    sources := make([]map[string]interface{}, len(result))
    for i, r := range result {
        sources[i] = map[string]interface{}{
            "source_ip": r.SourceIP,
            "count":     r.Count,
        }
    }
    
    return sources, nil
}
```

### 7. **Top Triggered Rules Analysis for Detection Effectiveness**

```go
// Rule effectiveness analysis for detection coverage assessment
func (s *DashboardService) GetTopTriggeredRules(timeRange string, limit int) ([]map[string]interface{}, error) {
    var result []struct {
        RuleName string
        Count    int64
    }
    
    // Complex JOIN query for rule-alert correlation
    s.DB.Table("alerts").
        Select("rules.name as rule_name, count(*) as count").
        Joins("JOIN rules ON alerts.rule_id = rules.id").   // Get rule metadata
        Where(getTimeFilter(timeRange)).
        Group("rules.name").
        Order("count DESC").                                // Most triggered first
        Limit(limit).
        Find(&result)
    
    // V2X-specific rules typically dominate due to high message volume
    // Expected top rules:
    // 1. "V2X Position Jump Detection" (GPS spoofing attempts)
    // 2. "V2X Invalid Digital Signature" (authentication failures)
    // 3. "V2X Message Flooding Attack" (DoS attempts)
    // 4. "V2X Speed Anomaly Detection" (physics violations)
    
    return transformToMapArray(result), nil
}
```

### 8. **V2X-Specific Dashboard Patterns**

```go
// V2X security event distribution patterns
type V2XDashboardPatterns struct {
    // Expected event severity distribution for V2X environments
    NormalTrafficPattern EventCountSummary{
        Total:    100000,  // High volume due to continuous V2X communication
        Info:     85000,   // 85% - Normal BSM/DENM messages
        Low:      10000,   // 10% - Minor protocol violations, timing issues
        Medium:   4000,    // 4% - Speed anomalies, moderate concerns
        High:     900,     // 0.9% - Position jumps, potential GPS spoofing
        Critical: 100,     // 0.1% - Invalid signatures, active attacks
    }
    
    // V2X time series patterns
    TemporalPatterns struct {
        PeakHours    []string // 07:00-09:00, 17:00-19:00 (commute times)
        WeekdayVsWeekend float64 // 3:1 ratio typical
        AttackPatterns []string // Concentrated bursts during specific periods
    }
    
    // V2X source analysis
    SourcePatterns struct {
        VehicleSimulators []string // 192.168.1.x range
        DSRCCollectors   []string // V2X infrastructure IPs
        CV2XCollectors   []string // Cellular V2X endpoints
    }
}
```

### 9. **Performance Optimization Strategies**

```go
// Dashboard query optimization techniques
type DashboardOptimization struct {
    // Index strategy for fast aggregations
    RequiredIndexes []string{
        "CREATE INDEX idx_events_created_severity ON security_events(created_at, severity)",
        "CREATE INDEX idx_alerts_created_status ON alerts(created_at, status)",
        "CREATE INDEX idx_events_source_ip ON security_events(source_ip)",
        "CREATE INDEX idx_alerts_rule_created ON alerts(rule_id, created_at)",
    }
    
    // Query optimization patterns
    Optimizations struct {
        // Single GROUP BY query instead of multiple COUNT queries
        SingleAggregationQuery string
        
        // Materialized views for frequently accessed aggregations
        MaterializedViews []string
        
        // Caching strategy for dashboard data
        CacheStrategy struct {
            TTL           time.Duration // 5 minutes for real-time dashboards
            InvalidateOn  []string      // New events, alert status changes
            CacheKeys     []string      // Per time range, per user access level
        }
    }
}

// Optimized event summary with single query
func (s *DashboardService) GetEventSummaryOptimized(timeRange string) (*EventCountSummary, error) {
    var results []struct {
        Severity string
        Count    int64
    }
    
    // Single GROUP BY query instead of 6 separate queries
    s.DB.Model(&models.SecurityEvent{}).
        Select("severity, count(*) as count").
        Where(getTimeFilter(timeRange)).
        Group("severity").
        Find(&results)
    
    // Map results to summary structure
    summary := &EventCountSummary{}
    for _, r := range results {
        summary.Total += r.Count
        switch r.Severity {
        case "critical": summary.Critical = r.Count
        case "high":     summary.High = r.Count
        case "medium":   summary.Medium = r.Count
        case "low":      summary.Low = r.Count
        case "info":     summary.Info = r.Count
        }
    }
    
    return summary, nil
}
```

### 10. **Dashboard Data Pipeline Architecture**

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   V2X Vehicle   │───▶│   Collectors     │───▶│ Security Events │
│   Messages      │    │  (DSRC/C-V2X)    │    │   Database      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Dashboard     │◀───│  Rule Engine     │◀───│   Event         │
│   Overview      │    │  Evaluation      │    │   Ingestion     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
        ▲                        │                        
        │                        ▼                        
┌─────────────────┐    ┌──────────────────┐              
│   Aggregation   │    │     Alerts       │              
│   Services      │    │    Database      │              
└─────────────────┘    └──────────────────┘              
```

### 11. **Real-Time Dashboard Updates**

```go
// WebSocket integration for real-time dashboard updates
type DashboardWebSocket struct {
    // Real-time event streaming
    EventStream chan models.SecurityEvent
    
    // Alert notifications
    AlertStream chan models.Alert
    
    // Dashboard metrics updates
    MetricsUpdate chan DashboardMetrics
}

// Dashboard update trigger patterns
func (d *DashboardService) TriggerUpdate(eventType string) {
    switch eventType {
    case "new_security_event":
        // Update event summary counters
        d.InvalidateEventSummaryCache()
        
    case "new_alert":
        // Update alert summary and top rules
        d.InvalidateAlertSummaryCache()
        d.InvalidateTopRulesCache()
        
    case "alert_status_change":
        // Update alert status distribution
        d.InvalidateAlertSummaryCache()
    }
}
```

### 12. **Dashboard Analytics and Insights**

```go
// Advanced analytics derived from dashboard data
type DashboardAnalytics struct {
    // Security posture metrics
    SecurityPosture struct {
        ThreatLevel      string  // Based on critical/high event ratio
        DetectionRate    float64 // Alerts per security event
        ResponseTime     float64 // Average time to alert resolution
        FalsePositiveRate float64 // False positives / total alerts
    }
    
    // V2X-specific metrics
    V2XMetrics struct {
        VehicleCommunicationHealth float64 // Info events / total events
        AttackFrequency           float64 // Critical events per hour
        GPSSpoofingAttempts       int64   // Position jump alerts
        AuthenticationFailures    int64   // Invalid signature events
    }
    
    // Operational efficiency
    SOCEfficiency struct {
        AlertBacklog        int64   // Open alerts count
        InvestigationLoad   int64   // In-progress alerts
        ClosureRate         float64 // Closed alerts / total alerts
        AnalystProductivity float64 // Alerts per analyst per day
    }
}
```

This dashboard architecture provides **comprehensive security intelligence aggregation** for V2X SIEM operations, enabling security teams to monitor threat landscapes, track detection effectiveness, and manage incident response workflows in real-time.