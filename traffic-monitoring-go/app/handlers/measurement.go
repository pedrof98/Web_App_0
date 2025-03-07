package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
)

// MeasurementHandler holds a reference to the database.
type MeasurementHandler struct {
	DB *gorm.DB
}

// NewMeasurementHandler creates a new MeasurementHandler.
func NewMeasurementHandler(db *gorm.DB) *MeasurementHandler {
	return &MeasurementHandler{DB: db}
}

// GetMeasurements handles GET /measurements.
func (h *MeasurementHandler) GetMeasurements(c *gin.Context) {
	var measurements []models.TrafficMeasurement
	if err := h.DB.Find(&measurements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, measurements)
}

// CreateMeasurement handles POST /measurements.
func (h *MeasurementHandler) CreateMeasurement(c *gin.Context) {
	var measurement models.TrafficMeasurement
	if err := c.ShouldBindJSON(&measurement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Create(&measurement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, measurement)
}

// GetMeasurement handles GET /measurements/:id.
func (h *MeasurementHandler) GetMeasurement(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid measurement ID"})
		return
	}
	var measurement models.TrafficMeasurement
	if err := h.DB.First(&measurement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Measurement not found"})
		return
	}
	c.JSON(http.StatusOK, measurement)
}

// CreateBatchMeasurements handles POST /measurements/batch.
func (h *MeasurementHandler) CreateBatchMeasurements(c *gin.Context) {
	var measurements []models.TrafficMeasurement
	if err := c.ShouldBindJSON(&measurements); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Create(&measurements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Batch measurements created", "count": len(measurements)})
}
