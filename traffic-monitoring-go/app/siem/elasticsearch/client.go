package elasticsearch

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"strings"

	"traffic-monitoring-go/app/models"
)

// ESClient is a simple Elasticsearch client
type ESClient struct {
	URL     	string
	HTTPClient 	*http.Client
}

// NewESClient creates a new Elasticsearch client
func NewESClient() *ESClient {
	url := os.Getenv("ELASTICSEARCH_URL")
	if url == "" {
		url = "http://elasticsearch:9200" // Default URL
	}

	return &ESClient{
		URL: url,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckConnection checks if the Elasticsearch server is available
func (c *ESClient) CheckConnection() error {
	resp, err := c.HTTPClient.Get(c.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("elasticsearch returned status %d", resp.StatusCode)
	}

	return nil
}

// EnsureIndices ensures that the required indices exist
func (c *ESClient) EnsureIndices() error {
	indices := []string{
		"security-events",
		"alerts",
	}

	for _, index := range indices {
		if err := c.createIndexIfNotExists(index); err != nil {
			return fmt.Errorf("failed to create index %s: %v", index, err)
		}
	}

	return nil
}

// createIndexIfNotExists creates an index if it doesn't exist
func (c *ESClient) createIndexIfNotExists(index string) error {
	// Check if index exists
	resp, err := c.HTTPClient.Head(fmt.Sprintf("%s/%s", c.URL, index))
	if err != nil {
		return err
	}

	// If it exists, return
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// If it doesn't exist, create it
	var mappings map[string]interface{}

	// Set up mappings based on index
    if strings.HasPrefix(index, "security-events-") {
        // Security events index
        mappings = map[string]interface{}{
            "mappings": map[string]interface{}{
                "properties": map[string]interface{}{
                    "timestamp": map[string]interface{}{
                        "type": "date",
                    },
                    "severity": map[string]interface{}{
                        "type": "keyword",
                    },
                    "category": map[string]interface{}{
                        "type": "keyword",
                    },
                    "source_ip": map[string]interface{}{
                        "type": "ip",
                        "ignore_malformed": true, // Add this to handle malformed IPs
                    },
                    "destination_ip": map[string]interface{}{
                        "type": "ip",
                        "ignore_malformed": true, // Add this to handle malformed IPs
                    },
                    "message": map[string]interface{}{
                        "type": "text",
                    },
                    // Add other fields as needed
                },
            },
            "settings": map[string]interface{}{
                "number_of_shards": 1,
                "number_of_replicas": 0,
            },
        }
    } else if strings.HasPrefix(index, "security-alerts-") {
        // Alerts index
        mappings = map[string]interface{}{
            "mappings": map[string]interface{}{
                "properties": map[string]interface{}{
                    "timestamp": map[string]interface{}{
                        "type": "date",
                    },
                    "severity": map[string]interface{}{
                        "type": "keyword",
                    },
                    "status": map[string]interface{}{
                        "type": "keyword",
                    },
                    "rule_id": map[string]interface{}{
                        "type": "integer",
                    },
                    "security_event_id": map[string]interface{}{
                        "type": "integer",
                    },
                    // Add other fields as needed
                },
            },
            "settings": map[string]interface{}{
                "number_of_shards": 1,
                "number_of_replicas": 0,
            },
        }
    }

	// Create index with mappings
	mappingsJSON, err := json.Marshal(mappings)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", c.URL, index), bytes.NewBuffer(mappingsJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create index %s: %s", index, string(body))
	}

	return nil
}

