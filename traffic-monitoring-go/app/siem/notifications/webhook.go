// File: app/siem/notifications/webhook.go

package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"traffic-monitoring-go/app/models"
)

// WebhookConfig contains configuration for webhook notifications
type WebhookConfig struct {
	BaseNotificationConfig
	URL             string            `json:"url"`
	Method          string            `json:"method"`
	Headers         map[string]string `json:"headers"`
	TimeoutSeconds  int               `json:"timeout_seconds"`
}

// WebhookChannel sends notifications via webhook
type WebhookChannel struct {
	Config WebhookConfig
	Client *http.Client
}

// NewWebhookChannel creates a new WebhookChannel
func NewWebhookChannel(config WebhookConfig) *WebhookChannel {
	// Set defaults if not specified
	if config.Method == "" {
		config.Method = "POST"
	}
	
	if config.TimeoutSeconds <= 0 {
		config.TimeoutSeconds = 10
	}
	
	return &WebhookChannel{
		Config: config,
		Client: &http.Client{
			Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
		},
	}
}

// Name returns the channel's name
func (c *WebhookChannel) Name() string {
	return c.Config.Name
}

// Type returns the channel's type
func (c *WebhookChannel) Type() string {
	return "webhook"
}

// Send sends a webhook notification for an alert
func (c *WebhookChannel) Send(alert *models.Alert) error {
	if !c.Config.Enabled {
		return nil // Channel is disabled, no-op
	}
	
	// Make sure we have a URL
	if c.Config.URL == "" {
		return fmt.Errorf("no webhook URL configured")
	}
	
	// Load related data if not already loaded
	if alert.Rule.ID == 0 {
		return fmt.Errorf("rule data not loaded for alert")
	}
	
	if alert.SecurityEvent.ID == 0 {
		return fmt.Errorf("security event data not loaded for alert")
	}
	
	// Prepare payload
	payload := struct {
		AlertID     uint                 `json:"alert_id"`
		RuleID      uint                 `json:"rule_id"`
		RuleName    string               `json:"rule_name"`
		EventID     uint                 `json:"event_id"`
		Timestamp   time.Time            `json:"timestamp"`
		Severity    models.EventSeverity `json:"severity"`
		Status      models.AlertStatus   `json:"status"`
		Message     string               `json:"message"`
		Category    models.EventCategory `json:"category"`
		SourceIP    string               `json:"source_ip,omitempty"`
		Description string               `json:"description,omitempty"`
	}{
		AlertID:     alert.ID,
		RuleID:      alert.RuleID,
		RuleName:    alert.Rule.Name,
		EventID:     alert.SecurityEventID,
		Timestamp:   alert.Timestamp,
		Severity:    alert.Severity,
		Status:      alert.Status,
		Message:     alert.SecurityEvent.Message,
		Category:    alert.SecurityEvent.Category,
		SourceIP:    alert.SecurityEvent.SourceIP,
		Description: alert.Rule.Description,
	}
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %v", err)
	}
	
	// Create request
	req, err := http.NewRequest(c.Config.Method, c.Config.URL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %v", err)
	}
	
	// Set Content-Type header if not specified
	if _, ok := c.Config.Headers["Content-Type"]; !ok {
		req.Header.Set("Content-Type", "application/json")
	}
	
	// Add custom headers
	for key, value := range c.Config.Headers {
		req.Header.Set(key, value)
	}
	
	// Send the request
	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %v", err)
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}
	
	log.Printf("Sent webhook notification for alert %d to %s", alert.ID, c.Config.URL)
	return nil
}
