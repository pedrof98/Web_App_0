
package models

import (
	"time"

	"gorm.io/gorm"

)

//TrafficMeasurement represents a single measurement from a sensor
type TrafficMeasurement struct {
	ID		uint 			`gorm:"primaryKey" json:"id"`
	SensorID	uint		`gorm:"not null" json:"sensor_id"`
	Timestamp	time.Time	`gorm:"not null" json:"timestamp"`
	Speed		*float64	`json:"speed"`	// Pointer allows null value
	VehicleCount	*int	`json:"vehicle_count"`	//Pointer allows null value
	CreatedAt	time.Time	`gorm:"autoCreateTime" json:"created_at"`

	// Association with Sensor
	Sensor Sensor `gorm:"foreignKey:SensorID" json:"sensor"`
}

// Sensor represents a traffic sensor.
type Sensor struct {
	ID		uint 			`gorm:"primaryKey" json:"id"`
	SensorID	string		`gorm:"unique;not null" json:"sensor_id"`
	StationID	uint		`gorm:"not null" json:"station_id"`
	MeasurementType string	`json:"measurement_type"`
	Status		string		`gorm:"default:'active'" json:"status"`

	// Association: a sensor belongs to one station and has many measurements
	Station	Station					  `gorm:"foreignKey:StationID" json:"station"`
	Measurements []TrafficMeasurement `gorm:"constraint:OnDelete:CASCADE;" json:"measurements"`
}

// Station represents a traffic station
type Station struct {
	ID			uint		`gorm:"primaryKey" json:"id"`
	Code		string		`gorm:"unique;not null" json:"code"`
	Name		string		`json:"name"`
	City		string		`json:"city"`
	Latitude	float64		`json:"latitude"`
	Longitude	float64		`json:"longitude"`
	DateOfInstallation time.Time	`json:"date_of_installation"`// use only date portion if needed

	// a station has many sensors and user events
	Sensors []Sensor		`gorm:"constraint:OnDelete:CASCADE;" json:"sensors"`
	Events  []UserEvent		`gorm:"constraint:OnDelete:CASCADE;" json:"events"`
}


// UserEvent represents an event submitted by a user (e.g., road closures)
type UserEvent struct {
	ID			uint		`gorm:"primaryKey" json:"id"`
	Date		time.Time	`gorm:"not null" json:"date"`
	City		string		`gorm:"not null" json:"city"`
	EventType	string		`gorm:"not null" json:"event_type"`
	Description	string		`json:"description"`
	ExpectedCongestionLevel string	`json:"expected_congestion_level"`
	StationID	*uint		`json:"station_id"`// optional foreign key

	// Association: an event can be linked to a station
	Station Station 		`gorm:"foreignKey:StationID" json:"station"`
}


// UserRole defines the role a user can have
type UserRole string

const (
	AdminRole UserRole = "admin"
	UserRoleUser UserRole = "user"
)

// User represents a user of the system
type User struct {
	ID			uint		`gorm:"primaryKey" json:"id"`
	Email		string		`gorm:"unique;not null" json:"email"`
	HashedPassword	string	`gorm:"not null" json:"hashed_password"`
	Role		UserRole	`gorm:"type:VARCHAR(20);default:'user'" json:"role"`
}













































