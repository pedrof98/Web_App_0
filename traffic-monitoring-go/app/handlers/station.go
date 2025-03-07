package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
)

// StationHandler holds a reference to the database.
type StationHandler struct {
	DB *gorm.DB
}

// NewStationHandler creates a new StationHandler.
func NewStationHandler(db *gorm.DB) *StationHandler {
	return &StationHandler{DB: db}
}

// GetStations handles GET /stations.
func (h *StationHandler) GetStations(c *gin.Context) {
	var stations []models.Station
	if err := h.DB.Find(&stations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stations)
}

// CreateStation handles POST /stations.
func (h *StationHandler) CreateStation(c *gin.Context) {
	var station models.Station
	if err := c.ShouldBindJSON(&station); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Create(&station).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, station)
}

// GetStation handles GET /stations/:id.
func (h *StationHandler) GetStation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid station ID"})
		return
	}
	var station models.Station
	if err := h.DB.First(&station, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Station not found"})
		return
	}
	c.JSON(http.StatusOK, station)
}

// UpdateStation handles PUT /stations/:id.
func (h *StationHandler) UpdateStation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid station ID"})
		return
	}
	var station models.Station
	if err := h.DB.First(&station, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Station not found"})
		return
	}
	if err := c.ShouldBindJSON(&station); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.DB.Save(&station).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, station)
}

// DeleteStation handles DELETE /stations/:id.
func (h *StationHandler) DeleteStation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid station ID"})
		return
	}
	if err := h.DB.Delete(&models.Station{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Station deleted successfully"})
}

// GetStationEvents handles GET /stations/:id/events.
func (h *StationHandler) GetStationEvents(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid station ID"})
		return
	}
	var station models.Station
	if err := h.DB.Preload("Events").First(&station, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Station not found"})
		return
	}
	c.JSON(http.StatusOK, station.Events)
}
