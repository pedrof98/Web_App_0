package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	SIEMAPIURL       string
	EventsPerMinute  int
	EnableAttackSim  bool
	AttackFrequency  int
	IncludeV2XEvents bool
}

func Load() *Config {
	cfg := &Config{}

	// Get SIEM API URL
	cfg.SIEMAPIURL = os.Getenv("SIEM_API_URL")
	if cfg.SIEMAPIURL == "" {
		cfg.SIEMAPIURL = "http://localhost:8080"
	}
	cfg.SIEMAPIURL = strings.TrimSuffix(cfg.SIEMAPIURL, "/")

	// Get events per minute
	eventsPerMinuteStr := os.Getenv("EVENTS_PER_MINUTE")
	if eventsPerMinuteStr == "" {
		cfg.EventsPerMinute = 60
	} else {
		fmt.Sscanf(eventsPerMinuteStr, "%d", &cfg.EventsPerMinute)
		if cfg.EventsPerMinute < 1 {
			cfg.EventsPerMinute = 1
		}
	}

	// Get attack simulation setting
	enableAttackSimStr := os.Getenv("ENABLE_ATTACK_SIMULATION")
	cfg.EnableAttackSim = strings.ToLower(enableAttackSimStr) == "true"

	// Get attack frequency
	attackFrequencyStr := os.Getenv("ATTACK_FREQUENCY")
	if attackFrequencyStr == "" {
		cfg.AttackFrequency = 1
	} else {
		fmt.Sscanf(attackFrequencyStr, "%d", &cfg.AttackFrequency)
		if cfg.AttackFrequency < 1 {
			cfg.AttackFrequency = 1
		}
	}

	// Get V2X events setting
	includeV2XEventsStr := os.Getenv("INCLUDE_V2X_EVENTS")
	cfg.IncludeV2XEvents = strings.ToLower(includeV2XEventsStr) == "true"

	return cfg
}
