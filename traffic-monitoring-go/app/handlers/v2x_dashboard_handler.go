package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/siem"
)

// V2XDashboardHandler handles V2X dashboard-related endpoints
type V2XDashboardHandler struct {
	DB               *gorm.DB
	V2XDashboardService *siem.V2XDashboardService
}

// NewV2XDashboardHandler creates a new V2XDashboardHandler
func NewV2XDashboardHandler(db *gorm.DB) *V2XDashboardHandler {
	return &V2XDashboardHandler{
		DB:               db,
		V2XDashboardService: siem.NewV2XDashboardService(db),
	}
}

// GetV2XSummary handles GET /v2x-dashboard/summary
func (h *V2XDashboardHandler) GetV2XSummary(c *gin.Context) {
	timeRange := c.DefaultQuery("timeRange", "last_hour")
	
	summary, err := h.V2XDashboardService.GetV2XSummary(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, summary)
}

// GetV2XSecuritySummary handles GET /v2x-dashboard/security-summary
func (h *V2XDashboardHandler) GetV2XSecuritySummary(c *gin.Context) {
	timeRange := c.DefaultQuery("timeRange", "last_hour")
	
	summary, err := h.V2XDashboardService.GetV2XSecuritySummary(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, summary)
}

// GetV2XAnomalySummary handles GET /v2x-dashboard/anomaly-summary
func (h *V2XDashboardHandler) GetV2XAnomalySummary(c *gin.Context) {
	timeRange := c.DefaultQuery("timeRange", "last_hour")
	
	summary, err := h.V2XDashboardService.GetV2XAnomalySummary(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, summary)
}

// GetVehicleLocations handles GET /v2x-dashboard/vehicle-locations
func (h *V2XDashboardHandler) GetVehicleLocations(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	
	locations, err := h.V2XDashboardService.GetRecentVehicleLocations(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, locations)
}

// GetActiveAlerts handles GET /v2x-dashboard/active-alerts
func (h *V2XDashboardHandler) GetActiveAlerts(c *gin.Context) {
	timeRange := c.DefaultQuery("timeRange", "last_hour")
	
	alerts, err := h.V2XDashboardService.GetActiveAlerts(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, alerts)
}

// GetV2XDashboardOverview handles GET /v2x-dashboard/overview
func (h *V2XDashboardHandler) GetV2XDashboardOverview(c *gin.Context) {
	timeRange := c.DefaultQuery("timeRange", "last_hour")
	
	// Get V2X summary
	summary, err := h.V2XDashboardService.GetV2XSummary(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get V2X summary: " + err.Error()})
		return
	}
	
	// Get security summary
	securitySummary, err := h.V2XDashboardService.GetV2XSecuritySummary(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get security summary: " + err.Error()})
		return
	}
	
	// Get anomaly summary
	anomalySummary, err := h.V2XDashboardService.GetV2XAnomalySummary(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get anomaly summary: " + err.Error()})
		return
	}
	
	// Get vehicle locations (limited)
	vehicleLocations, err := h.V2XDashboardService.GetRecentVehicleLocations(25)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get vehicle locations: " + err.Error()})
		return
	}
	
	// Get active alerts
	activeAlerts, err := h.V2XDashboardService.GetActiveAlerts(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active alerts: " + err.Error()})
		return
	}
	
	// Combine all data into one response
	c.JSON(http.StatusOK, gin.H{
		"summary":           summary,
		"security_summary":  securitySummary,
		"anomaly_summary":   anomalySummary,
		"vehicle_locations": vehicleLocations,
		"active_alerts":     activeAlerts,
	})
}