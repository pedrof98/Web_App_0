package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
	"traffic-monitoring-go/app/siem"
	"traffic-monitoring-go/app/siem/elasticsearch"
)

// Test constants
const (
	TestDSN           = "host=siem-test-db user=test_user password=test_pass dbname=test_db port=5432 sslmode=disable TimeZone=UTC"
	TestElasticsearch = "http://siem-test-es:9200"
	APIBaseURL        = "http://localhost:8080"
)

// getSIEMClient returns an HTTP client for SIEM API requests
func getSIEMClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

// getTestDB returns a test database connection
func getTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		dsn = TestDSN
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to test database")
	
	return db
}

// getElasticsearchService returns a test Elasticsearch service
func getElasticsearchService(t *testing.T) *elasticsearch.Service {
	esURL := os.Getenv("ELASTICSEARCH_URL")
	if esURL == "" {
		esURL = TestElasticsearch
	}

	// Create Elasticsearch client with custom URL
	service := elasticsearch.NewService()
	service.Client = &elasticsearch.ESClient{
		URL: esURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}

	err := service.Initialize()
	require.NoError(t, err, "Failed to initialize Elasticsearch service")
	
	return service
}

// TestLogSourceLifecycle tests the complete lifecycle of a log source
func TestLogSourceLifecycle(t *testing.T) {
	// Initialize database
	db := getTestDB(t)
	
	// Clean up existing log sources
	db.Exec("DELETE FROM log_sources")
	
	// Create a log source
	logSource := models.LogSource{
		Name:        "Test Log Source",
		Type:        models.SourceTypeSystem,
		Description: "Log source for integration testing",
		Enabled:     true,
	}
	
	err := db.Create(&logSource).Error
	require.NoError(t, err, "Failed to create log source")
	
	// Verify the log source was created
	var retrievedSource models.LogSource
	err = db.First(&retrievedSource, logSource.ID).Error
	require.NoError(t, err, "Failed to retrieve log source")
	
	assert.Equal(t, logSource.Name, retrievedSource.Name)
	assert.Equal(t, logSource.Type, retrievedSource.Type)
	assert.Equal(t, logSource.Description, retrievedSource.Description)
	assert.Equal(t, logSource.Enabled, retrievedSource.Enabled)
	
	// Update the log source
	retrievedSource.Description = "Updated description"
	err = db.Save(&retrievedSource).Error
	require.NoError(t, err, "Failed to update log source")
	
	// Verify the update
	var updatedSource models.LogSource
	err = db.First(&updatedSource, logSource.ID).Error
	require.NoError(t, err, "Failed to retrieve updated log source")
	
	assert.Equal(t, "Updated description", updatedSource.Description)
	
	// Delete the log source
	err = db.Delete(&models.LogSource{}, logSource.ID).Error
	require.NoError(t, err, "Failed to delete log source")
	
	// Verify deletion
	var deletedSource models.LogSource
	err = db.First(&deletedSource, logSource.ID).Error
	assert.Error(t, err, "Log source should be deleted")
}

