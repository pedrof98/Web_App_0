package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/handlers"
)

// RegisterRoutes sets up all the API endpoints and binds them to their handlers.
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	// Create handler instances.
	stationHandler := handlers.NewStationHandler(db)
	sensorHandler := handlers.NewSensorHandler(db)
	measurementHandler := handlers.NewMeasurementHandler(db)
	eventHandler := handlers.NewEventHandler(db)

	// Station routes.
	stationRoutes := router.Group("/stations")
	{
		stationRoutes.GET("/", stationHandler.GetStations)
		stationRoutes.POST("/", stationHandler.CreateStation)
		stationRoutes.GET("/:id", stationHandler.GetStation)
		stationRoutes.PUT("/:id", stationHandler.UpdateStation)
		stationRoutes.DELETE("/:id", stationHandler.DeleteStation)
		stationRoutes.GET("/:id/events", stationHandler.GetStationEvents)
	}

	// Sensor routes.
	sensorRoutes := router.Group("/sensors")
	{
		sensorRoutes.GET("/", sensorHandler.GetSensors)
		sensorRoutes.POST("/", sensorHandler.CreateSensor)
		sensorRoutes.GET("/:id", sensorHandler.GetSensor)
		sensorRoutes.PUT("/:id", sensorHandler.UpdateSensor)
		sensorRoutes.DELETE("/:id", sensorHandler.DeleteSensor)
	}

	// Measurement routes.
	measurementRoutes := router.Group("/measurements")
	{
		measurementRoutes.GET("/", measurementHandler.GetMeasurements)
		measurementRoutes.POST("/", measurementHandler.CreateMeasurement)
		measurementRoutes.GET("/:id", measurementHandler.GetMeasurement)
		measurementRoutes.POST("/batch", measurementHandler.CreateBatchMeasurements)
	}

	// Event routes.
	eventRoutes := router.Group("/events")
	{
		eventRoutes.GET("/", eventHandler.GetEvents)
		eventRoutes.POST("/", eventHandler.CreateEvent)
		eventRoutes.GET("/:id", eventHandler.GetEvent)
		eventRoutes.PUT("/:id", eventHandler.UpdateEvent)
		eventRoutes.DELETE("/:id", eventHandler.DeleteEvent)
	}
}