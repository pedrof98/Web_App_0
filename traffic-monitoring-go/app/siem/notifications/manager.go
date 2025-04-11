package notifications

import (
	
	"fmt"
	"log"
	"sync"

	"gorm.io/gorm"
	"traffic-monitoring-go/app/models"
)

// NotificationManager manages all notification channels
type NotificationManager struct {
	DB		*gorm.DB
	channels	map[string]NotificationChannel
	mutex		sync.Mutex
}

//NewNotificationManager creates a new NotificationManager
func NewNotificationManager(db *gorm.DB) *NotificationManager {
	return &NotificationManager{
		DB:		db,
		channels:	make(map[string]NotificationChannel),
	}
}

// RegisterChannel adds a notification channel to the manager
func (m *NotificationManager) RegisterChannel(channel NotificationChannel) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	name := channel.Name()
	if _, exists := m.channels[name]; exists {
		return fmt.Errorf("channel with name '%s' already registered", name)
	}

	m.channels[name] = channel
	log.Printf("Registered notification channel: %s (%s)", name, channel.Type())
	return nil
}

// SendNotification sends a notification for an alert through all enabled channels
func (m *NotificationManager) SendNotification(alertID uint) error {
	// Load the alert with related data
	var alert models.Alert
	if err := m.DB.Preload("Rule").Preload("SecurityEvent").First(&alert, alertID).Error; err != nil {
		return fmt.Errorf("failed to load alert %d: %v", alertID, err)
	}

	// Send through each channel
	m.mutex.Lock()
	channels := make([]NotificationChannel, 0, len(m.channels))
	for _, channel := range m.channels {
		channels = append(channels, channel)
	}
	m.mutex.Unlock()

	var errs []error
	var successCount int

	for _, channel := range channels {
		if err := channel.Send(&alert); err != nil {
			log.Printf("Error sending notification through channel '%s': %v", channel.Name(), err)
			errs = append(errs, fmt.Errorf("channel '%s': %v", channel.Name(), err))
		} else {
			successCount++
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to send notifications through %d channels (succeeded: %d): %v",
			len(errs), successCount, errs[0])
	}

	log.Printf("Successfully sent notifications for alert %d through %d channels", alertID, successCount)
	return nil
}

// GetChannelNames returns the names of all registered channels
func (m *NotificationManager) GetChannelNames() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	names := make([]string, 0, len(m.channels))
	for name := range m.channels {
		names = append(names, name)
	}

	return names
}
