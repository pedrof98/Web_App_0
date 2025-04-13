

package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// KibanaClient represents a client for interacting with Kibana API
type KibanaClient struct {
	URL        string
	HTTPClient *http.Client
}

// NewKibanaClient creates a new Kibana client from an Elasticsearch URL
func NewKibanaClient(esURL string) *KibanaClient {
    // In Docker environment, replace elasticsearch with kibana in the URL
    kibanaURL := strings.Replace(esURL, "elasticsearch:9200", "kibana:5601", 1)
    // For non-Docker environments, handle standard port replacement
    if !strings.Contains(kibanaURL, "kibana") {
        kibanaURL = strings.Replace(esURL, ":9200", ":5601", 1)
    }
    
    return &KibanaClient{
        URL: kibanaURL,
        HTTPClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

// CheckAvailability checks if Kibana is available
func (c *KibanaClient) CheckAvailability() error {
    // Try the status API first (Kibana 7.x+)
    resp, err := c.HTTPClient.Get(c.URL + "/api/status")
    if err == nil && resp.StatusCode == http.StatusOK {
        resp.Body.Close()
        return nil
    }
    
    // If that fails, try the root path
    resp, err = c.HTTPClient.Get(c.URL)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("kibana returned status %d", resp.StatusCode)
    }
    
    return nil
}

// CreateIndexPattern creates an index pattern in Kibana
func (c *KibanaClient) CreateIndexPattern(name, timeField string) error {
	log.Printf("Creating Kibana index pattern: %s", name)
	
	// Check if Kibana is available
	if err := c.CheckAvailability(); err != nil {
		log.Printf("Warning: Kibana is not available: %v", err)
		return err
	}
	
	// Try different API endpoints based on Kibana version
	// First try the newer version
	err := c.createIndexPatternV8(name, timeField)
	if err != nil {
		log.Printf("Trying older API format for creating index pattern: %v", err)
		err = c.createIndexPatternLegacy(name, timeField)
	}
	
	return err
}

// createIndexPatternV8 uses the newer Kibana API (v8.x)
func (c *KibanaClient) createIndexPatternV8(name, timeField string) error {
	// Create the index pattern
	body := map[string]interface{}{
		"override": true,
		"refresh_fields": true,
		"index_pattern": map[string]interface{}{
			"title": name,
			"timeFieldName": timeField,
			"fields": "[]",
		},
	}
	
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	
	// Create the request
	url := fmt.Sprintf("%s/api/index_patterns/index_pattern", c.URL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	
	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("kbn-xsrf", "true")
	
	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create index pattern: %s", string(body))
	}
	
	log.Printf("Created Kibana index pattern %s using v8 API", name)
	return nil
}

// createIndexPatternLegacy uses the older Kibana API
func (c *KibanaClient) createIndexPatternLegacy(name, timeField string) error {
	// Create the index pattern
	pattern := map[string]interface{}{
		"attributes": map[string]interface{}{
			"title": name,
			"timeFieldName": timeField,
		},
	}
	
	jsonPattern, err := json.Marshal(pattern)
	if err != nil {
		return err
	}
	
	// Create the request
	id := strings.Replace(name, "*", "-wildcard", -1)
	url := fmt.Sprintf("%s/api/saved_objects/index-pattern/%s", c.URL, id)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPattern))
	if err != nil {
		return err
	}
	
	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("kbn-xsrf", "true")
	
	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create index pattern: %s", string(body))
	}
	
	log.Printf("Created Kibana index pattern %s using legacy API", name)
	return nil
}

// ImportDashboard imports a dashboard into Kibana
func (c *KibanaClient) ImportDashboard(dashboard map[string]interface{}) error {
	// Marshal dashboard to JSON
	jsonDashboard, err := json.Marshal(dashboard)
	if err != nil {
		return err
	}
	
	// Create the request
	url := fmt.Sprintf("%s/api/kibana/dashboards/import", c.URL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonDashboard))
	if err != nil {
		return err
	}
	
	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("kbn-xsrf", "true")
	
	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check response
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to import dashboard: %s", string(body))
	}
	
	return nil
}

