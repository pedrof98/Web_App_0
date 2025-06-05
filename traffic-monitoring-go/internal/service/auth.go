package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"traffic-monitoring-go/internal/domain"
	"traffic-monitoring-go/internal/dto"
	"traffic-monitoring-go/internal/pkg/auth"
	"traffic-monitoring-go/internal/repository"
)

// AuthServiceImpl implements the AuthService interface
type AuthServiceImpl struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewAuthService creates a new AuthServiceImpl
func NewAuthService(userRepo repository.UserRepository, jwtManager *auth.JWTManager) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Login authenticates a user and returns a JWT token
func (s *AuthServiceImpl) Login(ctx context.Context, request *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.GetUserByEmail(ctx, request.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: invalid email or password", ErrBadRequest)
		}
		return nil, WrapError(err)
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("%w: account is deactivated", ErrForbidden)
	}

	// Verify password
	if !user.CheckPassword(request.Password) {
		return nil, fmt.Errorf("%w: invalid email or password", ErrBadRequest)
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate token", ErrInternal)
	}

	// Update last login time
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail the login
		fmt.Printf("Failed to update last login for user %d: %v\n", user.ID, err)
	}

	// Calculate token expiration
	expiresIn := int64(24 * time.Hour / time.Second) // 24 hours in seconds

	return &dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        dto.UserToResponse(user),
	}, nil
}

// Register creates a new user account
func (s *AuthServiceImpl) Register(ctx context.Context, request *dto.RegisterRequest) (*domain.User, error) {
	// Check if user already exists
	_, err := s.userRepo.GetUserByEmail(ctx, request.Email)
	if err == nil {
		return nil, fmt.Errorf("%w: user with email %s already exists", ErrConflict, request.Email)
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, WrapError(err)
	}

	// Convert DTO to domain model
	user := request.ToDomain()

	// Set password
	if err := user.SetPassword(request.Password); err != nil {
		return nil, fmt.Errorf("%w: failed to hash password", ErrInternal)
	}

	// Validate the user
	if !user.IsValid() {
		return nil, ErrBadRequest
	}

	// Save the user
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, WrapError(err)
	}

	return user, nil
}

// RefreshToken generates a new access token
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, tokenString string) (*dto.LoginResponse, error) {
	// Validate current token
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}

	// Get user from database to ensure they still exist and are active
	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("%w: user not found", ErrBadRequest)
		}
		return nil, WrapError(err)
	}

	if !user.IsActive {
		return nil, fmt.Errorf("%w: account is deactivated", ErrForbidden)
	}

	// Generate new token
	newToken, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to generate token", ErrInternal)
	}

	// Calculate token expiration
	expiresIn := int64(24 * time.Hour / time.Second) // 24 hours in seconds

	return &dto.LoginResponse{
		AccessToken: newToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		User:        dto.UserToResponse(user),
	}, nil
}

// GetCurrentUser retrieves the current user from context
func (s *AuthServiceImpl) GetCurrentUser(ctx context.Context) (*domain.User, error) {
	userID, ok := ctx.Value("user_id").(uint)
	if !ok {
		return nil, fmt.Errorf("%w: user not found in context", ErrBadRequest)
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, WrapError(err)
	}

	return user, nil
}

// ChangePassword changes a user's password
func (s *AuthServiceImpl) ChangePassword(ctx context.Context, userID uint, request *dto.ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return WrapError(err)
	}

	// Verify current password
	if !user.CheckPassword(request.CurrentPassword) {
		return fmt.Errorf("%w: current password is incorrect", ErrBadRequest)
	}

	// Set new password
	if err := user.SetPassword(request.NewPassword); err != nil {
		return fmt.Errorf("%w: failed to hash new password", ErrInternal)
	}

	// Update user in database
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// UpdateProfile updates a user's profile information
func (s *AuthServiceImpl) UpdateProfile(ctx context.Context, userID uint, request *dto.UpdateProfileRequest) (*domain.User, error) {
	// Get user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, WrapError(err)
	}

	// Apply updates
	request.ApplyToUser(user)

	// Validate the updated user
	if !user.IsValid() {
		return nil, ErrBadRequest
	}

	// Update user in database
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, WrapError(err)
	}

	return user, nil
}

// UserServiceImpl implements the UserService interface
type UserServiceImpl struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserServiceImpl
func NewUserService(userRepo repository.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo: userRepo,
	}
}

// ListUsers retrieves users based on query parameters
func (s *UserServiceImpl) ListUsers(ctx context.Context, query dto.UserQuery) ([]domain.User, *dto.MetaInfo, error) {
	users, total, err := s.userRepo.FindUsers(ctx, query)
	if err != nil {
		return nil, nil, WrapError(err)
	}

	meta := dto.CalculatePagination(query.Page, query.PageSize, total)
	return users, meta, nil
}

// GetUser retrieves a single user by ID
func (s *UserServiceImpl) GetUser(ctx context.Context, id uint) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, WrapError(err)
	}
	return user, nil
}

// CreateUser creates a new user (admin only)
func (s *UserServiceImpl) CreateUser(ctx context.Context, request *dto.RegisterRequest) (*domain.User, error) {
	// Check if user already exists
	_, err := s.userRepo.GetUserByEmail(ctx, request.Email)
	if err == nil {
		return nil, fmt.Errorf("%w: user with email %s already exists", ErrConflict, request.Email)
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, WrapError(err)
	}

	// Convert DTO to domain model
	user := request.ToDomain()

	// Set password
	if err := user.SetPassword(request.Password); err != nil {
		return nil, fmt.Errorf("%w: failed to hash password", ErrInternal)
	}

	// Validate the user
	if !user.IsValid() {
		return nil, ErrBadRequest
	}

	// Save the user
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, WrapError(err)
	}

	return user, nil
}

// UpdateUser updates an existing user (admin only)
func (s *UserServiceImpl) UpdateUser(ctx context.Context, id uint, request *dto.UpdateProfileRequest) (*domain.User, error) {
	// Get user
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, WrapError(err)
	}

	// Apply updates
	request.ApplyToUser(user)

	// Validate the updated user
	if !user.IsValid() {
		return nil, ErrBadRequest
	}

	// Update user in database
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, WrapError(err)
	}

	return user, nil
}

// DeleteUser removes a user by ID (admin only)
func (s *UserServiceImpl) DeleteUser(ctx context.Context, id uint) error {
	// Check if user exists
	_, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return WrapError(err)
	}

	// Delete the user
	err = s.userRepo.DeleteUser(ctx, id)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// ActivateUser activates a user account
func (s *UserServiceImpl) ActivateUser(ctx context.Context, id uint) error {
	// Get user
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return WrapError(err)
	}

	// Update active status
	user.IsActive = true

	// Update user in database
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return WrapError(err)
	}

	return nil
}

// DeactivateUser deactivates a user account
func (s *UserServiceImpl) DeactivateUser(ctx context.Context, id uint) error {
	// Get user
	user, err := s.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return WrapError(err)
	}

	// Update active status
	user.IsActive = false

	// Update user in database
	err = s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return WrapError(err)
	}

	return nil
}
