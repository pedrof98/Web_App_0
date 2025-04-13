package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"traffic-monitoring-go/app/database"
	"traffic-monitoring-go/app/routes"
	"traffic-monitoring-go/app/siem/elasticsearch"
	"traffic-monitoring-go/app/siem/collectors"
	"context"
)

func main() {
	// Initialize the database connection.
	db := database.SetupDatabase()

	// create default rules
	if err := database.CreateDefaultRules(db); err != nil {
		log.Printf("Warning: failed to create default rules: %v", err)
	}

	// initialize Elasticsearch service
	esService := elasticsearch.NewService(db)
	if err := esService.Initialize(); err != nil {
		log.Printf("Warning: Failed to initialize Elasticsearch: %v", err)
		log.Println("The application will continue without Elasticsearch integration\nBut try to fix this issue checking the codebase")
	}

	// Initialize the collector manager
	collectorManager := collectors.NewCollectorManager(db)
	
	// Register collectors with default ports
	dsrcCollector := collectors.NewEnhancedDSRCCollector(db, 5001, esService)
	cv2xCollector := collectors.NewEnhancedCV2XCollector(db, 5002, esService)
	
	collectorManager.RegisterCollector(dsrcCollector)
	collectorManager.RegisterCollector(cv2xCollector)
	
	// Start all collectors
	ctx := context.Background()
	if err := dsrcCollector.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start DSRC collector: %v", err)
	} else {
		log.Println("DSRC collector started successfully on port 5001")
	}
	
	if err := cv2xCollector.Start(ctx); err != nil {
		log.Printf("Warning: Failed to start CV2X collector: %v", err)
	} else {
		log.Println("CV2X collector started successfully on port 5002")
	}

	// Create a new Gin router with default middleware (logger and recovery).
	router := gin.Default()

	// Register all API routes.
	routes.RegisterRoutes(router, db, esService)

	// Start the server on port 8080.
	log.Println("Starting SIEM server on port 8080...")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}