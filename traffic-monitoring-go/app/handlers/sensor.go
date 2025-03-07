package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
)

// SensorHandler holds a reference to the database.
type SensorHandler struct {
	DB *gorm.DB
}

// NewSensorHandler creates a new SensorHandler.
func NewSensorHandler(db *gorm.DB) *SensorHandler {
	return &SensorHandler{DB: db}
}

// GetSensors handles GET /sensors.
func (h *SensorHandler) GetSensors(c *gin.Context) {
	var sensors []models.Sensor
	if err := h.DB.Find(&sensors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sensors)
}

// CreateSensor handles POST /sensors.
func (h *SensorHandler) CreateSensor(c *gin.Context) {
	var sensor models.Sensor
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Create(&sensor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sensor)
}

// GetSensor handles GET /sensors/:id.
func (h *SensorHandler) GetSensor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}
	var sensor models.Sensor
	if err := h.DB.First(&sensor, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}
	c.JSON(http.StatusOK, sensor)
}

// UpdateSensor handles PUT /sensors/:id.
func (h *SensorHandler) UpdateSensor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}
	var sensor models.Sensor
	if err := h.DB.First(&sensor, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Save(&sensor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sensor)
}

// DeleteSensor handles DELETE /sensors/:id.
func (h *SensorHandler) DeleteSensor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}
	if err := h.DB.Delete(&models.Sensor{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Sensor deleted successfully"})
}
