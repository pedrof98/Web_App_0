package repository

import (
    "context"
    "errors"
    "fmt"
    "strings"
    "time"

    "gorm.io/gorm"
    "traffic-monitoring-go/internal/domain"
    "traffic-monitoring-go/internal/dto"
)

// GormSecurityEventRepository implements SecurityEventRepository using Gorm
type GormSecurityEventRepository struct {
    db *gorm.DB
}

// NewGormSecurityEventRepository creates a new GormSecurityEventRepository
func NewGormSecurityEventRepository(db *gorm.DB) *GormSecurityEventRepository {
    return &GormSecurityEventRepository{db: db}
}

// dbSecurityEvent is the database model for security events
type dbSecurityEvent struct {
    ID              uint      `gorm:"primaryKey"`
    Timestamp       time.Time `gorm:"not null;index"`
    SourceIP        string    `gorm:"size:45"`
    SourcePort      *int
    DestinationIP   string    `gorm:"size:45"`
    DestinationPort *int
    Protocol        string    `gorm:"size:20"`
    Action          string    `gorm:"size:20"`
    Status          string    `gorm:"size:20"`
    UserID          *uint
    DeviceID        string    `gorm:"size:100"`
    LogSourceID     uint      `gorm:"not null"`
    Severity        string    `gorm:"not null;size:20;index"`
    Category        string    `gorm:"not null;size:30;index"`
    Message         string    `gorm:"not null;type:text"`
    RawData         string    `gorm:"type:text"`
    CreatedAt       int64     `gorm:"autoCreateTime"`
}

// TableName specifies the database table name
func (dbSecurityEvent) TableName() string {
    return "security_events"
}

// toDomain converts a database model to a domain model
func (e *dbSecurityEvent) toDomain() domain.SecurityEvent {
    return domain.SecurityEvent{
        ID:              e.ID,
        Timestamp:       e.Timestamp,
        SourceIP:        e.SourceIP,
        SourcePort:      e.SourcePort,
        DestinationIP:   e.DestinationIP,
        DestinationPort: e.DestinationPort,
        Protocol:        e.Protocol,
        Action:          e.Action,
        Status:          e.Status,
        UserID:          e.UserID,
        DeviceID:        e.DeviceID,
        LogSourceID:     e.LogSourceID,
        Severity:        domain.EventSeverity(e.Severity),
        Category:        domain.EventCategory(e.Category),
        Message:         e.Message,
        RawData:         e.RawData,
        CreatedAt:       timeFromTimestamp(e.CreatedAt),
    }
}

// fromDomain converts a domain model to a database model
func (e *dbSecurityEvent) fromDomain(event domain.SecurityEvent) {
    e.ID = event.ID
    e.Timestamp = event.Timestamp
    e.SourceIP = event.SourceIP
    e.SourcePort = event.SourcePort
    e.DestinationIP = event.DestinationIP
    e.DestinationPort = event.DestinationPort
    e.Protocol = event.Protocol
    e.Action = event.Action
    e.Status = event.Status
    e.UserID = event.UserID
    e.DeviceID = event.DeviceID
    e.LogSourceID = event.LogSourceID
    e.Severity = string(event.Severity)
    e.Category = string(event.Category)
    e.Message = event.Message
    e.RawData = event.RawData
    // CreatedAt is set by Gorm automatically
}

