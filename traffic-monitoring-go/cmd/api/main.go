package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"traffic-monitoring-go/internal/api"
	"traffic-monitoring-go/internal/api/handlers"
	"traffic-monitoring-go/internal/repository"
	"traffic-monitoring-go/innternal/service"
)

func main() {
	// initialize logger
	log := setupLogger()
	log.Info("Starting V2X SIEM API server")

	// connect to the database
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// initialize repositories
	ruleRepo := repository.NewGormRuleRepository(db)

	// initialize services
	ruleService := service.NewRuleService(ruleRepo)

	// initialize handlers
	ruleHandler := handlers.NewRuleHandler(ruleService)

	// setup router
	router := api.NewRouter(log, ruleHandler)
	router.Setup()

	// start the server
	srv := &http.Server{
		Addr:		":8080",
		Handler:	router.Engine(),
	}


	// graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Info("Server started on :8080")

	// wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exiting")
}


// setupLogger initializes and configures the logger
func setupLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	// set log level based on environment
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.Warnf("Invalid log level %s, defaulting to info", logLevel)
		level = logrus.InfoLevel
	}

	log.SetLevel(level)
	return log
}

// setupdatabase initializes the database connection
func setupDatabase() (*gorm.DB, error) {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = "host=db-go user=go_user password=go_pass dbname=go_db port=5432 sslmode=disable TimeZone=UTC"
	}

	// retry connection a few times
	var db *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger:	logger.Default.LogMode(logger.Info),
		})

		if err == nil {
			break
		}

		log.Printf("Database connection failed on attempt %d: %v. Retrying in 2 seconds...", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after multiple attempts: %w", err)
	}

	// Verify connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	err = sqlDB.Ping()
	if err != nil {
		return nil, fmt.Errof("failed to ping database: %w", err)
	}

	// configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection successful")
	return db, nil
}

