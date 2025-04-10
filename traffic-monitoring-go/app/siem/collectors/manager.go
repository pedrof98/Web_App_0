package collectors

import (
	"context"
	"fmt"
	"log"
	"sync"

	"gorm.io/gorm"
)

// CollectorInterface extends the Collector interface with status reporting
type CollectorInterface interface {
	Collector
	IsRunning() bool
}

// Ensure that BaseCollector implements CollectorInterface
var _ CollectorInterface = (*BaseCollector)(nil)

// CollectorManager manages all security event collectors
type CollectorManager struct {
	DB          *gorm.DB
	collectors  map[string]CollectorInterface
	mutex       sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewCollectorManager creates a new CollectorManager
func NewCollectorManager(db *gorm.DB) *CollectorManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &CollectorManager{
		DB:         db,
		collectors: make(map[string]CollectorInterface),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// RegisterCollector adds a collector to the manager
func (m *CollectorManager) RegisterCollector(collector CollectorInterface) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	name := collector.Name()
	if _, exists := m.collectors[name]; exists {
		return fmt.Errorf("collector with name '%s' already registered", name)
	}

	m.collectors[name] = collector
	log.Printf("Registered collector: %s", name)
	return nil
}

// StartCollector starts a specific collector
func (m *CollectorManager) StartCollector(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	collector, exists := m.collectors[name]
	if !exists {
		return fmt.Errorf("collector '%s' not found", name)
	}

	err := collector.Start(m.ctx)
	if err != nil {
		return fmt.Errorf("failed to start collector '%s': %v", name, err)
	}

	log.Printf("Started collector: %s", name)
	return nil
}

// StopCollector stops a specific collector
func (m *CollectorManager) StopCollector(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	collector, exists := m.collectors[name]
	if !exists {
		return fmt.Errorf("collector '%s' not found", name)
	}

	err := collector.Stop()
	if err != nil {
		return fmt.Errorf("failed to stop collector '%s': %v", name, err)
	}

	log.Printf("Stopped collector: %s", name)
	return nil
}

// StartAll starts all registered collectors
func (m *CollectorManager) StartAll() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for name, collector := range m.collectors {
		err := collector.Start(m.ctx)
		if err != nil {
			log.Printf("Failed to start collector '%s': %v", name, err)
			// continue starting other collectors instead of returning early
		} else {
			log.Printf("Started collector: %s", name)
		}
	}

	return nil
}

// StopAll stops all registered collectors
func (m *CollectorManager) StopAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// cancel the context to signal all collectors to stop
	m.cancel()

	// also call stop on each collector explicitly
	for name, collector := range m.collectors {
		err := collector.Stop()
		if err != nil {
			log.Printf("Error stopping collector '%s': %v", name, err)
		} else {
			log.Printf("Stopped collector: %s", name)
		}
	}
}

// GetCollectorNames returns a list of all registered collector names
func (m *CollectorManager) GetCollectorNames() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	names := make([]string, 0, len(m.collectors))
	for name := range m.collectors {
		names = append(names, name)
	}

	return names
}

// GetCollectorStatus returns the status of a specific collector
func (m *CollectorManager) GetCollectorStatus(name string) (bool, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	collector, exists := m.collectors[name]
	if !exists {
		return false, fmt.Errorf("collector '%s' not found", name)
	}

	// Use the IsRunning method directly
	return collector.IsRunning(), nil
}