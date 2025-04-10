package notifications

import (
	"traffic-monitoring-go/app/models"
)


// NotificationChannel defines the interface for sending notifications
type NotificationChannel interface {
	// send sends a notification about an alert
	Send(alert *models.Alert) error
	// name returns the channel's name
	Name() string
	// type returns the channel's type
	Type() string
}


// BaseNotificationConfig contains common configuration for all notification channels
type BaseNotificationConfig struct {
	Enabled 	bool	`json:"enabled"`
	Name		string	`json:"name"`
}


