
package main

import (
	"log"
	"time"

	"app/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupDatabase initializes the GORM connection to Postgres
func setupDatabase() *gorm.DB {
	// DSN (Data Source Name) contains connection information
	// In a real project, you might want to load these values from environment variables
	dsn := "host=localhost user=go_user password=go_pass dbname=go_db port=5420 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// automatically migrate the schema, creating or updating tables as needed
	err = db.AutoMigrate(
		&models.User{},
		&models.Sensor{},
		&models.Station{},
		&models.TrafficMeasurement{},
		&models.UserEvent{},
	)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	log.Println("Database connection successful and migrations complete")
	return db
}

