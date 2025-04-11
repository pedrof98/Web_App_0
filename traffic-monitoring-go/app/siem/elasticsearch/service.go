package elasticsearch

import (
	"fmt"
	"log"
	"sync"
	"time"
	"io"
	"encoding/json"
	"net/http"
	"bytes"



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

	// Create index templates for events and alerts
    if err := s.createIndexTemplates(); err != nil {
        return fmt.Errorf("failed to create index templates: %v", err)
    }

	s.initialized = true
	log.Println("Elasticsearch service initialized successfully")
	return nil
}


// createIndexTemplates creates index templates for security events and alerts
func (s *Service) createIndexTemplates() error {
    // Create template for security events
    eventsTemplate := map[string]interface{}{
        "index_patterns": []string{"security-events-*"},
        "template": map[string]interface{}{
            "settings": map[string]interface{}{
                "number_of_shards": 1,
                "number_of_replicas": 0,
            },
            "mappings": map[string]interface{}{
                "properties": map[string]interface{}{
                    "id": map[string]interface{}{
                        "type": "integer",
                    },
                    "timestamp": map[string]interface{}{
                        "type": "date",
                    },
                    "source_ip": map[string]interface{}{
                        "type": "ip",
                        "ignore_malformed": true,
                    },
                    "destination_ip": map[string]interface{}{
                        "type": "ip",
                        "ignore_malformed": true,
                    },
                    "source_port": map[string]interface{}{
                        "type": "integer",
                    },
                    "destination_port": map[string]interface{}{
                        "type": "integer",
                    },
                    "protocol": map[string]interface{}{
                        "type": "keyword",
                    },
                    "action": map[string]interface{}{
                        "type": "keyword",
                    },
                    "status": map[string]interface{}{
                        "type": "keyword",
                    },
                    "severity": map[string]interface{}{
                        "type": "keyword",
                    },
                    "category": map[string]interface{}{
                        "type": "keyword",
                    },
                    "message": map[string]interface{}{
                        "type": "text",
                    },
                    "device_id": map[string]interface{}{
                        "type": "keyword",
                    },
                    "log_source_id": map[string]interface{}{
                        "type": "integer",
                    },
                    "created_at": map[string]interface{}{
                        "type": "date",
                    },
                },
            },
        },
    }

    // Create template for alerts
    alertsTemplate := map[string]interface{}{
        "index_patterns": []string{"security-alerts-*"},
        "template": map[string]interface{}{
            "settings": map[string]interface{}{
                "number_of_shards": 1,
                "number_of_replicas": 0,
            },
            "mappings": map[string]interface{}{
                "properties": map[string]interface{}{
                    "id": map[string]interface{}{
                        "type": "integer",
                    },
                    "rule_id": map[string]interface{}{
                        "type": "integer",
                    },
                    "security_event_id": map[string]interface{}{
                        "type": "integer",
                    },
                    "timestamp": map[string]interface{}{
                        "type": "date",
                    },
                    "severity": map[string]interface{}{
                        "type": "keyword",
                    },
                    "status": map[string]interface{}{
                        "type": "keyword",
                    },
                    "assigned_to": map[string]interface{}{
                        "type": "integer",
                    },
                    "resolution": map[string]interface{}{
                        "type": "text",
                    },
                    "created_at": map[string]interface{}{
                        "type": "date",
                    },
                    "updated_at": map[string]interface{}{
                        "type": "date",
                    },
                },
            },
        },
    }

    // Put the templates to Elasticsearch
    eventsJSON, err := json.Marshal(eventsTemplate)
    if err != nil {
        return err
    }

    alertsJSON, err := json.Marshal(alertsTemplate)
    if err != nil {
        return err
    }

    // Create events template
    req, err := http.NewRequest("PUT", fmt.Sprintf("%s/_index_template/security-events-template", s.Client.URL), bytes.NewBuffer(eventsJSON))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.Client.HTTPClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("failed to create events template: %s", string(body))
    }

    // Create alerts template
    req, err = http.NewRequest("PUT", fmt.Sprintf("%s/_index_template/security-alerts-template", s.Client.URL), bytes.NewBuffer(alertsJSON))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err = s.Client.HTTPClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("failed to create alerts template: %s", string(body))
    }

    return nil
}


