package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
)


// LogSourceHandler handles log source-related endpoints
type LogSourceHandler struct {
	DB *gorm.DB
}


// NewLogSourceHandler creates a new LogSourceHandler
func NewLogSourceHandler(db *gorm.DB) *LogSourceHandler {
	return &LogSourceHandler{DB: db}
}


// GetLogSources handles GET /log-sources
func (h *LogSourceHandler) GetLogSources(c *gin.Context) {
	var sources []models.LogSource

	// basic filtering by type
	sourceType := c.Query("type")

	// create a query builder
	query := h.DB.Model(&models.LogSource{})

	if sourceType != "" {
		query = query.Where("type = ?", sourceType)
	}

	// By default, only show enabled sources unless specifically requested otherwise
	if c.Query("show_disabled") != "true" {
		query = query.Where("enabled = ?", true)
	}

	// Order by name ascending
	query = query.Order("name ASC")

	if err := query.Find(&sources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sources)
}

// GetLogSource handles GET /log-sources/:id
func (h *LogSourceHandler) GetLogSource(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log source ID"})
		return
	}

	var source models.LogSource
	if err := h.DB.First(&source, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log source not found"})
		return
	}

	c.JSON(http.StatusOK, source)
}



// CreateLogSource handles POST /log-sources
func (h *LogSourceHandler) CreateLogSource(c *gin.Context) {
	var source models.LogSource
	if err := c.ShouldBindJSON(&source); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Default to enabled if not specified
	if !source.Enabled {
		source.Enabled = true
	}

	if err := h.DB.Create(&source).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return

	}

	c.JSON(http.StatusCreated, source)
}

// UpdateLogSource handles PUT /log-sources/:id
func (h *LogSourceHandler) UpdateLogSource(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log source ID"})
		return
	}

	var source models.LogSource
	if err := h.DB.First(&source, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Log source not found"})
		return
	}

	if err := c.ShouldBindJSON(&source); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.DB.Save(&source).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, source)
}


// DeleteLogSource handles DELETE /log-sources/:id
func (h *LogSourceHandler) DeleteLogSource(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log source ID"})
		return
	}

	//Check if any security events reference this log source before deletion
	var eventCount int64
	if err := h.DB.Model(&models.SecurityEvent{}).Where("log_source_id = ?", id).Count(&eventCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if eventCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete log source with existing events",
			"event_count": eventCount,
		})
		return
	}

	if err := h.DB.Delete(&models.LogSource{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Log source deleted successfully"})
}
















