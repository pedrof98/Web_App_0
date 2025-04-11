package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/siem/collectors"
)

// Collectorhandler handles collector-related endpoints
type CollectorHandler struct {
	DB			*gorm.DB
	CollectorManager	*collectors.CollectorManager
}

// NewCollectorHandler creates a new CollectorHandler and initializes collectors
func NewCollectorHandler(db *gorm.DB) *CollectorHandler {
	manager := collectors.NewCollectorManager(db)

	// Register collectors with default ports
	syslogCollector := collectors.NewSyslogCollector(db, 514) // def syslog port
	snmpCollector := collectors.NewSNMPCollector(db, 162) // def SNMP trap port

	manager.RegisterCollector(syslogCollector)
	manager.RegisterCollector(snmpCollector)

	return &CollectorHandler{
		DB:			db,
		CollectorManager:	manager,
	}
}

// GetCollectors handles GET /collectors
func (h *CollectorHandler) GetCollectors(c *gin.Context) {
	collectorNames := h.CollectorManager.GetCollectorNames()
	collectors := make([]map[string]interface{}, 0, len(collectorNames))

	for _, name := range collectorNames {
		status, err := h.CollectorManager.GetCollectorStatus(name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		collectors = append(collectors, map[string]interface{}{
			"name":		name,
			"running":	status,
		})
	}

	c.JSON(http.StatusOK, collectors)
}

// StartCollector handles PST /collectors/:name/start
func (h *CollectorHandler) StartCollector(c *gin.Context) {
	name := c.Param("name")

	err := h.CollectorManager.StartCollector(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collector started successfully"})
}


// StopCollector handles POST /collectors/:name/stop
func (h *CollectorHandler) StopCollector(c *gin.Context) {
	name := c.Param("name")

	err := h.CollectorManager.StopCollector(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collector stopped successfully"})
}

// StartAllCollectors handles POST /collectors/start-all
func (h *CollectorHandler) StartAllCollectors(c *gin.Context) {
	err := h.CollectorManager.StartAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All collectors started"})
}

// StopAllCollectors handles POST /collectors/stop-all
func (h *CollectorHandler) StopAllCollectors(c *gin.Context) {
	h.CollectorManager.StopAll()
	c.JSON(http.StatusOK, gin.H{"message": "All collectors stopped"})
}


	
