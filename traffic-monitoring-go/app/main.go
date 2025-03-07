package main

import (
	"github.com/gin-gonic/gin"
	"traffic-monitoring-go/app/database"
	"traffic-monitoring-go/app/routes"
)

func main() {
	// Initialize the database connection.
	db := database.SetupDatabase()

	// Create a new Gin router with default middleware (logger and recovery).
	router := gin.Default()

	// Register all API routes.
	routes.RegisterRoutes(router, db)

	// Start the server on port 8080.
	router.Run(":8080")
}