// TestSecurityEventProcessing tests the complete event processing flow
func TestSecurityEventProcessing(t *testing.T) {
	// Initialize database
	db := getTestDB(t)
	
	// Initialize Elasticsearch
	esService := getElasticsearchService(t)
	
	// Clean up existing data
	db.Exec("DELETE FROM alerts")
	db.Exec("DELETE FROM security_events")
	db.Exec("DELETE FROM rules")
	db.Exec("DELETE FROM log_sources")
	
	// Create a log source
	logSource := models.LogSource{
		Name:        "Test Source",
		Type:        models.SourceTypeSystem,
		Description: "For testing",
		Enabled:     true,
	}
	
	err := db.Create(&logSource).Error
	require.NoError(t, err, "Failed to create log source")
	
	// Create a rule
	user := models.User{
		Email:          "test@example.com",
		HashedPassword: "test",
		Role:           models.AdminRole,
	}
	
	err = db.Create(&user).Error
	require.NoError(t, err, "Failed to create user")
	
	rule := models.Rule{
		Name:        "Test Rule",
		Description: "Rule for testing",
		Condition:   "severity = critical",
		Severity:    models.SeverityCritical,
		Category:    models.CategoryAuthentication,
		Status:      models.RuleStatusEnabled,
		CreatedBy:   user.ID,
	}
	
	err = db.Create(&rule).Error
	require.NoError(t, err, "Failed to create rule")
	
	// Create an event ingester
	eventIngester := siem.NewEventIngester(db)
	
	// Create an event
	rawEvent := struct {
		SourceName string                 `json:"source_name"`
		SourceType string                 `json:"source_type"`
		Timestamp  time.Time              `json:"timestamp"`
		Severity   string                 `json:"severity"`
		Category   string                 `json:"category"`
		Message    string                 `json:"message"`
		Details    map[string]interface{} `json:"details"`
	}{
		SourceName: "Test Source",
		SourceType: string(models.SourceTypeSystem),
		Timestamp:  time.Now(),
		Severity:   string(models.SeverityCritical),
		Category:   string(models.CategoryAuthentication),
		Message:    "Test event",
		Details: map[string]interface{}{
			"source_ip": "192.168.1.100",
			"username":  "admin",
			"status":    "failure",
		},
	}
	
	// Convert to JSON
	eventJSON, err := json.Marshal(rawEvent)
	require.NoError(t, err, "Failed to marshal event")
	
	// Process the event
	err = eventIngester.IngestEvent(eventJSON)
	require.NoError(t, err, "Failed to ingest event")
	
	// Verify the event was created
	var securityEvent models.SecurityEvent
	err = db.Where("message = ?", "Test event").First(&securityEvent).Error
	require.NoError(t, err, "Failed to retrieve security event")
	
	// Create a rule engine
	ruleEngine := siem.NewEnhancedRuleEngine(db)
	
	// Evaluate rules against the event
	err = ruleEngine.EvaluateEvent(&securityEvent)
	require.NoError(t, err, "Failed to evaluate rules")
	
	// Verify an alert was created
	var alert models.Alert
	err = db.Where("security_event_id = ?", securityEvent.ID).First(&alert).Error
	require.NoError(t, err, "Failed to retrieve alert")
	
	assert.Equal(t, rule.ID, alert.RuleID)
	assert.Equal(t, securityEvent.ID, alert.SecurityEventID)
	assert.Equal(t, models.SeverityCritical, alert.Severity)
	assert.Equal(t, models.AlertStatusOpen, alert.Status)
	
	// Index in Elasticsearch
	err = esService.IndexSecurityEvent(&securityEvent)
	require.NoError(t, err, "Failed to index security event in Elasticsearch")
	
	err = esService.IndexAlert(&alert)
	require.NoError(t, err, "Failed to index alert in Elasticsearch")
	
	// Give Elasticsearch time to index the documents
	time.Sleep(1 * time.Second)
	
	// Search Elasticsearch to verify indexing
	query := map[string]interface{}{
		"match": map[string]interface{}{
			"message": "Test event",
		},
	}
	
	events, total, err := esService.SearchSecurityEvents(query, 1, 10)
	require.NoError(t, err, "Failed to search Elasticsearch")
	
	assert.Greater(t, total, 0, "No events found in Elasticsearch")
	assert.Greater(t, len(events), 0, "No events returned from Elasticsearch")
}

// TestEventIngestionAPI tests the event ingestion API
func TestEventIngestionAPI(t *testing.T) {
	// Skip if no API server is running
	_, err := http.Get(APIBaseURL)
	if err != nil {
		t.Skip("API server not available, skipping test")
	}
	
	// Create a test event
	rawEvent := struct {
		SourceName string                 `json:"source_name"`
		SourceType string                 `json:"source_type"`
		Timestamp  time.Time              `json:"timestamp"`
		Severity   string                 `json:"severity"`
		Category   string                 `json:"category"`
		Message    string                 `json:"message"`
		Details    map[string]interface{} `json:"details"`
	}{
		SourceName: "API Test",
		SourceType: "system",
		Timestamp:  time.Now(),
		Severity:   "high",
		Category:   "network",
		Message:    "API test event",
		Details: map[string]interface{}{
			"source_ip": "10.0.0.1",
			"protocol":  "HTTP",
			"status":    "blocked",
		},
	}
	
	// Convert to JSON
	eventJSON, err := json.Marshal(rawEvent)
	require.NoError(t, err, "Failed to marshal event")
	
	// Send to API
	client := getSIEMClient()
	resp, err := client.Post(
		fmt.Sprintf("%s/ingest", APIBaseURL),
		"application/json",
		bytes.NewBuffer(eventJSON),
	)
	require.NoError(t, err, "Failed to send request")
	defer resp.Body.Close()
	
	// Verify response
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Unexpected status code")
	
	// Parse response
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Failed to decode response")
	
	// Verify event was created
	assert.Contains(t, result, "event_id", "Response missing event_id")
}
