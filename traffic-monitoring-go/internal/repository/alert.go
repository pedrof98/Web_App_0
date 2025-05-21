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

// this type implements AlertRepository using Gorm
type GormAlertRepository struct {
	db *gorm.DB
}

// this function creates a new GormAlertRepository
func NewGormAlertRepository(db *gorm.DB) *GormAlertRepository {
	return &GormAlertRepository{db: db}
}

// dbAlert is the database model for alerts
type dbAlert struct {
	ID			uint		`gorm:"primaryKey"`
	RuleID			uint		`gorm:"not null"`
	SecurityEventID		uint		`gorm:"not null"`
	Timestamp		time.Time	`gorm:"not null"`
	Severity		string		`gorm:"not null"`
	Status			string		`gorm:"not null"`
	AssignedTo		*uint		
	Resolution		string
	CreatedAt		int64		`gorm:"autoCreateTime"`
	UpdatedAt		int64		`gorm:"autoUpdateTime"`
}


// TableName specifies the database table name
func (dbAlert) TableName() string {
	return "alerts"
}

// toDomain converts a database model to a domain model
func (a *dbAlert) toDomain() domain.Alert {
	return domain.Alert{
		ID:			a.ID,
		RuleID:			a.RuleID,
		SecurityEventID:	a.SecurityEventID,
		Timestamp:		a.Timestamp,
		Severity:        domain.EventSeverity(a.Severity),
		Status:          domain.AlertStatus(a.Status),
		AssignedTo:      a.AssignedTo,
		Resolution:      a.Resolution,
		CreatedAt:       timeFromTimestamp(a.CreatedAt),
		UpdatedAt:       timeFromTimestamp(a.UpdatedAt),
	}
}

// fromDomain converts a domain model to a database model
func (a *dbAlert) fromDomain(alert domain.Alert) {
	a.ID = alert.ID
	a.RuleID = alert.RuleID
	a.SecurityEventID = alert.SecurityEventID
	a.Timestamp = alert.Timestamp
	a.Severity = string(alert.Severity)
	a.Status = string(alert.Status)
	a.AssignedTo = alert.AssignedTo
	a.Resolution = alert.Resolution
	// CreatedAt and UpdatedAt are set by Gorm automatically
}

// Find alerts implements AlertRepository.FindAlerts
func (r *GormAlertRepository) FindAlerts(ctx context.Context, query dto.AlertQuery) ([]domain.Alert, int64, error) {
	// build query
	dbQuery := r.db.WithContext(ctx).Model(&dbAlert{})

	// apply filters
	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}
	if query.Severity != "" {
		dbQuery = dbQuery.Where("severity = ?", query.Severity)
	}
	if query.AssignedTo != nil {
		dbQuery = dbQuery.Where("assigned_to = ?", *query.AssignedTo)
	}
	if query.RuleID != nil {
		dbQuery = dbQuery.Where("rule_id = ?", *query.RuleID)
	}

	// apply date filters if provided
	if query.FromDate != "" {
		toDate, _ := time.Parse("2006-01-02", query.toDate)
		dbQuery = dbQuery.Where("timestamp < ?", toDate)
	}
	if query.ToDate != "" {
		toDate, _ := time.Parse("2006-01-02", query.ToDate)
		// add a day to include the entire day
		toDate = toDate.Add(24 * time.Hour)
		dbQuery = dbQuery.Where("LOWER(resolution) LIKE ?", searchTerm)
	}
	
	// count total before pagination
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count alerts: %w", err)
	}

	// apply pagination and ordering
	offset := (query.Page - 1) * query.PageSize
	dbQuery = dbQuery.Offset(offset).Limit(query.Pagesize).Order("timestamp DESC")

	// execute query
	var dbAlerts []dbAlert
	if err := dbQuery.Find(&dbAlerts).Error; err != nil {
		return nil, 0, fmt.Errorf("find alerts: %w", err)
	}

	// convert to domain models
	alerts := make([]domain.Alert, len(dbAlerts))
	for i, dbAlert := range dbAlerts {
		alerts[i] = dbAlert.toDomain()
	}

	// optional: Load related rules
	if len(alerts) > 0 {
		ruleIDs := make([]uint, len(alerts))
		for i, alert := range alerts {
			ruleIDs[i] = alert.RuleID
		}

		var dbRules []dbRule
		if err := r.db.WithContext(ctx).Where("id IN ?", ruleIDs).Find(&dbRules).Error; err != nil {
			return alerts, total, fmt.Errof("load rules: %w", err)
		}

		ruleMap := make(map[uint]domain.Rule)
		for _, dbRule := range dbRules {
			rule := dbRule.toDomain()
			ruleMap[rule.ID] = rule
		}

		for i := range alerts {
			if rule, ok := ruleMap[alerts[i].RuleID]; ok {
				ruleCopy := rule
				alerts[i].Rule = &ruleCopy
			}
		}
	}

	return alerts, total, nil
}

// GetAlertByID implements AlertRepository.GetAlertByID
func (r *GormAlertRepository) GetAlertByID(ctx context.Context, id uint) (*domain.Alert, error) {
	var dbAlert dbAlert
	err := r.db.WithContext(ctx).First(&dbAlert, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get alert by id: %w", err)
	}

	alert := dbAlert.toDomain()

	// load related rule
	var dbRule dbRule
	if err := r.db.WithContext(ctx).First(&dbRule, alert.RuleID).Error; err == nil {
		rule := dbRule.toDomain()
		alert.Rule = &rule
	}

	return &alert, nil
}


// CreateAlert implements AlertRepository.CreateAlert
func (r *GormAlertRepository) CreateAlert(ctx context.Context, alert *domain.Alert) error {
	var dbAlert dbAlert
	dbAlert.fromDomain(*alert)

	err := r.db.WithContext(ctx).Create(&dbAlert).Error
	if err != nil {
		return fmt.errorf("create alert: %w", err)
	}

	// update the alert ID after creation
	alert.ID = dbAlert.ID
	alert.CreatedAt = timeFromTimestamp(dbAlert.CreatedAt)
	alert.UpdatedAt = timeFromTimestamp(dbAlert.UpdatedAt)

	return nil
}


// UpdateAlert implements AlertRepository.UpdateAlert
func (r *GormAlertrepository) UpdateAlert(ctx context.Context, alert *domain.Alert) error {
	var dbAlert dbAlert
	dbAlert.fromDomain(*alert)

	err := r.db.WithContext(ctx).Save(&dbAlert).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("update alert: %w", err)
	}

	// update the alert timestamps after update
	alert.UpdatedAt = timeFromTimestamp(dbAlert.UpdatedAt)
	return nil
}

// DeleteAlert implements AlertRepository.DeleteAlert
func (r *GormAlertRepository) DeleteAlert(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&dbAlert{}, id)

	if result.Error != nil {
		return fmt.Errorf("delete alert: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errNotFound
	}
	return nil
}

// CountAlertsByRuleID implements AlertRepository.CountAlertsByRuleID
func (r *GormAlertRepository) CountAlertsByRuleID(ctx context.Context, ruleID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&dbAlert{}).Where("rule_id = ?", ruleID).Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("count alerts by rule: %w", err)
	}

	return count, nil
}
			
	
