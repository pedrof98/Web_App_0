

package notifications

import (
	"bytes"
	
	"fmt"
	"log"
	"net/smtp"
	"text/template"
	"time"

	"traffic-monitoring-go/app/models"
)

// EmailConfig contains configuration for email notifications
type EmailConfig struct {
	BaseNotificationConfig
	SMTPServer   string   `json:"smtp_server"`
	SMTPPort     int      `json:"smtp_port"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	FromAddress  string   `json:"from_address"`
	ToAddresses  []string `json:"to_addresses"`
	SubjectTemplate string `json:"subject_template"`
	BodyTemplate    string `json:"body_template"`
}

// EmailChannel sends notifications via email
type EmailChannel struct {
	Config EmailConfig
}

// NewEmailChannel creates a new EmailChannel
func NewEmailChannel(config EmailConfig) *EmailChannel {
	// Set default templates if not provided
	if config.SubjectTemplate == "" {
		config.SubjectTemplate = "Security Alert: {{ .Rule.Name }} - {{ .Alert.Severity }}"
	}

	if config.BodyTemplate == "" {
		config.BodyTemplate = `
Security Alert

Rule: {{ .Rule.Name }}
Severity: {{ .Alert.Severity }}
Time: {{ .Alert.Timestamp }}
Status: {{ .Alert.Status }}

Event Details:
- Category: {{ .Event.Category }}
- Source IP: {{ .Event.SourceIP }}
- Message: {{ .Event.Message }}

Please check your SIEM system for more details.
`
	}

	return &EmailChannel{
		Config: config,
	}
}

// Name returns the channel's name
func (c *EmailChannel) Name() string {
	return c.Config.Name
}

// Type returns the channel's type
func (c *EmailChannel) Type() string {
	return "email"
}

// Send sends an email notification for an alert
func (c *EmailChannel) Send(alert *models.Alert) error {
	if !c.Config.Enabled {
		return nil // Channel is disabled, no-op
	}

	// Make sure we have all required data
	if len(c.Config.ToAddresses) == 0 {
		return fmt.Errorf("no recipient addresses configured")
	}

	// Load related data if not already loaded
	if alert.Rule.ID == 0 {
		return fmt.Errorf("rule data not loaded for alert")
	}

	if alert.SecurityEvent.ID == 0 {
		return fmt.Errorf("security event data not loaded for alert")
	}

	// Prepare template data
	data := struct {
		Alert *models.Alert
		Rule  *models.Rule
		Event *models.SecurityEvent
		Time  time.Time
	}{
		Alert: alert,
		Rule:  &alert.Rule,
		Event: &alert.SecurityEvent,
		Time:  time.Now(),
	}

	// Parse and execute the subject template
	subjectTmpl, err := template.New("subject").Parse(c.Config.SubjectTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse subject template: %v", err)
	}

	var subjectBuf bytes.Buffer
	if err := subjectTmpl.Execute(&subjectBuf, data); err != nil {
		return fmt.Errorf("failed to execute subject template: %v", err)
	}
	subject := subjectBuf.String()

	// Parse and execute the body template
	bodyTmpl, err := template.New("body").Parse(c.Config.BodyTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse body template: %v", err)
	}

	var bodyBuf bytes.Buffer
	if err := bodyTmpl.Execute(&bodyBuf, data); err != nil {
		return fmt.Errorf("failed to execute body template: %v", err)
	}
	body := bodyBuf.String()

	// Compose the email
	auth := smtp.PlainAuth("", c.Config.Username, c.Config.Password, c.Config.SMTPServer)

	// Construct the message
	to := bytes.NewBufferString("To: ")
	for i, addr := range c.Config.ToAddresses {
		if i > 0 {
			to.WriteString(", ")
		}
		to.WriteString(addr)
	}

	msg := []byte(fmt.Sprintf("%s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=utf-8\r\n"+
		"\r\n"+
		"%s", to.String(), c.Config.FromAddress, subject, body))

	// Send the email
	err = smtp.SendMail(
		fmt.Sprintf("%s:%d", c.Config.SMTPServer, c.Config.SMTPPort),
		auth,
		c.Config.FromAddress,
		c.Config.ToAddresses,
		msg,
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("Sent email notification for alert %d to %d recipients", alert.ID, len(c.Config.ToAddresses))
	return nil
}
