package collectors

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/siem"
)

// Collector defines the interface for all security event collectors
type Collector interface {
	// Start begins collection process
	Start(ctx context.Context) error
	// Stop ends the collection process
	Stop() error
	// Name returns the collector's name
	Name() string
}

// BaseCollector contains common fields and methods for all collectors
type BaseCollector struct {
	DB           *gorm.DB
	EventIngester *siem.EventIngester
	Running      bool
	StopChan     chan struct{}
}

// NewBaseCollector creates a new BaseCollector
func NewBaseCollector(db *gorm.DB) *BaseCollector {
	return &BaseCollector{
		DB:           db,
		EventIngester: siem.NewEventIngester(db),
		Running:      false,
		StopChan:     make(chan struct{}),
	}
}

// IsRunning returns whether the collector is running
func (c *BaseCollector) IsRunning() bool {
	return c.Running
}

// Name returns a default name for the base collector
// Note: This should be overridden by derived collectors
func (c *BaseCollector) Name() string {
	return "base-collector"
}

// Start is a base implementation that should be overridden by derived collectors
func (c *BaseCollector) Start(ctx context.Context) error {
	if c.Running {
		return errors.New("collector is already running")
	}
	// Base implementation just updates the status
	c.Running = true
	return nil
}

// Stop is a base implementation that should be overridden by derived collectors
func (c *BaseCollector) Stop() error {
	if !c.Running {
		return errors.New("collector is not running")
	}
	// Base implementation just updates the status
	c.Running = false
	return nil
}