// FindSecurityEvents implements SecurityEventRepository.FindSecurityEvents
func (r *GormSecurityEventRepository) FindSecurityEvents(ctx context.Context, query dto.SecurityEventQuery) ([]domain.SecurityEvent, int64, error) {
    // Build query
    dbQuery := r.db.WithContext(ctx).Model(&dbSecurityEvent{})

    // Apply filters
    if query.Severity != "" {
        dbQuery = dbQuery.Where("severity = ?", query.Severity)
    }
    if query.Category != "" {
        dbQuery = dbQuery.Where("category = ?", query.Category)
    }
    if query.SourceIP != "" {
        dbQuery = dbQuery.Where("source_ip = ?", query.SourceIP)
    }
    if query.DestinationIP != "" {
        dbQuery = dbQuery.Where("destination_ip = ?", query.DestinationIP)
    }
    if query.Protocol != "" {
        dbQuery = dbQuery.Where("protocol = ?", query.Protocol)
    }
    if query.Action != "" {
        dbQuery = dbQuery.Where("action = ?", query.Action)
    }
    if query.Status != "" {
        dbQuery = dbQuery.Where("status = ?", query.Status)
    }
    if query.LogSourceID != nil {
        dbQuery = dbQuery.Where("log_source_id = ?", *query.LogSourceID)
    }
    if query.DeviceID != "" {
        dbQuery = dbQuery.Where("device_id = ?", query.DeviceID)
    }

    // Apply date filters if provided
    if query.FromDate != "" {
        fromDate, _ := time.Parse("2006-01-02", query.FromDate)
        dbQuery = dbQuery.Where("timestamp >= ?", fromDate)
    }
    if query.ToDate != "" {
        toDate, _ := time.Parse("2006-01-02", query.ToDate)
        // Add a day to include the entire day
        toDate = toDate.Add(24 * time.Hour)
        dbQuery = dbQuery.Where("timestamp < ?", toDate)
    }

    // Apply search if provided
    if query.Search != "" {
        searchTerm := "%" + strings.ToLower(query.Search) + "%"
        dbQuery = dbQuery.Where("LOWER(message) LIKE ?", searchTerm)
    }

    // Count total before pagination
    var total int64
    if err := dbQuery.Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("count security events: %w", err)
    }

    // Apply pagination and ordering
    offset := (query.Page - 1) * query.PageSize
    dbQuery = dbQuery.Offset(offset).Limit(query.PageSize).Order("timestamp DESC")

    // Execute query
    var dbEvents []dbSecurityEvent
    if err := dbQuery.Find(&dbEvents).Error; err != nil {
        return nil, 0, fmt.Errorf("find security events: %w", err)
    }

    // Convert to domain models
    events := make([]domain.SecurityEvent, len(dbEvents))
    for i, dbEvent := range dbEvents {
        events[i] = dbEvent.toDomain()
    }

    return events, total, nil
}

func (r *GormSecurityEventRepository) GetSecurityEventByID(ctx context.Context, id uint) (*domain.SecurityEvent, error) {
	var dbEvent dbSecurityEvent
	if err := r.db.WithContext(ctx).First(&dbEvent, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get security event by ID: %w", err)
	}

	event := dbEvent.toDomain()
	return &event, nil
}


func (r *GormSecurityEventRepository) CreateSecurityEvent(ctx context.Context, event *domain.SecurityEvent) error {
	var dbEvent dbSecurityEvent
	dbEvent.fromDomain(*event)

	err := r.db.WithContext(ctx).Create(&dbEvent).Error
	if err != nil {
		return fmt.Errorf("create security event: %w", err)
	}

	// update the event ID after creation
	event.ID = dbEvent.ID
	event.CreatedAt = timeFromTimestamp(dbEvent.CreatedAt)

	return nil
}

func (r *GormSecurityEventRepository) BatchCreateSecurityEvents(ctx context.Context, events []*domain.SecurityEvent) error {
	if len(events) == 0 {
		return nil
	}

	dbEvents := make([]dbSecurityEvent, len(events))
	for i, event := range events {
		dbEvents[i].fromDomain(*event)
	}

	err := r.db.WithContext(ctx).Create(&dbEvents).Error
	if err != nil {
		return fmt.Errorf("batch create security events: %w", err)
	}

	// update the event IDs after creation
	for i, dbEvent := range dbEvents {
		events[i].ID = dbEvent.ID
		events[i].CreatedAt = timeFromTimestamp(dbEvent.CreatedAt)
	}

	return nil
}

func (r *GormSecurityEventRepository) DeleteSecurityEvent(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&dbSecurityEvent{}, id)

	if result.Error != nil {
		return fmt.Errorf("delete security event: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

