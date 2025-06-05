package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port        string
	Environment string
	LogLevel    string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	DSN string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey     string
	TokenDuration time.Duration
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:        getEnv("PORT", "8080"),
			Environment: getEnv("GO_ENV", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DSN", "host=db-go user=go_user password=go_pass dbname=go_db port=5432 sslmode=disable TimeZone=UTC"),
		},
		JWT: JWTConfig{
			SecretKey:     getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
			TokenDuration: getDurationEnv("JWT_DURATION", 24*time.Hour),
		},
		CORS: CORSConfig{
			AllowOrigins:     []string{"*"}, // In production, specify exact origins
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"*"},
			AllowCredentials: true,
		},
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getDurationEnv gets a duration environment variable with a fallback value
func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}

// getIntEnv gets an integer environment variable with a fallback value
func getIntEnv(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// getBoolEnv gets a boolean environment variable with a fallback value
func getBoolEnv(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}