// InitializeKibana sets up initial Kibana configuration for V2X SIEM
func (s *Service) InitializeKibana() error {
    log.Println("Initializing Kibana configuration for V2X SIEM...")
    
    // Create Kibana client
    kibana := NewKibanaClient(s.Client.URL)
    
    // Check if Kibana is available with retries
    const maxRetries = 10
    for i := 0; i < maxRetries; i++ {
        err := kibana.CheckAvailability()
        if err == nil {
            break
        }
        
        if i == maxRetries-1 {
            log.Printf("Warning: Kibana is not available after %d retries: %v", maxRetries, err)
            log.Printf("Kibana configuration will be skipped. Try manually later.")
            return err
        }
        
        log.Printf("Waiting for Kibana to be available (attempt %d/%d)...", i+1, maxRetries)
        time.Sleep(5 * time.Second)
    }
	
	// Create index patterns
	patterns := []struct {
		name      string
		timeField string
	}{
		{"security-events-*", "timestamp"},
		{"security-alerts-*", "timestamp"},
		{"v2x-messages-*", "timestamp"},
	}
	
	for _, pattern := range patterns {
		if err := kibana.CreateIndexPattern(pattern.name, pattern.timeField); err != nil {
			log.Printf("Warning: Failed to create index pattern %s: %v", pattern.name, err)
		}
	}
	
	// Import dashboards
	log.Println("Setting up default dashboards...")
	
	// Create the V2X Overview dashboard
	v2xOverviewDashboard := getV2XOverviewDashboard()
	if err := kibana.ImportDashboard(v2xOverviewDashboard); err != nil {
		log.Printf("Warning: Failed to import V2X Overview dashboard: %v", err)
	}
	
	// Create the V2X Security dashboard
	v2xSecurityDashboard := getV2XSecurityDashboard()
	if err := kibana.ImportDashboard(v2xSecurityDashboard); err != nil {
		log.Printf("Warning: Failed to import V2X Security dashboard: %v", err)
	}
	
	// Create the V2X Map dashboard
	v2xMapDashboard := getV2XMapDashboard()
	if err := kibana.ImportDashboard(v2xMapDashboard); err != nil {
		log.Printf("Warning: Failed to import V2X Map dashboard: %v", err)
	}
	
	log.Println("Kibana initialization completed!")
	return nil
}

// getV2XOverviewDashboard returns a basic dashboard for V2X overview
func getV2XOverviewDashboard() map[string]interface{} {
	return map[string]interface{}{
		"objects": []map[string]interface{}{
			{
				"type": "dashboard",
				"id": "v2x-overview",
				"attributes": map[string]interface{}{
					"title": "V2X Overview Dashboard",
					"description": "Overview of V2X messages and events",
					"panels": []map[string]interface{}{
						{
							"type": "visualization",
							"title": "V2X Messages by Protocol",
							"visState": map[string]interface{}{
								"title": "V2X Messages by Protocol",
								"type": "pie",
								"params": map[string]interface{}{
									"type": "pie",
									"legendPosition": "right",
								},
								"aggs": []map[string]interface{}{
									{
										"id": "1",
										"enabled": true,
										"type": "count",
										"schema": "metric",
									},
									{
										"id": "2",
										"enabled": true,
										"type": "terms",
										"schema": "segment",
										"params": map[string]interface{}{
											"field": "protocol",
											"size": 10,
										},
									},
								},
							},
							"gridData": map[string]interface{}{
								"x": 0,
								"y": 0,
								"w": 24,
								"h": 12,
								"i": "1",
							},
						},
						{
							"type": "visualization",
							"title": "V2X Messages Over Time",
							"visState": map[string]interface{}{
								"title": "V2X Messages Over Time",
								"type": "line",
								"params": map[string]interface{}{
									"addLegend": true,
									"showCircles": true,
									"interpolate": "linear",
								},
								"aggs": []map[string]interface{}{
									{
										"id": "1",
										"enabled": true,
										"type": "count",
										"schema": "metric",
									},
									{
										"id": "2",
										"enabled": true,
										"type": "date_histogram",
										"schema": "segment",
										"params": map[string]interface{}{
											"field": "timestamp",
											"useNormalizedEsInterval": true,
											"interval": "auto",
										},
									},
									{
										"id": "3",
										"enabled": true,
										"type": "terms",
										"schema": "group",
										"params": map[string]interface{}{
											"field": "message_type",
											"size": 5,
										},
									},
								},
							},
							"gridData": map[string]interface{}{
								"x": 0,
								"y": 12,
								"w": 48,
								"h": 16,
								"i": "2",
							},
						},
					},
				},
			},
		},
	}
}

