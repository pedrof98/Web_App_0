package database

import (
	"log"
	"time"
	"os"

	"traffic-monitoring-go/app/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupDatabase() *gorm.DB {
	dsn := os.Getenv("DSN")

	if dsn == "" {
		dsn = "host=db-go user=go_user password=go_pass dbname=go_db port=5432 sslmode=disable TimeZone=UTC"
	}
	
	var db *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err == nil {
			break
		}
		log.Printf("Database connection failed on attempt %d: %v. Retrying in 2 seconds...", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(
        &models.User{},
        &models.Station{},
        &models.Sensor{},
        &models.TrafficMeasurement{},
        &models.UserEvent{},
		&models.LogSource{},
		&models.SecurityEvent{},
		&models.Rule{},
		&models.Alert{},
    )
    if err != nil {
        log.Fatalf("failed to migrate models: %v", err)
    }

	// Verify database connection by executing simple query
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("Failed to ping the DB: %v", err)
	}
	

	log.Println("Database connection successful and migrations complete")
	return db
}
