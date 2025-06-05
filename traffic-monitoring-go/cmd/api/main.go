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
	"traffic-monitoring-go/internal/pkg/auth"
	"traffic-monitoring-go/internal/pkg/config"
	"traffic-monitoring-go/internal/repository"
	"traffic-monitoring-go/internal/service"
)

func main() {
	cfg := config.Load()

	// initialize logger
	logger := setupLogger(cfg.Server.LogLevel)
	logger.Info("Starting V2X SIEM API server")

	// connect to the database
	db, err := setupDatabase(cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	jwtManager := auth.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.TokenDuration)

	// initialize repositories
	ruleRepo := repository.NewGormRuleRepository(db)
	alertRepo := repository.NewGormAlertRepository(db)
	securityEventRepo := repository.NewGormSecurityEventRepository(db)
	userRepo := repository.NewGormUserRepository(db)

	// initialize services
	ruleService := service.NewRuleService(ruleRepo)
	alertService := service.NewAlertService(alertRepo, ruleRepo)
	securityEventService := service.NewSecurityEventService(securityEventRepo, alertRepo, ruleRepo)
	authService := service.NewAuthService(userRepo, jwtManager)
	userService := service.NewUserService(userRepo)

	// initialize handlers
	ruleHandler := handlers.NewRuleHandler(ruleService)
	alertHandler := handlers.NewAlertHandler(alertService)
	securityEventHandler := handlers.NewSecurityEventHandler(securityEventService)
	authHandler := handlers.NewAuthHandler(authService, userService)
	userHandler := handlers.NewUserHandler(userService)

	// setup router
	router := api.NewRouter(
		logger,
		jwtManager,
		ruleHandler,
		alertHandler,
		securityEventHandler,
		authHandler,
		userHandler,
	)
	router.Setup()

	// start the server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router.Engine(),
	}

	// graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.Info("Server started on :8080")

	// wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exiting")
}

// setupLogger initializes and configures the logger
func setupLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.Warnf("Invalid log level %s, defaulting to info", logLevel)
		level = logrus.InfoLevel
	}

	logger.SetLevel(level)
	return logger
}

// setupdatabase initializes the database connection
func setupDatabase(dsn string) (*gorm.DB, error) {

	// retry connection a few times
	var db *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
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
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection successful")
	return db, nil
}
