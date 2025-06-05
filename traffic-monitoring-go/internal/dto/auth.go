package dto

import (
	"time"
	"traffic-monitoring-go/internal/domain"

	"github.com/gin-gonic/gin"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"`
	User        UserResponse `json:"user"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RegisterRequest represents the user registration request
type RegisterRequest struct {
	Email     string          `json:"email" binding:"required,email"`
	Password  string          `json:"password" binding:"required,min=6"`
	FirstName string          `json:"first_name" binding:"required,min=2,max=50"`
	LastName  string          `json:"last_name" binding:"required,min=2,max=50"`
	Role      domain.UserRole `json:"role" binding:"omitempty,oneof=admin analyst viewer operator"`
}

// ChangePasswordRequest represents the change password request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// UpdateProfileRequest represents the update profile request
type UpdateProfileRequest struct {
	FirstName *string `json:"first_name" binding:"omitempty,min=2,max=50"`
	LastName  *string `json:"last_name" binding:"omitempty,min=2,max=50"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID          uint            `json:"id"`
	Email       string          `json:"email"`
	Role        domain.UserRole `json:"role"`
	FirstName   string          `json:"first_name"`
	LastName    string          `json:"last_name"`
	FullName    string          `json:"full_name"`
	IsActive    bool            `json:"is_active"`
	LastLoginAt *time.Time      `json:"last_login_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// UserQuery represents query parameters for filtering users
type UserQuery struct {
	PaginationQuery
	Role     string `form:"role" binding:"omitempty,oneof=admin analyst viewer operator"`
	IsActive *bool  `form:"is_active" binding:"omitempty"`
	Search   string `form:"search" binding:"omitempty"`
}

// ParseUserQuery parses query parameters from the request
func ParseUserQuery(c *gin.Context) (UserQuery, error) {
	var q UserQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		return q, err
	}

	// Set defaults if not provided
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 50
	}

	return q, nil
}

// ToDomain converts the register request to a domain user
func (r *RegisterRequest) ToDomain() *domain.User {
	role := r.Role
	if role == "" {
		role = domain.ViewerRole // Default role
	}

	user := &domain.User{
		Email:     r.Email,
		Role:      role,
		FirstName: r.FirstName,
		LastName:  r.LastName,
		IsActive:  true,
	}

	// Password will be set separately using SetPassword method
	return user
}

// ApplyToUser applies update request fields to a domain user
func (r *UpdateProfileRequest) ApplyToUser(user *domain.User) {
	if r.FirstName != nil {
		user.FirstName = *r.FirstName
	}
	if r.LastName != nil {
		user.LastName = *r.LastName
	}
}

// UserToResponse converts a domain user to a response DTO
func UserToResponse(user *domain.User) UserResponse {
	return UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		Role:        user.Role,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		FullName:    user.FullName(),
		IsActive:    user.IsActive,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

// UsersToResponses converts a slice of domain users to response DTOs
func UsersToResponses(users []domain.User) []UserResponse {
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = UserToResponse(&user)
	}
	return responses
}
