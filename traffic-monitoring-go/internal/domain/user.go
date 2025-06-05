package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserRole defines the role a user can have
type UserRole string

const (
	AdminRole    UserRole = "admin"
	AnalystRole  UserRole = "analyst"
	ViewerRole   UserRole = "viewer"
	OperatorRole UserRole = "operator"
)

// ValidUserRoles returns all valid user role values
func ValidUserRoles() []UserRole {
	return []UserRole{
		AdminRole,
		AnalystRole,
		ViewerRole,
		OperatorRole,
	}
}

// User represents a user of the system
type User struct {
	ID             uint
	Email          string
	HashedPassword string
	Role           UserRole
	FirstName      string
	LastName       string
	IsActive       bool
	LastLoginAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (u *User) IsValid() bool {
	return u.Email != "" && u.HashedPassword != "" && u.Role != ""
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	return err == nil
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.HashedPassword = string(hashedPassword)
	return nil
}

func (u *User) FullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Email
	}
	return u.FirstName + " " + u.LastName
}

func (u *User) HasRole(role UserRole) bool {
	return u.Role == role
}

func (u *User) IsAdmin() bool {
	return u.Role == AdminRole
}

func (u *User) CanManageUsers() bool {
	return u.Role == AdminRole
}

func (u *User) CanManageRules() bool {
	return u.Role == AdminRole || u.Role == AnalystRole
}

func (u *User) CanManageAlerts() bool {
	return u.Role == AdminRole || u.Role == AnalystRole || u.Role == OperatorRole
}

func (u *User) CanViewEvents() bool {
	return true // all authenticated users can view events
}
