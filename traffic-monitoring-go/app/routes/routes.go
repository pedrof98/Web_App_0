package routes

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/handlers"
	"traffic-monitoring-go/app/siem/elasticsearch"
)

// RegisterRoutes sets up all the API endpoints and binds them to their handlers.
func RegisterRoutes(router *gin.Engine, db *gorm.DB, esService *elasticsearch.Service) {
	// Create handler instances.
	stationHandler := handlers.NewStationHandler(db)
	sensorHandler := handlers.NewSensorHandler(db)
	measurementHandler := handlers.NewMeasurementHandler(db)
	eventHandler := handlers.NewEventHandler(db)
	collectorHandler := handlers.NewCollectorHandler(db)


	// Create handler instances for SIEM funcitonality
	securityEventHandler := handlers.NewSecurityEventHandler(db, esService)
	alertHandler := handlers.NewAlertHandler(db, esService)
	ruleHandler := handlers.NewRuleHandler(db)
	logSourceHandler := handlers.NewLogSourceHandler(db)


	// Create ingestion handler
	ingestionHandler := handlers.NewIngestionHandler(db, esService)

	
	// create a dashboard handler
	dashboardHandler := handlers.NewDashboardHandler(db, esService)



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

	// Security event routes
	securityEventRoutes := router.Group("/security-events")
	{
		securityEventRoutes.GET("/", securityEventHandler.GetSecurityEvents)
		securityEventRoutes.POST("/", securityEventHandler.CreateSecurityEvent)
		securityEventRoutes.GET("/:id", securityEventHandler.GetSecurityEvent)
		securityEventRoutes.POST("/batch", securityEventHandler.CreateBatchSecurityEvents)
	}


	// Alert routes
	alertRoutes := router.Group("/alerts")
	{
		alertRoutes.GET("/", alertHandler.GetAlerts)
		alertRoutes.GET("/:id", alertHandler.GetAlert)
		alertRoutes.PUT("/:id", alertHandler.UpdateAlert)
		alertRoutes.POST("/:id/notify", alertHandler.SendNotification)
		alertRoutes.GET("/channels", alertHandler.GetNotificationChannels)
	}

	// Rule routes
	ruleRoutes := router.Group("/rules")
	{
		ruleRoutes.GET("/", ruleHandler.GetRules)
		ruleRoutes.POST("/", ruleHandler.CreateRule)
		ruleRoutes.GET("/:id", ruleHandler.GetRule)
		ruleRoutes.PUT("/:id", ruleHandler.UpdateRule)
		ruleRoutes.DELETE("/:id", ruleHandler.DeleteRule)
	}

	// Log source routes
	logSourceRoutes := router.Group("/log-sources")
	{
		logSourceRoutes.GET("/", logSourceHandler.GetLogSources)
		logSourceRoutes.POST("/", logSourceHandler.CreateLogSource)
		logSourceRoutes.GET("/:id", logSourceHandler.GetLogSource)
		logSourceRoutes.PUT("/:id", logSourceHandler.UpdateLogSource)
		logSourceRoutes.DELETE("/:id", logSourceHandler.DeleteLogSource)
	}



	// Ingestion routes
	ingestionRoutes := router.Group("/ingest")
	{
		ingestionRoutes.POST("/", ingestionHandler.IngestEvent)
	}


	// Collector routes
	collectorRoutes := router.Group("/collectors")
	{
		collectorRoutes.GET("/", collectorHandler.GetCollectors)
		collectorRoutes.POST("/:name/start", collectorHandler.StartCollector)
		collectorRoutes.POST("/:name/stop", collectorHandler.StopCollector)
		collectorRoutes.POST("/start-all", collectorHandler.StartAllCollectors)
		collectorRoutes.POST("/stop-all", collectorHandler.StopAllCollectors)
	}


	// Dashboard routes
	dashboardRoutes := router.Group("/dashboard")
	{
		dashboardRoutes.GET("/overview", dashboardHandler.GetDashboardOverview)
		dashboardRoutes.GET("/events/summary", dashboardHandler.GetEventSummary)
		dashboardRoutes.GET("/alerts/summary", dashboardHandler.GetAlertSummary)
		dashboardRoutes.GET("/events/timeseries", dashboardHandler.GetEventTimeSeries)
		dashboardRoutes.GET("/events/top-sources", dashboardHandler.GetTopSourceIPs)
		dashboardRoutes.GET("/alerts/top-rules", dashboardHandler.GetTopTriggeredRules)
	}


	// Health check endpoint for service discovery
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})


}
