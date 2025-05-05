

package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "traffic-monitoring-go/app/siem"
    "traffic-monitoring-go/app/siem/elasticsearch"
)

// DashboardHandler handles dashboard-related endpoints
type DashboardHandler struct {
    DB               *gorm.DB
    DashboardService *siem.DashboardService
    ESService        *elasticsearch.Service
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(db *gorm.DB, esService *elasticsearch.Service) *DashboardHandler {
    return &DashboardHandler{
        DB:               db,
        DashboardService: siem.NewDashboardService(db),
        ESService:        esService,
    }
}

// GetEventSummary handles GET /dashboard/events/summary
func (h *DashboardHandler) GetEventSummary(c *gin.Context) {
    timeRange := c.DefaultQuery("timeRange", "last_30_days")
    
    summary, err := h.DashboardService.GetEventSummary(timeRange)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, summary)
}

// GetAlertSummary handles GET /dashboard/alerts/summary
func (h *DashboardHandler) GetAlertSummary(c *gin.Context) {
    timeRange := c.DefaultQuery("timeRange", "last_30_days")
    
    summary, err := h.DashboardService.GetAlertSummary(timeRange)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, summary)
}

// GetEventTimeSeries handles GET /dashboard/events/timeseries
func (h *DashboardHandler) GetEventTimeSeries(c *gin.Context) {
    timeRange := c.DefaultQuery("timeRange", "last_30_days")
    groupBy := c.DefaultQuery("groupBy", "day")
    
    data, err := h.DashboardService.GetEventTimeSeries(timeRange, groupBy)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, data)
}

// GetTopSourceIPs handles GET /dashboard/events/top-sources
func (h *DashboardHandler) GetTopSourceIPs(c *gin.Context) {
    timeRange := c.DefaultQuery("timeRange", "last_30_days")
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    
    data, err := h.DashboardService.GetTopSourceIPs(timeRange, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, data)
}

// GetTopTriggeredRules handles GET /dashboard/alerts/top-rules
func (h *DashboardHandler) GetTopTriggeredRules(c *gin.Context) {
    timeRange := c.DefaultQuery("timeRange", "last_30_days")
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    
    data, err := h.DashboardService.GetTopTriggeredRules(timeRange, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, data)
}

// GetDashboardOverview handles GET /dashboard/overview
func (h *DashboardHandler) GetDashboardOverview(c *gin.Context) {
    timeRange := c.DefaultQuery("timeRange", "last_30_days")
    
    // Get event summary
    eventSummary, err := h.DashboardService.GetEventSummary(timeRange)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get event summary: " + err.Error()})
        return
    }
    
    // Get alert summary
    alertSummary, err := h.DashboardService.GetAlertSummary(timeRange)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get alert summary: " + err.Error()})
        return
    }
    
    // Get event time series
    eventTimeSeries, err := h.DashboardService.GetEventTimeSeries(timeRange, "day")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get event time series: " + err.Error()})
        return
    }
    
    // Get top source IPs
    topSources, err := h.DashboardService.GetTopSourceIPs(timeRange, 5)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top sources: " + err.Error()})
        return
    }
    
    // Get top triggered rules
    topRules, err := h.DashboardService.GetTopTriggeredRules(timeRange, 5)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get top rules: " + err.Error()})
        return
    }
    
    // Combine all data into one response
    c.JSON(http.StatusOK, gin.H{
        "event_summary":     eventSummary,
        "alert_summary":     alertSummary,
        "event_time_series": eventTimeSeries,
        "top_sources":       topSources,
        "top_rules":         topRules,
    })
}


// GetElasticsearchDashboard handles GET /dashboard/es/overview
func (h *DashboardHandler) GetElasticsearchDashboard(c *gin.Context) {
    // Check if Elasticsearch is available
    if h.ESService == nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Elasticsearch service not available"})
        return
    }
    
    timeRange := c.DefaultQuery("timeRange", "last_30_days")
    
    // Get dashboard stats from Elasticsearch
    stats, err := h.ESService.GetDashboardStats(timeRange)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard stats from Elasticsearch: " + err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, stats)
}
