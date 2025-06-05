package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"traffic-monitoring-go/internal/domain"
	"traffic-monitoring-go/internal/dto"

	"gorm.io/gorm"
)

// GormUserRepository implements UserRepository using Gorm
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new GormUserRepository
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// dbUser is the database model for users
type dbUser struct {
	ID             uint   `gorm:"primaryKey"`
	Email          string `gorm:"not null;unique;size:255"`
	HashedPassword string `gorm:"not null;size:255"`
	Role           string `gorm:"not null;size:20"`
	FirstName      string `gorm:"size:50"`
	LastName       string `gorm:"size:50"`
	IsActive       bool   `gorm:"not null;default:true"`
	LastLoginAt    *time.Time
	CreatedAt      int64 `gorm:"autoCreateTime"`
	UpdatedAt      int64 `gorm:"autoUpdateTime"`
}

// TableName specifies the database table name
func (dbUser) TableName() string {
	return "users"
}

// toDomain converts a database model to a domain model
func (u *dbUser) toDomain() domain.User {
	return domain.User{
		ID:             u.ID,
		Email:          u.Email,
		HashedPassword: u.HashedPassword,
		Role:           domain.UserRole(u.Role),
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		IsActive:       u.IsActive,
		LastLoginAt:    u.LastLoginAt,
		CreatedAt:      timeFromTimestamp(u.CreatedAt),
		UpdatedAt:      timeFromTimestamp(u.UpdatedAt),
	}
}

// fromDomain converts a domain model to a database model
func (u *dbUser) fromDomain(user domain.User) {
	u.ID = user.ID
	u.Email = user.Email
	u.HashedPassword = user.HashedPassword
	u.Role = string(user.Role)
	u.FirstName = user.FirstName
	u.LastName = user.LastName
	u.IsActive = user.IsActive
	u.LastLoginAt = user.LastLoginAt
	// CreatedAt and UpdatedAt are set by Gorm automatically
}

// FindUsers implements UserRepository.FindUsers
func (r *GormUserRepository) FindUsers(ctx context.Context, query dto.UserQuery) ([]domain.User, int64, error) {
	// Build query
	dbQuery := r.db.WithContext(ctx).Model(&dbUser{})

	// Apply filters
	if query.Role != "" {
		dbQuery = dbQuery.Where("role = ?", query.Role)
	}
	if query.IsActive != nil {
		dbQuery = dbQuery.Where("is_active = ?", *query.IsActive)
	}
	if query.Search != "" {
		searchTerm := "%" + strings.ToLower(query.Search) + "%"
		dbQuery = dbQuery.Where(
			"LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?",
			searchTerm, searchTerm, searchTerm,
		)
	}

	// Count total before pagination
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	// Apply pagination and ordering
	offset := (query.Page - 1) * query.PageSize
	dbQuery = dbQuery.Offset(offset).Limit(query.PageSize).Order("email ASC")

	// Execute query
	var dbUsers []dbUser
	if err := dbQuery.Find(&dbUsers).Error; err != nil {
		return nil, 0, fmt.Errorf("find users: %w", err)
	}

	// Convert to domain models
	users := make([]domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = dbUser.toDomain()
	}

	return users, total, nil
}

// GetUserByID implements UserRepository.GetUserByID
func (r *GormUserRepository) GetUserByID(ctx context.Context, id uint) (*domain.User, error) {
	var dbUser dbUser
	err := r.db.WithContext(ctx).First(&dbUser, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	user := dbUser.toDomain()
	return &user, nil
}

// GetUserByEmail implements UserRepository.GetUserByEmail
func (r *GormUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var dbUser dbUser
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&dbUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	user := dbUser.toDomain()
	return &user, nil
}

// CreateUser implements UserRepository.CreateUser
func (r *GormUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	var dbUser dbUser
	dbUser.fromDomain(*user)

	err := r.db.WithContext(ctx).Create(&dbUser).Error
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return ErrDuplicate
		}
		return fmt.Errorf("create user: %w", err)
	}

	// Update the user ID after creation
	user.ID = dbUser.ID
	user.CreatedAt = timeFromTimestamp(dbUser.CreatedAt)
	user.UpdatedAt = timeFromTimestamp(dbUser.UpdatedAt)

	return nil
}

// UpdateUser implements UserRepository.UpdateUser
func (r *GormUserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	var dbUser dbUser
	dbUser.fromDomain(*user)

	err := r.db.WithContext(ctx).Save(&dbUser).Error
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return ErrDuplicate
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("update user: %w", err)
	}

	// Update the user timestamps after update
	user.UpdatedAt = timeFromTimestamp(dbUser.UpdatedAt)
	return nil
}

// DeleteUser implements UserRepository.DeleteUser
func (r *GormUserRepository) DeleteUser(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&dbUser{}, id)

	if result.Error != nil {
		return fmt.Errorf("delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateLastLogin implements UserRepository.UpdateLastLogin
func (r *GormUserRepository) UpdateLastLogin(ctx context.Context, userID uint) error {
	now := time.Now()
	err := r.db.WithContext(ctx).Model(&dbUser{}).
		Where("id = ?", userID).
		Update("last_login_at", now).Error

	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}

	return nil
}
