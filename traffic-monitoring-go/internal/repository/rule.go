package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"traffic-monitoring-go/internal/domain"
	"traffic-monitoring-go/internal/dto"

)

// GormRuleRepository implements RuleRepository usin Gorm
type GormRuleRepository struct {
	db *gorm.DB
}

// NewGormRuleRepository creates a new ""
func NewGormRuleRepository(db *gorm.DB) *GormRuleRepository {
	return &GormRuleRepository{db: db}
}

// dbRule is the database model for rules
type dbRule struct {
	ID		uint		`gorm:"primaryKey"`
	Name		string		`gorm:"not null;unique"`
	Description	string
	Condition	string		`gorm:"not null"`
	Severity	string		`gorm:"not null"`
	Category	string		`gorm:"not null"`
	Status		string		`gorm:"not null"`
	CreatedBy	uint		`gorm:"not null"`
	CreatedAt	int64		`gorm:"autoCreateTime"`
	UpdatedAt	int64		`gorm:"autoUpdateTime"`
}

// TableName specifies the database table name
func (dbRule) TableName() string {
	return "rules"
}


// toDomain converts a database model to a domain model
func (r *dbRule) toDomain() domain.Rule {
	return domain.Rule{
		ID:		r.ID,
		Name:		r.Name,
		Description:	r.Description,
		Condition:	r.Condition,
		Severity:	domain.EventSeverity(r.Severity),
		Category:	domain.EventCategory(r.Category).
		Status:		domain.RuleStatus(r.Status),
		CreatedBy:	r.CreatedBy,
		CreatedAt:	timeFromTimestamp(r.CreatedAt),
		UpdatedAt:	timeFromTimestamp(r.UpdatedAt),
	}
}

// fromDomain converts a domain model to a database model
func (r *dbRule) fromDomain(rule domain.Rule) {
	r.ID = rule.ID
	r.Name = rule.Name
	r.Description = rule.Description
	r.Condition = rule.Condition
	r.Severity = string(rule.Severity)
	r.Category = string(rule.Category)
	r.Status = string(rule.Status)
	r.CreatedBy = rule.CreatedBy
	// CreatedAt and UpdatedAt are set by Gorm automatically
}

// Findrules implements Rulerepository.Findrules
func (r *GormRuleRepository) FindRules(ctx context.Context, query dto.RuleQuery) ([]domain.Rule, int64, error) {
	// build query
	dbQuery := r.db.WithContext(ctx).Model(&dbERule{})

	// apply filters
	if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}
	if query.Category != "" {
		dbQuery = dbQuery.Where("category = ?", query.Category)
	}
	if query.Search != "" {
		searchTerm := "%" + strings.ToLower(query.Search) + "%"
		dbQuery = dbQuery.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	// count total before pagination
	var total int64if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count rules: %w", err)
	}

	// apply pagination and ordering
	offset := (query.Page - 1) * query.PageSize
	dbQuery = dbQuery.Offset(offset).Limit(query.PageSize).Order("name ASC")

	// execute query
	var dbRules []dbRule
	if err := dbQuery.Find(&dbRules).Error; err != nil {
		return nil, 0, fmt.Errorf("find rules: %w", err)
	}

	// convert to domain models
	rules := make([]domain.Rule, len(dbRules))
	for i, dbRule := range dbRules {
		rules[i] = dbRule.toDomain()
	}

	return rules, total, nil
}

// GetRuleByID implements Rulerepository.GetRuleByID
func (r *GormRuleRepository) GetRuleByID(ctx context.Context, id uint) (*domain.Rule, error) {
	var dbRule dbRule
	err := r.db.WithContext(ctx).First(&dbRule, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrrecordNotFound) {
			return nul, ErrNotFound
		}
		return nil, fmt.Errorf("get rule by id: %w", err)
	}

	rule := dbRule.toDomain()
	return &rule, nil
}

// CreateRule implements Rulerepository.CreateRule
func (r *GormRuleRepository) CreateRule(ctx context.Context, rule *domain.Rule) error {
	var dbRule dbRule
	dbRule.fromDomain(*rule)

	err := r.db.WithContext(ctx).Create(&dbRule).Error
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return ErrDuplicate
		}
		return fmt.Errorf("create rule: %w", err)
	}

	// update the rule ID after creation
	rule.ID = dbRule.ID
	rule.CreatedAt = timeFromTimestamp(dbRule.CreatedAt)
	rule.UpdatedAt = timeFromTimestamp(dbRule.UpdatedAt)

	return nil
}

// UpdateRule implements Rulerepository.UpdateRule
func (r *GormRuleRepository) UpdateRule(ctx context.Context, rule *domain.Rule) error {
	var dbRule dbRule
	dbRule.fromDomain(*rule)

	err := r.db.WithContext(ctx).Save(&dbRule).Error
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return ErrDuplicate
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("update rule: %w", err)
	}

	// update the rule timestamps after update
	rule.UpdatedAt = timeFromTimestamp(dbRule.UpdatedAt)
	return nil
}


// DeleteRule implements Rulerepository.DeleteRule
func (r *GormRuleRepository) DeleteRule(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&dbRule{}, id)
	
	if result.Error != nil {
		return fmt.Errorf("delete rule: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// CountAlertsByRuleID implements Rulerepository.CountAlertsByRuleID
func (r *GormRuleRepository) CountAlertsByRuleID(ctx context.Context, ruleID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&struct {
		ID uint `gorm:"primaryKey"`
		RuleID uint `gorm:"not null"`
	}{}).Where("rule_id = ?", ruleID).Count(&count).Error

	if err != nil {
		return 0, fmt.Errof("count alerts by rule: %w", err)
	}

	return count, nil
}


