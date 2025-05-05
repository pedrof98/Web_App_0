package main

import (
	"log"
	"github.com/gin-gonic/gin"
	"traffic-monitoring-go/app/database"
	"traffic-monitoring-go/app/routes"
	"traffic-monitoring-go/app/siem/elasticsearch"
)

func main() {
	// Initialize the database connection.
	db := database.SetupDatabase()

	// create default rules
	if err := database.CreateDefaultRules(db); err != nil {
		log.Printf("Warning: failed to create default rules: %v", err)
	}

	// initialize Elasticsearch service
	esService := elasticsearch.NewService()
	if err := esService.Initialize(); err != nil {
		log.Printf("Warning: Failed to initialize Elasticsearch: %v", err)
		log.Println("The application will continue without Elasticsearch integration\nBut try to fix this issue checking the codebase")
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
