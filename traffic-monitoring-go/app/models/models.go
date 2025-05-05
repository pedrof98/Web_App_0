package models

import "time"

// Station represents a traffic station.
type Station struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	Code               string    `gorm:"unique;not null" json:"code"`
	Name               string    `json:"name"`
	City               string    `json:"city"`
	Latitude           float64   `json:"latitude"`
	Longitude          float64   `json:"longitude"`
	DateOfInstallation time.Time `json:"date_of_installation"`
	Sensors            []Sensor  `gorm:"constraint:OnDelete:CASCADE;" json:"sensors"`
	Events             []UserEvent `gorm:"constraint:OnDelete:CASCADE;" json:"events"`
}

// TableName returns the table name for Station.
func (Station) TableName() string {
	return "stations"
}

// Sensor represents a traffic sensor.
type Sensor struct {
	ID              uint                `gorm:"primaryKey" json:"id"`
	SensorID        string              `gorm:"unique;not null" json:"sensor_id"`
	StationID       uint                `gorm:"not null" json:"station_id"`
	MeasurementType string              `json:"measurement_type"`
	Status          string              `json:"status"`
	Station         Station             `gorm:"foreignKey:StationID;references:ID" json:"station"`
	Measurements    []TrafficMeasurement `gorm:"-" json:"measurements"`
}

// TableName returns the table name for Sensor.
func (Sensor) TableName() string {
	return "sensors"
}

// TrafficMeasurement represents a single measurement from a sensor.
type TrafficMeasurement struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	SensorID     uint      `gorm:"not null" json:"sensor_id"`
	Timestamp    time.Time `gorm:"not null" json:"timestamp"`
	Speed        *float64  `json:"speed"`
	VehicleCount *int      `json:"vehicle_count"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	Sensor       Sensor    `gorm:"foreignKey:SensorID;references:ID;constraint:OnDelete:CASCADE" json:"sensor"`
}

// TableName returns the table name for TrafficMeasurement.
func (TrafficMeasurement) TableName() string {
	return "traffic_measurements"
}

// UserEvent represents an event submitted by a user.
type UserEvent struct {
	ID                      uint      `gorm:"primaryKey" json:"id"`
	Date                    time.Time `gorm:"not null" json:"date"`
	City                    string    `gorm:"not null" json:"city"`
	EventType               string    `gorm:"not null" json:"event_type"`
	Description             string    `json:"description"`
	ExpectedCongestionLevel string    `json:"expected_congestion_level"`
	StationID               *uint     `json:"station_id"`
	Station                 Station   `gorm:"foreignKey:StationID;references:ID" json:"station"`
}

// TableName returns the table name for UserEvent.
func (UserEvent) TableName() string {
	return "user_events"
}

// UserRole defines the role a user can have.
type UserRole string

const (
	AdminRole    UserRole = "admin"
	UserRoleUser UserRole = "user"
)

// User represents a user of the system.
type User struct {
	ID             uint     `gorm:"primaryKey" json:"id"`
	Email          string   `gorm:"unique;not null" json:"email"`
	HashedPassword string   `gorm:"not null" json:"hashed_password"`
	Role           UserRole `gorm:"type:VARCHAR(20)" json:"role"`
}

// TableName returns the table name for User.
func (User) TableName() string {
	return "users"
}
