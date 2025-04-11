
package siem

import (
    "time"
    "gorm.io/gorm"
    "traffic-monitoring-go/app/models"
)

// DashboardService provides data for SIEM dashboards
type DashboardService struct {
    DB *gorm.DB
}

// NewDashboardService creates a new DashboardService
func NewDashboardService(db *gorm.DB) *DashboardService {
    return &DashboardService{DB: db}
}

// EventCountSummary contains event count totals by severity
type EventCountSummary struct {
    Total     int64 `json:"total"`
    Critical  int64 `json:"critical"`
    High      int64 `json:"high"`
    Medium    int64 `json:"medium"`
    Low       int64 `json:"low"`
    Info      int64 `json:"info"`
}

// AlertSummary contains alert count totals by status and severity
type AlertSummary struct {
    Total        int64 `json:"total"`
    Open         int64 `json:"open"`
    InProgress   int64 `json:"in_progress"`
    Closed       int64 `json:"closed"`
    FalsePositive int64 `json:"false_positive"`
    
    Critical     int64 `json:"critical"`
    High         int64 `json:"high"`
    Medium       int64 `json:"medium"`
    Low          int64 `json:"low"`
}

// TimeSeriesData contains time-based counts for events or alerts
type TimeSeriesData struct {
    Labels []string `json:"labels"`
    Data   []int64  `json:"data"`
}

// GetEventSummary returns summary counts of security events
func (s *DashboardService) GetEventSummary(timeRange string) (*EventCountSummary, error) {
    var summary EventCountSummary
    
    // Build query based on time range
    query := s.DB.Model(&models.SecurityEvent{})
    timeFilter := getTimeFilter(timeRange)
    if timeFilter != "" {
        query = query.Where(timeFilter)
    }
    
    // Get total count
    if err := query.Count(&summary.Total).Error; err != nil {
        return nil, err
    }
    
    // Get counts by severity
    if err := query.Where("severity = ?", models.SeverityCritical).Count(&summary.Critical).Error; err != nil {
        return nil, err
    }
    
    if err := query.Where("severity = ?", models.SeverityHigh).Count(&summary.High).Error; err != nil {
        return nil, err
    }
    
    if err := query.Where("severity = ?", models.SeverityMedium).Count(&summary.Medium).Error; err != nil {
        return nil, err
    }
    
    if err := query.Where("severity = ?", models.SeverityLow).Count(&summary.Low).Error; err != nil {
        return nil, err
    }
    
    if err := query.Where("severity = ?", models.SeverityInfo).Count(&summary.Info).Error; err != nil {
        return nil, err
    }
    
    return &summary, nil
}

// GetAlertSummary returns summary counts of alerts
func (s *DashboardService) GetAlertSummary(timeRange string) (*AlertSummary, error) {
    var summary AlertSummary
    
    // Build query based on time range
    query := s.DB.Model(&models.Alert{})
    timeFilter := getTimeFilter(timeRange)
    if timeFilter != "" {
        query = query.Where(timeFilter)
    }
    
    // Get total count
    if err := query.Count(&summary.Total).Error; err != nil {
        return nil, err
    }
    
    // Get counts by status
    if err := query.Where("status = ?", models.AlertStatusOpen).Count(&summary.Open).Error; err != nil {
        return nil, err
    }
    
    if err := query.Where("status = ?", models.AlertStatusInProgress).Count(&summary.InProgress).Error; err != nil {
        return nil, err
    }
    
    if err := query.Where("status = ?", models.AlertStatusClosed).Count(&summary.Closed).Error; err != nil {
        return nil, err
    }
    
    if err := query.Where("status = ?", models.AlertStatusFalsePositive).Count(&summary.FalsePositive).Error; err != nil {
        return nil, err
    }
    
    // Get counts by severity
    if err := query.Where("severity = ?", models.SeverityCritical).Count(&summary.Critical).Error; err != nil {
        return nil, err
    }
    
    if err := query.Where("severity = ?", models.SeverityHigh).Count(&summary.High).Error; err != nil {
        return nil, err
    }
    
    if err := query.Where("severity = ?", models.SeverityMedium).Count(&summary.Medium).Error; err != nil {
        return nil, err
    }
    
    if err := query.Where("severity = ?", models.SeverityLow).Count(&summary.Low).Error; err != nil {
        return nil, err
    }
    
    return &summary, nil
}

