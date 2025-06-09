package config

import (
	"os"
	"strconv"
)

// Config holds the V2X simulator configuration
type Config struct {
	// SIEM Integration
	SIEMEndpoint string `json:"siem_endpoint"`

	// Legacy UDP support (for backward compatibility)
	DSRCPort int    `json:"dsrc_port"`
	CV2XPort int    `json:"cv2x_port"`
	Host     string `json:"host"`

	// Simulation parameters
	Interval      int  `json:"interval"`       // milliseconds between messages
	VehicleCount  int  `json:"vehicle_count"`  // number of simulated vehicles
	AttackEnabled bool `json:"attack_enabled"` // enable attack simulation
	AttackRate    int  `json:"attack_rate"`    // percentage chance of attack per message (0-100)
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	cfg := &Config{
		// Default values
		SIEMEndpoint:  "http://app:8080/ingest",
		DSRCPort:      5001,
		CV2XPort:      5002,
		Host:          "localhost",
		Interval:      200,
		VehicleCount:  10,
		AttackEnabled: true,
		AttackRate:    5, // 5% chance of attack
	}

	// Override with environment variables if present
	if endpoint := os.Getenv("SIEM_ENDPOINT"); endpoint != "" {
		cfg.SIEMEndpoint = endpoint
	}

	if host := os.Getenv("HOST"); host != "" {
		cfg.Host = host
	}

	if port := os.Getenv("DSRC_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.DSRCPort = p
		}
	}

	if port := os.Getenv("CV2X_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.CV2XPort = p
		}
	}

	if interval := os.Getenv("INTERVAL"); interval != "" {
		if i, err := strconv.Atoi(interval); err == nil {
			cfg.Interval = i
		}
	}

	if count := os.Getenv("VEHICLE_COUNT"); count != "" {
		if c, err := strconv.Atoi(count); err == nil {
			cfg.VehicleCount = c
		}
	}

	if enabled := os.Getenv("ATTACK_ENABLED"); enabled != "" {
		cfg.AttackEnabled = enabled == "true"
	}

	if rate := os.Getenv("ATTACK_RATE"); rate != "" {
		if r, err := strconv.Atoi(rate); err == nil && r >= 0 && r <= 100 {
			cfg.AttackRate = r
		}
	}

	return cfg
}