// IndexSecurityEvent indexes a security event in Elasticsearch
func (c *ESClient) IndexSecurityEvent(event *models.SecurityEvent) error {
	// Convert event to map for indexing
	eventMap := map[string]interface{}{
		"id":              event.ID,
		"timestamp":       event.Timestamp,
		"source_ip":       event.SourceIP,
		"source_port":     event.SourcePort,
		"destination_ip":  event.DestinationIP,
		"destination_port": event.DestinationPort,
		"protocol":        event.Protocol,
		"action":          event.Action,
		"status":          event.Status,
		"user_id":         event.UserID,
		"device_id":       event.DeviceID,
		"log_source_id":   event.LogSourceID,
		"severity":        event.Severity,
		"category":        event.Category,
		"message":         event.Message,
		"created_at":      event.CreatedAt,
	}

	// Convert to JSON
	eventJSON, err := json.Marshal(eventMap)
	if err != nil {
		return err
	}

	// Index document
	url := fmt.Sprintf("%s/security-events/_doc/%d", c.URL, event.ID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(eventJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
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
func (c *ESClient) IndexAlert(alert *models.Alert) error {
	// Convert alert to map for indexing
	alertMap := map[string]interface{}{
		"id":               alert.ID,
		"rule_id":          alert.RuleID,
		"security_event_id": alert.SecurityEventID,
		"timestamp":        alert.Timestamp,
		"severity":         alert.Severity,
		"status":           alert.Status,
		"assigned_to":      alert.AssignedTo,
		"resolution":       alert.Resolution,
		"created_at":       alert.CreatedAt,
		"updated_at":       alert.UpdatedAt,
	}

	// Convert to JSON
	alertJSON, err := json.Marshal(alertMap)
	if err != nil {
		return err
	}

	// Index document
	url := fmt.Sprintf("%s/alerts/_doc/%d", c.URL, alert.ID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(alertJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
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
func (c *ESClient) SearchSecurityEvents(query map[string]interface{}, from, size int, timeRange string) ([]map[string]interface{}, int, error) {
    // Determine the indices to search based on timeRange
    var indexPattern string
    switch timeRange {
    case "today":
        indexPattern = fmt.Sprintf("security-events-%s", time.Now().Format("2006.01.02"))
    case "yesterday":
        yesterday := time.Now().AddDate(0, 0, -1)
        indexPattern = fmt.Sprintf("security-events-%s", yesterday.Format("2006.01.02"))
    case "last_7_days":
        indexPattern = "security-events-*"
        // Add a date range filter to the query
        if query == nil {
            query = make(map[string]interface{})
        }
        
        rangeQuery := map[string]interface{}{
            "range": map[string]interface{}{
                "timestamp": map[string]interface{}{
                    "gte": "now-7d/d",
                    "lte": "now/d",
                },
            },
        }
        
        if existingQuery, ok := query["bool"]; ok {
            // Add to existing boolean query
            boolQuery := existingQuery.(map[string]interface{})
            if must, ok := boolQuery["must"]; ok {
                mustArray := must.([]interface{})
                mustArray = append(mustArray, rangeQuery)
                boolQuery["must"] = mustArray
            } else {
                boolQuery["must"] = []interface{}{rangeQuery}
            }
        } else {
            // Create a new boolean query
            query = map[string]interface{}{
                "bool": map[string]interface{}{
                    "must": []interface{}{query, rangeQuery},
                },
            }
        }
    case "last_30_days":
		indexPattern = "security-events-*"
		if query == nil {
			query = make(map[string]interface{})
		}
		
		rangeQuery := map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"gte": "now-30d/d",
					"lte": "now/d",
				},
			},
		}
		
		if existingQuery, ok := query["bool"]; ok {
			boolQuery := existingQuery.(map[string]interface{})
			if must, ok := boolQuery["must"]; ok {
				mustArray := must.([]interface{})
				mustArray = append(mustArray, rangeQuery)
				boolQuery["must"] = mustArray
			} else {
				boolQuery["must"] = []interface{}{rangeQuery}
			}
		} else {
			query = map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []interface{}{query, rangeQuery},
				},
			}
		}
	
	case "this_month":
		indexPattern = fmt.Sprintf("security-events-%s.*", time.Now().Format("2006.01"))
		if query == nil {
			query = make(map[string]interface{})
		}
		
		startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC)
		rangeQuery := map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"gte": startOfMonth.Format(time.RFC3339),
					"lte": "now",
				},
			},
		}
		
		if existingQuery, ok := query["bool"]; ok {
			boolQuery := existingQuery.(map[string]interface{})
			if must, ok := boolQuery["must"]; ok {
				mustArray := must.([]interface{})
				mustArray = append(mustArray, rangeQuery)
				boolQuery["must"] = mustArray
			} else {
				boolQuery["must"] = []interface{}{rangeQuery}
			}
		} else {
			query = map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []interface{}{query, rangeQuery},
				},
			}
		}
	
	case "last_month":
		lastMonth := time.Now().AddDate(0, -1, 0)
		indexPattern = fmt.Sprintf("security-events-%s.*", lastMonth.Format("2006.01"))
		if query == nil {
			query = make(map[string]interface{})
		}
		
		startOfLastMonth := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
		endOfLastMonth := startOfLastMonth.AddDate(0, 1, 0).Add(-time.Second)
		rangeQuery := map[string]interface{}{
			"range": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"gte": startOfLastMonth.Format(time.RFC3339),
					"lte": endOfLastMonth.Format(time.RFC3339),
				},
			},
		}
		
		if existingQuery, ok := query["bool"]; ok {
			boolQuery := existingQuery.(map[string]interface{})
			if must, ok := boolQuery["must"]; ok {
				mustArray := must.([]interface{})
				mustArray = append(mustArray, rangeQuery)
				boolQuery["must"] = mustArray
			} else {
				boolQuery["must"] = []interface{}{rangeQuery}
			}
		} else {
			query = map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []interface{}{query, rangeQuery},
				},
			}
		}
		
    default:
        indexPattern = "security-events-*"
    }

    // Add pagination parameters
    searchQuery := map[string]interface{}{
        "query": query,
        "from":  from,
        "size":  size,
        "sort": []map[string]interface{}{
            {
                "timestamp": map[string]interface{}{
                    "order": "desc",
                },
            },
        },
    }

    searchJSON, err := json.Marshal(searchQuery)
    if err != nil {
        return nil, 0, err
    }

    // Execute search
    url := fmt.Sprintf("%s/%s/_search", c.URL, indexPattern)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(searchJSON))
    if err != nil {
        return nil, 0, err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, 0, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, 0, fmt.Errorf("failed to search security events: %s", string(body))
    }

    // Parse response
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, 0, err
    }

    // Extract hits
    hitsMap, ok := result["hits"].(map[string]interface{})
    if !ok {
        return nil, 0, errors.New("unexpected response format: missing hits")
    }

    totalMap, ok := hitsMap["total"].(map[string]interface{})
    if !ok {
        return nil, 0, errors.New("unexpected response format: missing total")
    }

    totalValue, ok := totalMap["value"].(float64)
    if !ok {
        return nil, 0, errors.New("unexpected response format: missing total value")
    }
    total := int(totalValue)

    hitsArray, ok := hitsMap["hits"].([]interface{})
    if !ok {
        return nil, total, errors.New("unexpected response format: hits is not an array")
    }

    // Extract events from hits
    events := make([]map[string]interface{}, 0, len(hitsArray))
    for _, hit := range hitsArray {
        hitMap, ok := hit.(map[string]interface{})
        if !ok {
            continue
        }

        source, ok := hitMap["_source"].(map[string]interface{})
        if !ok {
            continue
        }

        events = append(events, source)
    }

    return events, total, nil
}