// getV2XSecurityDashboard returns a security-focused dashboard for V2X
func getV2XSecurityDashboard() map[string]interface{} {
	return map[string]interface{}{
		"objects": []map[string]interface{}{
			{
				"type": "dashboard",
				"id": "v2x-security",
				"attributes": map[string]interface{}{
					"title": "V2X Security Dashboard",
					"description": "Security events and anomalies in V2X communications",
					"panels": []map[string]interface{}{
						{
							"type": "visualization",
							"title": "V2X Security Events by Severity",
							"visState": map[string]interface{}{
								"title": "V2X Security Events by Severity",
								"type": "pie",
								"params": map[string]interface{}{
									"type": "pie",
									"legendPosition": "right",
								},
								"aggs": []map[string]interface{}{
									{
										"id": "1",
										"enabled": true,
										"type": "count",
										"schema": "metric",
									},
									{
										"id": "2",
										"enabled": true,
										"type": "terms",
										"schema": "segment",
										"params": map[string]interface{}{
											"field": "severity",
											"size": 5,
										},
									},
								},
							},
							"gridData": map[string]interface{}{
								"x": 0,
								"y": 0,
								"w": 24,
								"h": 12,
								"i": "1",
							},
						},
						{
							"type": "visualization",
							"title": "V2X Anomaly Types",
							"visState": map[string]interface{}{
								"title": "V2X Anomaly Types",
								"type": "horizontal_bar",
								"params": map[string]interface{}{
									"type": "histogram",
									"grid": map[string]interface{}{
										"categoryLines": false,
									},
									"categoryAxes": []map[string]interface{}{
										{
											"position": "left",
											"show": true,
											"scale": map[string]interface{}{
												"type": "linear",
											},
										},
									},
								},
								"aggs": []map[string]interface{}{
									{
										"id": "1",
										"enabled": true,
										"type": "count",
										"schema": "metric",
									},
									{
										"id": "2",
										"enabled": true,
										"type": "terms",
										"schema": "segment",
										"params": map[string]interface{}{
											"field": "v2x.anomalies.type",
											"size": 10,
										},
									},
								},
							},
							"gridData": map[string]interface{}{
								"x": 24,
								"y": 0,
								"w": 24,
								"h": 12,
								"i": "2",
							},
						},
						{
							"type": "visualization",
							"title": "Security Events Timeline",
							"visState": map[string]interface{}{
								"title": "Security Events Timeline",
								"type": "line",
								"params": map[string]interface{}{
									"addLegend": true,
									"showCircles": true,
								},
								"aggs": []map[string]interface{}{
									{
										"id": "1",
										"enabled": true,
										"type": "count",
										"schema": "metric",
									},
									{
										"id": "2",
										"enabled": true,
										"type": "date_histogram",
										"schema": "segment",
										"params": map[string]interface{}{
											"field": "timestamp",
											"useNormalizedEsInterval": true,
											"interval": "auto",
										},
									},
									{
										"id": "3",
										"enabled": true,
										"type": "terms",
										"schema": "group",
										"params": map[string]interface{}{
											"field": "severity",
											"size": 5,
										},
									},
								},
							},
							"gridData": map[string]interface{}{
								"x": 0,
								"y": 12,
								"w": 48,
								"h": 16,
								"i": "3",
							},
						},
					},
				},
			},
		},
	}
}

// getV2XMapDashboard returns a map-based dashboard for V2X locations
func getV2XMapDashboard() map[string]interface{} {
	return map[string]interface{}{
		"objects": []map[string]interface{}{
			{
				"type": "dashboard",
				"id": "v2x-map",
				"attributes": map[string]interface{}{
					"title": "V2X Map Dashboard",
					"description": "Geographic visualization of V2X messages",
					"panels": []map[string]interface{}{
						{
							"type": "visualization",
							"title": "V2X Message Map",
							"visState": map[string]interface{}{
								"title": "V2X Message Map",
								"type": "tile_map",
								"params": map[string]interface{}{
									"mapType": "Scaled Circle Markers",
									"isDesaturated": false,
									"addTooltip": true,
									"heatClusterSize": 1.5,
									"legendPosition": "bottomright",
									"mapZoom": 2,
									"mapCenter": []float64{0, 0},
								},
								"aggs": []map[string]interface{}{
									{
										"id": "1",
										"enabled": true,
										"type": "count",
										"schema": "metric",
									},
									{
										"id": "2",
										"enabled": true,
										"type": "geohash_grid",
										"schema": "segment",
										"params": map[string]interface{}{
											"field": "v2x.location",
											"precision": 3,
										},
									},
								},
							},
							"gridData": map[string]interface{}{
								"x": 0,
								"y": 0,
								"w": 48,
								"h": 20,
								"i": "1",
							},
						},
						{
							"type": "visualization",
							"title": "V2X Messages by Vehicle ID",
							"visState": map[string]interface{}{
								"title": "V2X Messages by Vehicle ID",
								"type": "pie",
								"params": map[string]interface{}{
									"type": "pie",
									"legendPosition": "right",
								},
								"aggs": []map[string]interface{}{
									{
										"id": "1",
										"enabled": true,
										"type": "count",
										"schema": "metric",
									},
									{
										"id": "2",
										"enabled": true,
										"type": "terms",
										"schema": "segment",
										"params": map[string]interface{}{
											"field": "v2x.vehicle_id",
											"size": 10,
										},
									},
								},
							},
							"gridData": map[string]interface{}{
								"x": 0,
								"y": 20,
								"w": 24,
								"h": 12,
								"i": "2",
							},
						},
						{
							"type": "visualization",
							"title": "V2X Active Alerts Map",
							"visState": map[string]interface{}{
								"title": "V2X Active Alerts Map",
								"type": "heatmap",
								"params": map[string]interface{}{
									"type": "heatmap",
									"addTooltip": true,
									"colorsNumber": 5,
									"colorSchema": "Yellow to Red",
									"legendPosition": "right",
								},
								"aggs": []map[string]interface{}{
									{
										"id": "1",
										"enabled": true,
										"type": "count",
										"schema": "metric",
									},
									{
										"id": "2",
										"enabled": true,
										"type": "geohash_grid",
										"schema": "segment",
										"params": map[string]interface{}{
											"field": "v2x.location",
											"precision": 5,
										},
									},
								},
							},
							"gridData": map[string]interface{}{
								"x": 24,
								"y": 20,
								"w": 24,
								"h": 12,
								"i": "3",
							},
						},
					},
				},
			},
		},
	}
}