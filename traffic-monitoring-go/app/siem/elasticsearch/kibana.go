package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
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
	// Create the index pattern using the modern data views API
	body := map[string]interface{}{
		"data_view": map[string]interface{}{
			"title":         name,
			"timeFieldName": timeField,
			"name":          strings.Replace(name, "*", "", -1),
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Create the request using the modern data views API
	url := fmt.Sprintf("%s/api/data_views/data_view", c.URL)
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
	// Create the index pattern using the saved objects API
	pattern := map[string]interface{}{
		"attributes": map[string]interface{}{
			"title":         name,
			"timeFieldName": timeField,
		},
	}

	jsonPattern, err := json.Marshal(pattern)
	if err != nil {
		return err
	}

	// Create the request using the saved objects API
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

// ImportDashboard imports a dashboard into Kibana using the modern API with file upload
func (c *KibanaClient) ImportDashboard(dashboard map[string]interface{}) error {
	// The modern saved objects import API expects an NDJSON file upload
	objects, ok := dashboard["objects"].([]map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid dashboard format: missing objects array")
	}

	// Create NDJSON content (newline-delimited JSON)
	var ndjsonContent bytes.Buffer
	for _, obj := range objects {
		objJSON, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		ndjsonContent.Write(objJSON)
		ndjsonContent.WriteString("\n")
	}

	// Create multipart form data for file upload
	var formData bytes.Buffer
	writer := multipart.NewWriter(&formData)

	// Add the NDJSON file as form field
	fileWriter, err := writer.CreateFormFile("file", "dashboard.ndjson")
	if err != nil {
		return err
	}

	_, err = fileWriter.Write(ndjsonContent.Bytes())
	if err != nil {
		return err
	}

	// Close the multipart writer
	err = writer.Close()
	if err != nil {
		return err
	}

	// Create the request using the modern saved objects import API with file upload
	url := fmt.Sprintf("%s/api/saved_objects/_import?overwrite=true", c.URL)
	req, err := http.NewRequest("POST", url, &formData)
	if err != nil {
		return err
	}

	// Set required headers for multipart form data
	req.Header.Set("Content-Type", writer.FormDataContentType())
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

	log.Printf("Successfully imported dashboard using modern API with file upload")
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

// getV2XOverviewDashboard creates a dashboard with saved searches and basic visualizations
func getV2XOverviewDashboard() map[string]interface{} {
	return map[string]interface{}{
		"objects": []map[string]interface{}{
			// Simple saved search for V2X messages
			{
				"id":   "v2x-search",
				"type": "search",
				"attributes": map[string]interface{}{
					"title":       "V2X Messages Search",
					"description": "Search and view V2X message data including protocol, message type, and source information",
					"sort":        []interface{}{[]string{"@timestamp", "desc"}},
					"columns":     []string{"protocol", "message_type", "timestamp", "source_ip"},
					"kibanaSavedObjectMeta": map[string]interface{}{
						"searchSourceJSON": `{"index":"v2x-messages--wildcard","query":{"query":"","language":"kuery"},"filter":[]}`,
					},
				},
				"references": []map[string]interface{}{
					{
						"id":   "v2x-messages--wildcard",
						"name": "kibanaSavedObjectMeta.searchSourceJSON.index",
						"type": "index-pattern",
					},
				},
			},
			// Simple dashboard with just the search panel (no problematic Lens)
			{
				"id":   "v2x-overview-dashboard",
				"type": "dashboard",
				"attributes": map[string]interface{}{
					"title":       "V2X Overview Dashboard",
					"description": "Overview of V2X traffic monitoring data",
					"panelsJSON":  `[{"version":"8.10.2","type":"search","gridData":{"x":0,"y":0,"w":48,"h":20,"i":"panel-1"},"panelIndex":"panel-1","embeddableConfig":{"savedObjectId":"v2x-search"}}]`,
					"version":     1,
					"timeRestore": false,
					"kibanaSavedObjectMeta": map[string]interface{}{
						"searchSourceJSON": `{"query":{"query":"","language":"kuery"},"filter":[]}`,
					},
				},
				"references": []map[string]interface{}{
					{
						"id":   "v2x-search",
						"name": "panel_panel-1",
						"type": "search",
					},
				},
			},
		},
	}
}

// getV2XSecurityDashboard returns a simplified security-focused dashboard for V2X
func getV2XSecurityDashboard() map[string]interface{} {
	return map[string]interface{}{
		"objects": []map[string]interface{}{
			// Simple saved search for security events
			{
				"id":   "security-events-search",
				"type": "search",
				"attributes": map[string]interface{}{
					"title":       "Security Events Search",
					"description": "Search and analyze security events with severity levels and IP addresses",
					"sort":        []interface{}{[]string{"@timestamp", "desc"}},
					"columns":     []string{"severity", "source_ip", "destination_ip", "timestamp", "event_type"},
					"kibanaSavedObjectMeta": map[string]interface{}{
						"searchSourceJSON": `{"index":"security-events--wildcard","query":{"query":"","language":"kuery"},"filter":[]}`,
					},
				},
				"references": []map[string]interface{}{
					{
						"id":   "security-events--wildcard",
						"name": "kibanaSavedObjectMeta.searchSourceJSON.index",
						"type": "index-pattern",
					},
				},
			},
			// Simple dashboard with the security search
			{
				"id":   "v2x-security-dashboard",
				"type": "dashboard",
				"attributes": map[string]interface{}{
					"title":       "V2X Security Dashboard",
					"description": "Security events and anomalies in V2X communications",
					"panelsJSON":  `[{"version":"8.10.2","type":"search","gridData":{"x":0,"y":0,"w":48,"h":15,"i":"panel-1"},"panelIndex":"panel-1","embeddableConfig":{"savedObjectId":"security-events-search"}}]`,
					"version":     1,
					"timeRestore": false,
					"kibanaSavedObjectMeta": map[string]interface{}{
						"searchSourceJSON": `{"query":{"query":"","language":"kuery"},"filter":[]}`,
					},
				},
				"references": []map[string]interface{}{
					{
						"id":   "security-events-search",
						"name": "panel_panel-1",
						"type": "search",
					},
				},
			},
		},
	}
}

// getV2XMapDashboard returns a dashboard with geographic visualization using Maps app
func getV2XMapDashboard() map[string]interface{} {
	return map[string]interface{}{
		"objects": []map[string]interface{}{
			// Location search for map data
			{
				"id":   "v2x-location-search",
				"type": "search",
				"attributes": map[string]interface{}{
					"title":       "V2X Location Data",
					"description": "V2X messages with geographic coordinates",
					"sort":        []interface{}{[]string{"@timestamp", "desc"}},
					"columns":     []string{"protocol", "message_type", "timestamp", "location.lat", "location.lon", "source_id"},
					"kibanaSavedObjectMeta": map[string]interface{}{
						"searchSourceJSON": `{"index":"v2x-messages--wildcard","query":{"query":"location.lat:* AND location.lon:*","language":"kuery"},"filter":[]}`,
					},
				},
				"references": []map[string]interface{}{
					{
						"id":   "v2x-messages--wildcard",
						"name": "kibanaSavedObjectMeta.searchSourceJSON.index",
						"type": "index-pattern",
					},
				},
			},
			// Maps application visualization for geographic data
			{
				"id":   "v2x-geographic-map",
				"type": "map",
				"attributes": map[string]interface{}{
					"title":       "V2X Geographic Distribution",
					"description": "Geographic distribution of V2X messages using coordinate mapping",
					"mapStateJSON": `{
						"zoom": 6,
						"center": {
							"lat": 52.3676,
							"lon": 4.9041
						},
						"timeFilters": {
							"from": "now-15m",
							"to": "now"
						},
						"refreshConfig": {
							"isPaused": true,
							"interval": 0
						},
						"query": {
							"query": "",
							"language": "kuery"
						},
						"filters": [],
						"settings": {
							"autoFitToDataBounds": true,
							"backgroundColor": "#ffffff",
							"disableInteractive": false,
							"disableTooltipControl": false,
							"hideToolbarOverlay": false,
							"hideLayerControl": false,
							"hideViewControl": false,
							"initialLocation": "LAST_SAVED_LOCATION",
							"fixedLocation": {
								"lat": 0,
								"lon": 0,
								"zoom": 2
							},
							"browserLocation": {
								"zoom": 2
							},
							"maxZoom": 24,
							"minZoom": 0,
							"showScaleControl": false,
							"showSpatialFilters": true,
							"spatialFiltersAlpa": 0.3,
							"spatialFiltersFillColor": "#DA8B45",
							"spatialFiltersLineColor": "#DA8B45"
						}
					}`,
					"layerListJSON": `[
						{
							"id": "v2x-data-layer",
							"label": "V2X Messages",
							"minZoom": 0,
							"maxZoom": 24,
							"alpha": 0.7,
							"sourceDescriptor": {
								"type": "ES_SEARCH",
								"geoField": "location",
								"limit": 2048,
								"filterByMapBounds": true,
								"tooltipProperties": ["protocol", "message_type", "source_id", "timestamp"],
								"sortField": "@timestamp",
								"sortOrder": "desc",
								"scalingType": "LIMIT",
								"topHitsSplitField": "source_id.keyword",
								"topHitsSize": 1,
								"id": "v2x-source",
								"indexPatternRefName": "layer_1_source_index_pattern"
							},
							"style": {
								"type": "VECTOR",
								"properties": {
									"icon": {
										"type": "STATIC",
										"options": {
											"value": "marker"
										}
									},
									"fillColor": {
										"type": "DYNAMIC",
										"options": {
											"field": {
												"name": "protocol.keyword",
												"origin": "source"
											},
											"color": "Blues"
										}
									},
									"lineColor": {
										"type": "STATIC",
										"options": {
											"color": "#41937c"
										}
									},
									"lineWidth": {
										"type": "STATIC",
										"options": {
											"size": 1
										}
									},
									"iconSize": {
										"type": "DYNAMIC",
										"options": {
											"field": {
												"name": "__kbn_isClusteringCount__",
												"origin": "source"
											},
											"minSize": 7,
											"maxSize": 25
										}
									},
									"iconOrientation": {
										"type": "STATIC",
										"options": {
											"orientation": 0
										}
									},
									"labelText": {
										"type": "STATIC",
										"options": {
											"value": ""
										}
									},
									"labelColor": {
										"type": "STATIC",
										"options": {
											"color": "#000000"
										}
									},
									"labelSize": {
										"type": "STATIC",
										"options": {
											"size": 14
										}
									},
									"labelBorderColor": {
										"type": "STATIC",
										"options": {
											"color": "#FFFFFF"
										}
									},
									"symbolizeAs": {
										"options": {
											"value": "circle"
										}
									},
									"labelBorderSize": {
										"options": {
											"size": "SMALL"
										}
									}
								},
								"isTimeAware": true
							},
							"type": "VECTOR",
							"visible": true
						}
					]`,
					"uiStateJSON": `{"isLayerTOCOpen":true,"openTOCDetails":[]}`,
				},
				"references": []map[string]interface{}{
					{
						"id":   "v2x-messages--wildcard",
						"name": "layer_1_source_index_pattern",
						"type": "index-pattern",
					},
				},
			},
			// Protocol distribution chart
			{
				"id":   "v2x-protocol-pie",
				"type": "visualization",
				"attributes": map[string]interface{}{
					"title":       "V2X Protocol Distribution",
					"description": "Distribution of V2X protocols in geographic data",
					"visState": `{
						"title": "V2X Protocol Distribution",
						"type": "pie",
						"aggs": [
							{
								"id": "1",
								"enabled": true,
								"type": "count",
								"params": {},
								"schema": "metric"
							},
							{
								"id": "2",
								"enabled": true,
								"type": "terms",
								"params": {
									"field": "protocol.keyword",
									"orderBy": "1",
									"order": "desc",
									"size": 10
								},
								"schema": "segment"
							}
						],
						"params": {
							"addTooltip": true,
							"addLegend": true,
							"legendPosition": "right",
							"isDonut": false
						}
					}`,
					"uiStateJSON": `{}`,
					"kibanaSavedObjectMeta": map[string]interface{}{
						"searchSourceJSON": `{"index":"v2x-messages--wildcard","query":{"query":"location.lat:* AND location.lon:*","language":"kuery"},"filter":[]}`,
					},
				},
				"references": []map[string]interface{}{
					{
						"id":   "v2x-messages--wildcard",
						"name": "kibanaSavedObjectMeta.searchSourceJSON.index",
						"type": "index-pattern",
					},
				},
			},
			// Location-based statistics table
			{
				"id":   "v2x-location-stats",
				"type": "visualization",
				"attributes": map[string]interface{}{
					"title":       "V2X Location Statistics",
					"description": "Statistics of V2X messages by geographic regions",
					"visState": `{
						"title": "V2X Location Statistics",
						"type": "table",
						"aggs": [
							{
								"id": "1",
								"enabled": true,
								"type": "count",
								"params": {},
								"schema": "metric"
							},
							{
								"id": "2",
								"enabled": true,
								"type": "geohash_grid",
								"params": {
									"field": "location",
									"autoPrecision": true,
									"precision": 2,
									"useGeocentroid": true
								},
								"schema": "bucket"
							},
							{
								"id": "3",
								"enabled": true,
								"type": "terms",
								"params": {
									"field": "protocol.keyword",
									"orderBy": "1",
									"order": "desc",
									"size": 5
								},
								"schema": "bucket"
							}
						],
						"params": {
							"perPage": 10,
							"showPartialRows": false,
							"showMetricsAtAllLevels": false,
							"showTotal": false,
							"totalFunc": "sum",
							"percentageCol": ""
						}
					}`,
					"uiStateJSON": `{"vis":{"params":{"sort":{"columnIndex":null,"direction":null}}}}`,
					"kibanaSavedObjectMeta": map[string]interface{}{
						"searchSourceJSON": `{"index":"v2x-messages--wildcard","query":{"query":"location.lat:* AND location.lon:*","language":"kuery"},"filter":[]}`,
					},
				},
				"references": []map[string]interface{}{
					{
						"id":   "v2x-messages--wildcard",
						"name": "kibanaSavedObjectMeta.searchSourceJSON.index",
						"type": "index-pattern",
					},
				},
			},
			// Dashboard with map as main visualization
			{
				"id":   "v2x-map-dashboard",
				"type": "dashboard",
				"attributes": map[string]interface{}{
					"title":       "V2X Geographic Dashboard",
					"description": "Geographic overview of V2X communications with interactive map",
					"panelsJSON": `[
						{
							"version": "8.10.2",
							"type": "map",
							"gridData": {
								"x": 0,
								"y": 0,
								"w": 32,
								"h": 30,
								"i": "map-panel"
							},
							"panelIndex": "map-panel",
							"embeddableConfig": {
								"savedObjectId": "v2x-geographic-map"
							},
							"panelRefName": "panel_map-panel"
						},
						{
							"version": "8.10.2",
							"type": "visualization",
							"gridData": {
								"x": 32,
								"y": 0,
								"w": 16,
								"h": 15,
								"i": "protocol-panel"
							},
							"panelIndex": "protocol-panel",
							"embeddableConfig": {
								"savedObjectId": "v2x-protocol-pie"
							},
							"panelRefName": "panel_protocol-panel"
						},
						{
							"version": "8.10.2",
							"type": "visualization",
							"gridData": {
								"x": 32,
								"y": 15,
								"w": 16,
								"h": 15,
								"i": "stats-panel"
							},
							"panelIndex": "stats-panel",
							"embeddableConfig": {
								"savedObjectId": "v2x-location-stats"
							},
							"panelRefName": "panel_stats-panel"
						},
						{
							"version": "8.10.2",
							"type": "search",
							"gridData": {
								"x": 0,
								"y": 30,
								"w": 48,
								"h": 15,
								"i": "search-panel"
							},
							"panelIndex": "search-panel",
							"embeddableConfig": {
								"savedObjectId": "v2x-location-search"
							},
							"panelRefName": "panel_search-panel"
						}
					]`,
					"version":     1,
					"timeRestore": false,
					"kibanaSavedObjectMeta": map[string]interface{}{
						"searchSourceJSON": `{"query":{"query":"","language":"kuery"},"filter":[]}`,
					},
				},
				"references": []map[string]interface{}{
					{
						"id":   "v2x-geographic-map",
						"name": "panel_map-panel",
						"type": "map",
					},
					{
						"id":   "v2x-protocol-pie",
						"name": "panel_protocol-panel",
						"type": "visualization",
					},
					{
						"id":   "v2x-location-stats",
						"name": "panel_stats-panel",
						"type": "visualization",
					},
					{
						"id":   "v2x-location-search",
						"name": "panel_search-panel",
						"type": "search",
					},
				},
			},
		},
	}
}