// GetEventDashboardStats returns statistics for the dashboard
func (c *ESClient) GetEventDashboardStats(timeRange string) (map[string]interface{}, error) {
	// Build time filter
	timeFilter := buildTimeFilter(timeRange)

	// Build aggregation query
	queryMap := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					timeFilter,
				},
			},
		},
		"aggs": map[string]interface{}{
			"severity_counts": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "severity",
				},
			},
			"category_counts": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "category",
				},
			},
			"events_over_time": map[string]interface{}{
				"date_histogram": map[string]interface{}{
					"field":    "timestamp",
					"interval": "hour",
				},
			},
		},
	}

	queryJSON, err := json.Marshal(queryMap)
	if err != nil {
		return nil, err
	}

	// Execute search
	url := fmt.Sprintf("%s/security-events/_search", c.URL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(queryJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get dashboard stats: %s", string(body))
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Helper function to build time filter for Elasticsearch
func buildTimeFilter(timeRange string) map[string]interface{} {
	now := time.Now()
	var startTime time.Time

	switch timeRange {
	case "today":
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "yesterday":
		yesterday := now.AddDate(0, 0, -1)
		startTime = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, now.Location())
	case "last_7_days":
		startTime = now.AddDate(0, 0, -7)
	case "last_30_days":
		startTime = now.AddDate(0, 0, -30)
	case "this_month":
		startTime = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	case "last_month":
		lastMonth := now.AddDate(0, -1, 0)
		startTime = time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, now.Location())
	case "this_year":
		startTime = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	default:
		// Default to last 24 hours
		startTime = now.Add(-24 * time.Hour)
	}

	return map[string]interface{}{
		"range": map[string]interface{}{
			"timestamp": map[string]interface{}{
				"gte": startTime.Format(time.RFC3339),
				"lte": now.Format(time.RFC3339),
			},
		},
	}
}
