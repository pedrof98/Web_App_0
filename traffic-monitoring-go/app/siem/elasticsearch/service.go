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
	"gorm.io/gorm"



	"traffic-monitoring-go/app/models"
)

// Service is a service for interacting with Elasticsearch
type Service struct {
	Client      *ESClient
	DB		    *gorm.DB
	initialized bool
	mutex       sync.RWMutex
}

// NewService creates a new Elasticsearch Service
func NewService(db *gorm.DB) *Service {
	return &Service{
		Client:      NewESClient(),
		DB:          db,
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
	
	// Initialize Kibana (this happens after we mark as initialized to avoid deadlocks)
	go func() {
		// Wait a bit to ensure Elasticsearch has fully started
		time.Sleep(20 * time.Second)
		
		// This is run in a goroutine so it doesn't block the main initialization
		if err := s.InitializeKibana(); err != nil {
			log.Printf("Warning: Failed to initialize Kibana: %v", err)
			log.Println("Kibana visualizations and dashboards may need to be set up manually")
		} else {
			log.Println("Kibana initialized successfully with V2X SIEM dashboards")
		}
	}()
	
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
                        "fields": map[string]interface{}{
                            "keyword": map[string]interface{}{
                                "type": "keyword",
                                "ignore_above": 256,
                            },
                        },
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
                    // V2X-specific fields
                    "v2x": map[string]interface{}{
                        "properties": map[string]interface{}{
                            "protocol": map[string]interface{}{
                                "type": "keyword",
                            },
                            "message_type": map[string]interface{}{
                                "type": "keyword",
                            },
                            "vehicle_id": map[string]interface{}{
                                "type": "keyword", 
                            },
                            "location": map[string]interface{}{
                                "type": "geo_point",
                            },
                            "speed": map[string]interface{}{
                                "type": "float",
                            },
                            "heading": map[string]interface{}{
                                "type": "float",
                            },
                            "rssi": map[string]interface{}{
                                "type": "integer",
                            },
                            "anomalies": map[string]interface{}{
                                "type": "nested",
                                "properties": map[string]interface{}{
                                    "type": map[string]interface{}{"type": "keyword"},
                                    "confidence": map[string]interface{}{"type": "float"},
                                    "description": map[string]interface{}{"type": "text"},
                                },
                            },
                        },
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
    
    // Create template specifically for V2X messages
    v2xTemplate := map[string]interface{}{
        "index_patterns": []string{"v2x-messages-*"},
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
                    "protocol": map[string]interface{}{
                        "type": "keyword",
                    },
                    "message_type": map[string]interface{}{
                        "type": "keyword",
                    },
                    "source_id": map[string]interface{}{
                        "type": "keyword",
                    },
                    "location": map[string]interface{}{
                        "type": "geo_point",
                    },
                    "speed": map[string]interface{}{
                        "type": "float",
                    },
                    "heading": map[string]interface{}{
                        "type": "float",
                    },
                    "rssi": map[string]interface{}{
                        "type": "integer",
                    },
                    "message_count": map[string]interface{}{
                        "type": "integer",
                    },
                    "interface_type": map[string]interface{}{
                        "type": "keyword",
                    },
                    "security": map[string]interface{}{
                        "properties": map[string]interface{}{
                            "signature_valid": map[string]interface{}{"type": "boolean"},
                            "trust_level": map[string]interface{}{"type": "integer"},
                            "certificate_id": map[string]interface{}{"type": "keyword"},
                        },
                    },
                    "anomalies": map[string]interface{}{
                        "type": "nested",
                        "properties": map[string]interface{}{
                            "type": map[string]interface{}{"type": "keyword"},
                            "confidence": map[string]interface{}{"type": "float"},
                            "description": map[string]interface{}{"type": "text"},
                        },
                    },
                    "created_at": map[string]interface{}{
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
    
    v2xJSON, err := json.Marshal(v2xTemplate)
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
    
    // Create V2X template
    req, err = http.NewRequest("PUT", fmt.Sprintf("%s/_index_template/v2x-messages-template", s.Client.URL), bytes.NewBuffer(v2xJSON))
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
        return fmt.Errorf("failed to create v2x template: %s", string(body))
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
	indexDate := event.Timestamp.Format("2006.01.02")
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
	
	// Special handling for V2X data
	if event.Category == models.CategoryV2X && event.RawData != "" {
		var rawData map[string]interface{}
		if err := json.Unmarshal([]byte(event.RawData), &rawData); err == nil {
			v2xDetails := make(map[string]interface{})
			
			// Extract V2X specific details from the raw data
			if details, ok := rawData["details"].(map[string]interface{}); ok {
				// Extract message type
				if msgType, ok := details["message_type"].(string); ok {
					v2xDetails["message_type"] = msgType
				}
				
				// Extract vehicle ID
				if vehicleID, ok := details["vehicle_id"].(string); ok {
					v2xDetails["vehicle_id"] = vehicleID
				}
				
				// Extract position as geo_point
				if position, ok := details["position"].(map[string]interface{}); ok {
					if lat, latOk := position["latitude"].(float64); latOk {
						if lon, lonOk := position["longitude"].(float64); lonOk {
							v2xDetails["location"] = map[string]interface{}{
								"lat": lat,
								"lon": lon,
							}
						}
					}
				}
				
				// Extract speed and heading
				if speed, ok := details["speed"].(float64); ok {
					v2xDetails["speed"] = speed
				}
				if heading, ok := details["heading"].(float64); ok {
					v2xDetails["heading"] = heading
				}
				
				// Extract protocol
				if protocol, ok := details["protocol"].(string); ok {
					v2xDetails["protocol"] = protocol
				}
				
				// Extract interface type for C-V2X
				if interfaceType, ok := details["interface_type"].(string); ok {
					v2xDetails["interface_type"] = interfaceType
				}
				
				// Extract security information
				if sigValid, ok := details["signature_valid"].(bool); ok {
					secInfo := map[string]interface{}{
						"signature_valid": sigValid,
					}
					if trustLevel, ok := details["trust_level"].(float64); ok {
						secInfo["trust_level"] = trustLevel
					}
					if certID, ok := details["certificate_id"].(string); ok {
						secInfo["certificate_id"] = certID
					}
					v2xDetails["security"] = secInfo
				}
				
				// Extract anomalies if present
				if anomalies, ok := details["anomalies"].([]interface{}); ok && len(anomalies) > 0 {
					v2xDetails["anomalies"] = anomalies
				}
			}
			
			// Add all V2X details to the event map
			if len(v2xDetails) > 0 {
				eventMap["v2x"] = v2xDetails
			}
		}
	}

	// For V2X events, also index them in a separate V2X index
	if event.Category == models.CategoryV2X && event.RawData != "" {
		// Create a time-based index name for V2X specific data
		v2xIndexName := fmt.Sprintf("v2x-messages-%s", indexDate)
		
		// Ensure the V2X index exists
		if err := s.Client.createIndexIfNotExists(v2xIndexName); err != nil {
			log.Printf("Warning: failed to create V2X index: %v", err)
			// Continue anyway - we'll still index in the main events index
		} else {
			// Create a copy of the event map with just V2X-specific data
			v2xEventMap := map[string]interface{}{
				"id":          event.ID,
				"timestamp":   event.Timestamp,
				"severity":    event.Severity,
				"message":     event.Message,
				"created_at":  event.CreatedAt,
			}
			
			// Add the V2X data if available
			if v2xData, ok := eventMap["v2x"].(map[string]interface{}); ok {
				for k, v := range v2xData {
					v2xEventMap[k] = v
				}
			}
			
			// Index in the V2X-specific index
			v2xJSON, err := json.Marshal(v2xEventMap)
			if err != nil {
				log.Printf("Warning: failed to marshal V2X event: %v", err)
			} else {
				url := fmt.Sprintf("%s/%s/_doc/%d", s.Client.URL, v2xIndexName, event.ID)
				req, err := http.NewRequest("PUT", url, bytes.NewBuffer(v2xJSON))
				if err != nil {
					log.Printf("Warning: failed to create V2X index request: %v", err)
				} else {
					req.Header.Set("Content-Type", "application/json")
					resp, err := s.Client.HTTPClient.Do(req)
					if err != nil {
						log.Printf("Warning: failed to index V2X event: %v", err)
					} else {
						if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
							body, _ := io.ReadAll(resp.Body)
							log.Printf("Warning: failed to index V2X event: %s", string(body))
						}
						resp.Body.Close()
					}
				}
			}
		}
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


// IndexV2XMessage directly indexes a V2X message in Elasticsearch
func (s *Service) IndexV2XMessage(v2xMessage *models.V2XMessage) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if !s.initialized {
		return fmt.Errorf("elasticsearch service not initialized")
	}

	// Create a time-based index name for V2X messages
	indexDate := v2xMessage.Timestamp.Format("2006.01.02")
	indexName := fmt.Sprintf("v2x-messages-%s", indexDate)

	// Ensure index exists
	if err := s.Client.createIndexIfNotExists(indexName); err != nil {
		return fmt.Errorf("failed to create index: %v", err)
	}

	// Create document data for the V2X message
	docData := map[string]interface{}{
		"id":           v2xMessage.ID,
		"timestamp":    v2xMessage.Timestamp,
		"protocol":     v2xMessage.Protocol,
		"message_type": v2xMessage.MessageType,
		"source_id":    v2xMessage.SourceID,
		"rssi":         v2xMessage.RSSI,
		"created_at":   v2xMessage.CreatedAt,
		// Format location as geo_point
		"location": map[string]interface{}{
			"lat": v2xMessage.Latitude,
			"lon": v2xMessage.Longitude,
		},
	}

	// Try to get additional data based on message type
	switch v2xMessage.MessageType {
	case models.MessageTypeBSM, models.MessageTypeCV2XBSM, models.MessageTypeCAM:
		// Get BSM data
		var bsm models.BasicSafetyMessage
		if err := s.DB.Where("v2x_message_id = ?", v2xMessage.ID).First(&bsm).Error; err == nil {
			docData["speed"] = bsm.Speed
			docData["heading"] = bsm.Heading
			docData["vehicle_id"] = fmt.Sprintf("%08X", bsm.TemporaryID)
			
			// Add vehicle-specific fields
			if bsm.Width > 0 && bsm.Length > 0 {
				docData["vehicle_size"] = map[string]interface{}{
					"width":  bsm.Width,
					"length": bsm.Length,
					"height": bsm.Height,
				}
			}
			
			// Add acceleration data if available
			if bsm.LateralAccel != 0 || bsm.LongitudinalAccel != 0 {
				docData["acceleration"] = map[string]interface{}{
					"lateral":      bsm.LateralAccel,
					"longitudinal": bsm.LongitudinalAccel,
					"yaw_rate":     bsm.YawRate,
				}
			}
		}
		
		// For C-V2X, get interface type
		if v2xMessage.Protocol == models.ProtocolCV2XMode4 || v2xMessage.Protocol == models.ProtocolCV2XUu {
			var cv2xInfo models.CV2XMessage
			if err := s.DB.Where("v2x_message_id = ?", v2xMessage.ID).First(&cv2xInfo).Error; err == nil {
				docData["interface_type"] = cv2xInfo.InterfaceType
				docData["qos_info"] = cv2xInfo.QoSInfo
				if cv2xInfo.PLMNInfo != "" {
					docData["plmn_info"] = cv2xInfo.PLMNInfo
				}
			}
		}
		
	case models.MessageTypeRSA, models.MessageTypeDENM:
		// Get roadside alert data
		var rsa models.RoadsideAlert
		if err := s.DB.Where("v2x_message_id = ?", v2xMessage.ID).First(&rsa).Error; err == nil {
			docData["alert_type"] = rsa.AlertType
			docData["description"] = rsa.Description
			docData["priority"] = rsa.Priority
			docData["radius"] = rsa.Radius
			docData["duration"] = rsa.Duration
		}
		
	case models.MessageTypeSPAT:
		// Get signal phase and timing data
		var spat models.SignalPhaseAndTiming
		if err := s.DB.Where("v2x_message_id = ?", v2xMessage.ID).First(&spat).Error; err == nil {
			docData["intersection_id"] = spat.IntersectionID
			
			// Get phase states
			var phases []models.PhaseState
			if err := s.DB.Where("spat_message_id = ?", spat.ID).Find(&phases).Error; err == nil {
				phaseData := make([]map[string]interface{}, len(phases))
				for i, phase := range phases {
					phaseData[i] = map[string]interface{}{
						"phase_id":    phase.PhaseID,
						"light_state": phase.LightState,
						"start_time":  phase.StartTime,
						"min_end_time": phase.MinEndTime,
						"max_end_time": phase.MaxEndTime,
					}
				}
				docData["phases"] = phaseData
			}
		}
	}
	
	// Add security information if available
	var securityInfo models.V2XSecurityInfo
	if err := s.DB.Where("v2x_message_id = ?", v2xMessage.ID).First(&securityInfo).Error; err == nil {
		docData["security"] = map[string]interface{}{
			"signature_valid":  securityInfo.SignatureValid,
			"trust_level":      securityInfo.TrustLevel,
			"certificate_id":   securityInfo.CertificateID,
			"validation_error": securityInfo.ValidationError,
		}
	}
	
	// Add anomaly detections if available
	var anomalies []models.V2XAnomalyDetection
	if err := s.DB.Where("v2x_message_id = ?", v2xMessage.ID).Find(&anomalies).Error; err == nil && len(anomalies) > 0 {
		anomalyData := make([]map[string]interface{}, len(anomalies))
		for i, anomaly := range anomalies {
			anomalyData[i] = map[string]interface{}{
				"type":         anomaly.AnomalyType,
				"confidence":   anomaly.ConfidenceScore,
				"description":  anomaly.Description,
				"created_at":   anomaly.CreatedAt,
			}
		}
		docData["anomalies"] = anomalyData
	}

	// Convert to JSON
	jsonData, err := json.Marshal(docData)
	if err != nil {
		return err
	}

	// Index the document
	url := fmt.Sprintf("%s/%s/_doc/%d", s.Client.URL, indexName, v2xMessage.ID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
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
		return fmt.Errorf("failed to index V2X message: %s", string(body))
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