// GetEventTimeSeries returns time series data for security events
func (s *DashboardService) GetEventTimeSeries(timeRange string, groupBy string) (*TimeSeriesData, error) {
    // Set default grouping if not specified
    if groupBy == "" {
        groupBy = "day"
    }
    
    var result []struct {
        TimeGroup string
        Count     int64
    }
    
    // Build query based on time range
    query := s.DB.Model(&models.SecurityEvent{})
    timeFilter := getTimeFilter(timeRange)
    if timeFilter != "" {
        query = query.Where(timeFilter)
    }
    
    // Format time grouping based on group by parameter
    var timeFormat string
    switch groupBy {
    case "hour":
        timeFormat = "date_format(timestamp, '%Y-%m-%d %H:00')"
    case "day":
        timeFormat = "date_format(timestamp, '%Y-%m-%d')"
    case "week":
        timeFormat = "date_format(date_sub(timestamp, interval weekday(timestamp) day), '%Y-%m-%d')"
    case "month":
        timeFormat = "date_format(timestamp, '%Y-%m')"
    default:
        timeFormat = "date_format(timestamp, '%Y-%m-%d')"
    }
    
    // Execute the query
    if err := query.Select(timeFormat + " as time_group, count(*) as count").
        Group("time_group").
        Order("time_group").
        Find(&result).Error; err != nil {
        return nil, err
    }
    
    // Convert to time series format
    data := &TimeSeriesData{
        Labels: make([]string, len(result)),
        Data:   make([]int64, len(result)),
    }
    
    for i, r := range result {
        data.Labels[i] = r.TimeGroup
        data.Data[i] = r.Count
    }
    
    return data, nil
}

// GetTopSourceIPs returns the most common source IPs for security events
func (s *DashboardService) GetTopSourceIPs(timeRange string, limit int) ([]map[string]interface{}, error) {
    if limit <= 0 {
        limit = 10 // Default limit
    }
    
    var result []struct {
        SourceIP string
        Count    int64
    }
    
    // Build query based on time range
    query := s.DB.Model(&models.SecurityEvent{})
    timeFilter := getTimeFilter(timeRange)
    if timeFilter != "" {
        query = query.Where(timeFilter)
    }
    
    // Execute the query
    if err := query.Select("source_ip, count(*) as count").
        Where("source_ip is not null and source_ip != ''").
        Group("source_ip").
        Order("count desc").
        Limit(limit).
        Find(&result).Error; err != nil {
        return nil, err
    }
    
    // Convert to result format
    data := make([]map[string]interface{}, len(result))
    for i, r := range result {
        data[i] = map[string]interface{}{
            "source_ip": r.SourceIP,
            "count":     r.Count,
        }
    }
    
    return data, nil
}

// GetTopTriggeredRules returns the most frequently triggered rules
func (s *DashboardService) GetTopTriggeredRules(timeRange string, limit int) ([]map[string]interface{}, error) {
    if limit <= 0 {
        limit = 10 // Default limit
    }
    
    var result []struct {
        RuleID   uint
        RuleName string
        Count    int64
    }
    
    // Build query based on time range
    query := s.DB.Model(&models.Alert{}).
        Joins("JOIN rules ON alerts.rule_id = rules.id")
    
    timeFilter := getTimeFilter(timeRange)
    if timeFilter != "" {
        query = query.Where(timeFilter)
    }
    
    // Execute the query
    if err := query.Select("alerts.rule_id, rules.name as rule_name, count(*) as count").
        Group("alerts.rule_id, rules.name").
        Order("count desc").
        Limit(limit).
        Find(&result).Error; err != nil {
        return nil, err
    }
    
    // Convert to result format
    data := make([]map[string]interface{}, len(result))
    for i, r := range result {
        data[i] = map[string]interface{}{
            "rule_id":   r.RuleID,
            "rule_name": r.RuleName,
            "count":     r.Count,
        }
    }
    
    return data, nil
}

// Helper function to convert time range to SQL filter
func getTimeFilter(timeRange string) string {
    now := time.Now()
    
    switch timeRange {
    case "today":
        return "date(timestamp) = curdate()"
    case "yesterday":
        yesterday := now.AddDate(0, 0, -1)
        return "date(timestamp) = '" + yesterday.Format("2006-01-02") + "'"
    case "last_7_days":
        startDate := now.AddDate(0, 0, -7)
        return "timestamp >= '" + startDate.Format("2006-01-02") + "'"
    case "last_30_days":
        startDate := now.AddDate(0, 0, -30)
        return "timestamp >= '" + startDate.Format("2006-01-02") + "'"
    case "this_month":
        return "year(timestamp) = year(curdate()) and month(timestamp) = month(curdate())"
    case "last_month":
        lastMonth := now.AddDate(0, -1, 0)
        return "year(timestamp) = year('" + lastMonth.Format("2006-01-02") + "') and month(timestamp) = month('" + lastMonth.Format("2006-01-02") + "')"
    case "this_year":
        return "year(timestamp) = year(curdate())"
    default:
        return "" // No filter
    }
}
