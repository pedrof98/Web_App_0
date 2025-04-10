package elasticsearch

import (
	"fmt"
	"log"
	"sync"
	"time"

	"traffic-monitoring-go/app/models"
)

// Service is a service for interacting with Elasticsearch
type Service struct {
	Client      *ESClient
	initialized bool
	mutex       sync.RWMutex
}

// NewService creates a new Elasticsearch Service
func NewService() *Service {
	return &Service{
		Client:      NewESClient(),
		initialized: false,
	}
}

// Initialize initializes the Elasticsearch service
func (s *Service) Initialize() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.initialized {
		return nil
	}

	// Test connection to Elasticsearch
	const maxRetries = 5
	for i := 0; i < maxRetries; i++ {
		err := s.Client.CheckConnection()
		if err == nil {
			break
		}

		if i == maxRetries-1 {
			return fmt.Errorf("failed to connect to Elasticsearch after %d retries: %v", maxRetries, err)
		}

		log.Printf("Failed to connect to Elasticsearch, retrying in 10 seconds... (%d/%d)", i+1, maxRetries)
		time.Sleep(10 * time.Second)
	}

	// Ensure indices exist
	if err := s.Client.EnsureIndices(); err != nil {
		return fmt.Errorf("failed to ensure indices: %v", err)
	}

	s.initialized = true
	log.Println("Elasticsearch service initialized successfully")
	return nil
}

// IndexSecurityEvent indexes a security event in Elasticsearch
func (s *Service) IndexSecurityEvent(event *models.SecurityEvent) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return fmt.Errorf("elasticsearch service not initialized")
	}

	return s.Client.IndexSecurityEvent(event)
}

// IndexAlert indexes an alert in Elasticsearch
func (s *Service) IndexAlert(alert *models.Alert) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return fmt.Errorf("elasticsearch service not initialized")
	}

	return s.Client.IndexAlert(alert)
}

// SearchSecurityEvents searches for security events in Elasticsearch
func (s *Service) SearchSecurityEvents(query map[string]interface{}, page, pageSize int) ([]map[string]interface{}, int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return nil, 0, fmt.Errorf("elasticsearch service not initialized")
	}

	from := (page - 1) * pageSize
	return s.Client.SearchSecurityEvents(query, from, pageSize)
}

// GetDashboardStats gets dashboard statistics from Elasticsearch
func (s *Service) GetDashboardStats(timeRange string) (map[string]interface{}, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return nil, fmt.Errorf("elasticsearch service not initialized")
	}

	return s.Client.GetEventDashboardStats(timeRange)
}
