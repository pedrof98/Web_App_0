package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
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

// ImportDashboardsFromNDJSON imports dashboards from an NDJSON file
func (c *KibanaClient) ImportDashboardsFromNDJSON(filePath string) error {
	log.Printf("Importing dashboards from NDJSON file: %s", filePath)

	// Read the NDJSON file
	ndjsonContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read NDJSON file: %v", err)
	}

	// Create multipart form data for file upload
	var formData bytes.Buffer
	writer := multipart.NewWriter(&formData)

	// Add the NDJSON file as form field
	fileWriter, err := writer.CreateFormFile("file", "dashboards.ndjson")
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	_, err = fileWriter.Write(ndjsonContent)
	if err != nil {
		return fmt.Errorf("failed to write file content: %v", err)
	}

	// Close the multipart writer
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close multipart writer: %v", err)
	}

	// Create the request using the modern saved objects import API
	url := fmt.Sprintf("%s/api/saved_objects/_import?overwrite=true", c.URL)
	req, err := http.NewRequest("POST", url, &formData)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set required headers for multipart form data
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("kbn-xsrf", "true")

	// Send the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to import dashboards (status %d): %s", resp.StatusCode, string(body))
	}

	// Read and log the response for debugging
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Dashboard import response: %s", string(body))
	log.Printf("Successfully imported dashboards from NDJSON file")
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

	// Import all dashboards from the NDJSON file
	log.Println("Importing dashboards from NDJSON file...")
	dashboardFilePath := "./config/dashboards.ndjson" // Adjust path as needed

	if err := kibana.ImportDashboardsFromNDJSON(dashboardFilePath); err != nil {
		log.Printf("Warning: Failed to import dashboards from NDJSON: %v", err)
		log.Println("Dashboards may need to be imported manually via Kibana UI")
	} else {
		log.Println("Successfully imported all dashboards from NDJSON file")
	}

	log.Println("Kibana initialization completed!")
	return nil

}
