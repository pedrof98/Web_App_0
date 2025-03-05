
package main

import (
	"net/http"
	"strconv"

	"app/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var db *gorm.DB // global database instance

func main() {
	// initialize the database connection
	db = setupDatabase()

	// create a new Gin router with default middleware (logger and recovery)
	router := gin.Default()


	// Station endpoints

	stationRoutes := router.Group("/stations")
	{
		stationRoutes.GET("/", getStationsHandler) 		// list all stations
		stationRoutes.POST("/", createStationHandler)		// create new station
		stationRoutes.GET("/:id", getStationHandler)		// get a station by ID
		stationRoutes.PUT("/:id", updateStationHandler)		// update a station
		stationRoutes.DELETE("/:id", deleteStationHandler)	// delete a station
		stationRoutes.GET("/:id/events", getStationEventsHandler) // get events for a specific station
	}


	// Sensors Endpoints

	sensorRoutes := router.Group("/sensors")
	{
		sensorRoutes.GET("/", getSensorsHandler)		// list all sensors
		sensorRoutes.POST("/", createSensorHandler)		// create new sensor
		sensorRoutes.GET("/:id", getSensorHandler)		// get a sensor by ID
		sensorRoutes.PUT("/:id", updateSensorHandler)		// update a sensor
		sensorRoutes.DELETE("/:id", deleteSensorHandler)	// delete a sensor
	}


	// Measurements Endpoints

	measurementRoutes := router.Group("/measurements")
	{
		measurementRoutes.GET("/", getMeasurementsHandler)	// list all measurements
		measurementRoutes.POST("/", createMeasurementHandler)	// create new measurement
		measurementRoutes.GET("/:id", getMeasurementHandler)		// get a measurement by ID
		measurementRoutes.POST("/batch", createBatchMeasurementsHandler) // create multiple measurements at once
	}


	// User Events Endpoints

	eventRoutes := router.Group("/events")
	{
		eventRoutes.GET("/", getEventsHandler)			// list all events
		eventRoutes.POST("/", createEventHandler)		// create a new event
		eventRoutes.GET("/:id", getEventHandler)		// get an event by ID
		eventRoutes.PUT("/:id", updateEventHandler)		// update an event
		eventRoutes.DELETE("/:id", deleteEventHandler)		// delete an event
	}



	// start the server on port 8080
	router.Run(":8080")

}


//------------------------HANDLERS--------------------------------------

//Station Handlers

func getStationsHandler(c *gin.Context) {
	var stations []models.Station
	if err := db.Find(&stations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stations)
}


func createStationHandler(c *gin.Context) {
	var station models.Station
	if err := c.ShouldBindJSON(&station); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&station).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, station)
}



func getStationHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid station ID"})
		return
	}
	var station models.Station
	if err := db.First(&station, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Station not found"})
		return
	}
	c.JSON(http.StatusOK, station)
}


func updateStationHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid station ID"})
		return
	}
	var station models.Station
	if err := db.First(&station, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Station not found"})
		return
	}
	if err := c.ShouldBindJSON(&station); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, station)
}


func deleteStationHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid station ID"})
		return
	}
	if err := db.Delete(&models.Station{}, id).Error; err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Station deleted successfully"})
}

func getStationEventsHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid station ID"})
		return
	}
	var station models.Station
	// Preload the events association so that events are loaded with the station
	if err := db.Preload("Events").First(&station, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Station not found"})
		return
	}
	c.JSON(http.StatusOK, station.Events)
}



//-------------------------

// Sensor Handlers


func getSensorsHandler(c *gin.Context) {
	var sensors []models.Sensor
	if err := db.Find(&sensors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sensors)
}

// createSensorHandler handles POST /sensors.
func createSensorHandler(c *gin.Context) {
	var sensor models.Sensor
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&sensor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sensor)
}

// getSensorHandler handles GET /sensors/:id.
func getSensorHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}
	var sensor models.Sensor
	if err := db.First(&sensor, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}
	c.JSON(http.StatusOK, sensor)
}

// updateSensorHandler handles PUT /sensors/:id.
func updateSensorHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}
	var sensor models.Sensor
	if err := db.First(&sensor, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}
	if err := c.ShouldBindJSON(&sensor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Save(&sensor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sensor)
}

// deleteSensorHandler handles DELETE /sensors/:id.
func deleteSensorHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor ID"})
		return
	}
	if err := db.Delete(&models.Sensor{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Sensor deleted successfully"})
}

// -------------------------------
// Measurements Handlers
// -------------------------------

// getMeasurementsHandler handles GET /measurements.
func getMeasurementsHandler(c *gin.Context) {
	var measurements []models.TrafficMeasurement
	if err := db.Find(&measurements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, measurements)
}

// createMeasurementHandler handles POST /measurements.
func createMeasurementHandler(c *gin.Context) {
	var measurement models.TrafficMeasurement
	if err := c.ShouldBindJSON(&measurement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&measurement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, measurement)
}

// getMeasurementHandler handles GET /measurements/:id.
func getMeasurementHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid measurement ID"})
		return
	}
	var measurement models.TrafficMeasurement
	if err := db.First(&measurement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Measurement not found"})
		return
	}
	c.JSON(http.StatusOK, measurement)
}

// createBatchMeasurementsHandler handles POST /measurements/batch.
func createBatchMeasurementsHandler(c *gin.Context) {
	var measurements []models.TrafficMeasurement
	if err := c.ShouldBindJSON(&measurements); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&measurements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Batch measurements created", "count": len(measurements)})
}

// -------------------------------
// User Events Handlers
// -------------------------------

// getEventsHandler handles GET /events.
func getEventsHandler(c *gin.Context) {
	var events []models.UserEvent
	if err := db.Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

// createEventHandler handles POST /events.
func createEventHandler(c *gin.Context) {
	var event models.UserEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

// getEventHandler handles GET /events/:id.
func getEventHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}
	var event models.UserEvent
	if err := db.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

// updateEventHandler handles PUT /events/:id.
func updateEventHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}
	var event models.UserEvent
	if err := db.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

// deleteEventHandler handles DELETE /events/:id.
func deleteEventHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}
	if err := db.Delete(&models.UserEvent{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