// IndexSecurityEvent indexes a security event in Elasticsearch
func (s *Service) IndexSecurityEvent(event *models.SecurityEvent) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return fmt.Errorf("elasticsearch service not initialized")
	}

	// create a time-based index name in the format "security-events-YYYY.MM.DD"
	indexDate := event.Timestamp.Format("2020.01.02")
	indexName := fmt.Sprintf("security-events-%s", indexDate)

	// ensure index exists
	if err := s.Client.createIndexIfNotExists(indexName); err != nil {
		return fmt.Errorf("failed to create index: %v", err)
	}

	// create a copy of the event with proper handling of empty fields
	eventMap := map[string]interface{}{
		"id":			event.ID,
		"timestamp":		event.Timestamp,
		"log_source_id":	event.LogSourceID,
		"severity":		event.Severity,
		"category":		event.Category,
		"message":		event.Message,
		"created_at":		event.CreatedAt,
	}

	// only add non-empty string fields
	if event.SourceIP != "" {
		eventMap["source_ip"] = event.SourceIP
	}
	if event.DestinationIP != "" {
		eventMap["destination_ip"] = event.DestinationIP
	}
	if event.Protocol != "" {
		eventMap["protocol"] = event.Protocol
	}
	if event.Action != "" {
		eventMap["action"] = event.Action
	}
	if event.Status != "" {
		eventMap["status"] = event.Status
	}
	if event.DeviceID != "" {
		eventMap["device_id"] = event.DeviceID
	}

	
	// only add non-nil pointer fields
	if event.SourcePort != nil {
		eventMap["source_port"] = *event.SourcePort
	}
	if event.DestinationPort != nil {
		eventMap["destination_port"] = *event.DestinationPort
	}
	if event.UserID != nil {
		eventMap["user_id"] = *event.UserID
	}

	// convert to JSON
	eventJSON, err := json.Marshal(eventMap)
	if err != nil {
		return err
	}

	// index document
	url := fmt.Sprintf("%s/%s/_doc/%d", s.Client.URL, indexName, event.ID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(eventJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to index security event: %s", string(body))
	}

	return nil


}

// IndexAlert indexes an alert in Elasticsearch
func (s *Service) IndexAlert(alert *models.Alert) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return fmt.Errorf("elasticsearch service not initialized")
	}

	
	// Create a time-based index name in the format "security-alerts-YYYY.MM.DD"
    indexDate := alert.Timestamp.Format("2006.01.02")
    indexName := fmt.Sprintf("security-alerts-%s", indexDate)

    // Ensure the index exists
    if err := s.Client.createIndexIfNotExists(indexName); err != nil {
        return fmt.Errorf("failed to create index: %v", err)
    }

    // Convert alert to map for indexing
    alertMap := map[string]interface{}{
        "id":                alert.ID,
        "rule_id":           alert.RuleID,
        "security_event_id": alert.SecurityEventID,
        "timestamp":         alert.Timestamp,
        "severity":          alert.Severity,
        "status":            alert.Status,
        "created_at":        alert.CreatedAt,
        "updated_at":        alert.UpdatedAt,
    }

    // Only add non-nil fields
    if alert.AssignedTo != nil {
        alertMap["assigned_to"] = *alert.AssignedTo
    }
    if alert.Resolution != "" {
        alertMap["resolution"] = alert.Resolution
    }

    // Convert to JSON
    alertJSON, err := json.Marshal(alertMap)
    if err != nil {
        return err
    }

    // Index document
    url := fmt.Sprintf("%s/%s/_doc/%d", s.Client.URL, indexName, alert.ID)
    req, err := http.NewRequest("PUT", url, bytes.NewBuffer(alertJSON))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.Client.HTTPClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("failed to index alert: %s", string(body))
    }

    return nil


}

// SearchSecurityEvents searches for security events in Elasticsearch
func (s *Service) SearchSecurityEvents(query map[string]interface{}, page, pageSize int) ([]map[string]interface{}, int, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return nil, 0, fmt.Errorf("elasticsearch service not initialized")
	}

	from := (page - 1) * pageSize
	return s.Client.SearchSecurityEvents(query, from, pageSize, "last_30_days")
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
