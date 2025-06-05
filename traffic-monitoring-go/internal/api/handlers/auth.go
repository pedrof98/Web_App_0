package handlers

import (
	"fmt"
	"strconv"

	"traffic-monitoring-go/internal/api/middleware"
	"traffic-monitoring-go/internal/domain"
	"traffic-monitoring-go/internal/dto"
	"traffic-monitoring-go/internal/pkg/respond"
	"traffic-monitoring-go/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	authService service.AuthService
	userService service.UserService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService service.AuthService, userService service.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

// Login handles POST /api/v1/auth/login
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.Success[dto.LoginResponse]
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	// Parse request body
	var request dto.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	response, err := h.authService.Login(c.Request.Context(), &request)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Send response
	respond.OK(c, response, nil)
}

// Register handles POST /api/v1/auth/register
// @Summary User registration
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param user body dto.RegisterRequest true "User registration details"
// @Success 201 {object} dto.Success[dto.UserResponse]
// @Failure 400 {object} dto.Error
// @Failure 409 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	// Parse request body
	var request dto.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}

	// For public registration, force role to viewer
	request.Role = domain.ViewerRole

	// Call service
	user, err := h.authService.Register(c.Request.Context(), &request)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Transform domain model to response DTO
	response := dto.UserToResponse(user)

	// Send response
	respond.Created(c, response)
}

// RefreshToken handles POST /api/v1/auth/refresh
// @Summary Refresh access token
// @Description Generate a new access token using the current token
// @Tags auth
// @Accept json
// @Produce json
// @Param token body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.Success[dto.LoginResponse]
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Parse request body
	var request dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	response, err := h.authService.RefreshToken(c.Request.Context(), request.RefreshToken)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Send response
	respond.OK(c, response, nil)
}

// GetProfile handles GET /api/v1/auth/profile
// @Summary Get current user profile
// @Description Get the current authenticated user's profile
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Success[dto.UserResponse]
// @Failure 401 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get current user
	user, err := h.authService.GetCurrentUser(c.Request.Context())
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Transform domain model to response DTO
	response := dto.UserToResponse(user)

	// Send response
	respond.OK(c, response, nil)
}

// UpdateProfile handles PUT /api/v1/auth/profile
// @Summary Update current user profile
// @Description Update the current authenticated user's profile
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body dto.UpdateProfileRequest true "Profile update details"
// @Success 200 {object} dto.Success[dto.UserResponse]
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/auth/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		respond.BadRequest(c, fmt.Errorf("user ID not found in context"))
		return
	}

	// Parse request body
	var request dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	user, err := h.authService.UpdateProfile(c.Request.Context(), userID, &request)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Transform domain model to response DTO
	response := dto.UserToResponse(user)

	// Send response
	respond.OK(c, response, nil)
}

// ChangePassword handles POST /api/v1/auth/change-password
// @Summary Change user password
// @Description Change the current authenticated user's password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body dto.ChangePasswordRequest true "Password change details"
// @Success 200 {object} dto.Success[string]
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	// Get user ID from context
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		respond.BadRequest(c, fmt.Errorf("user ID not found in context"))
		return
	}

	// Parse request body
	var request dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	err := h.authService.ChangePassword(c.Request.Context(), userID, &request)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Send response
	respond.OK(c, "Password changed successfully", nil)
}

// UserHandler handles HTTP requests for user management (admin only)
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ListUsers handles GET /api/v1/users
// @Summary List users
// @Description Get a list of users with pagination and filtering (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param role query string false "Filter by role (admin, analyst, viewer, operator)"
// @Param is_active query bool false "Filter by active status"
// @Param search query string false "Search in email, first name, or last name"
// @Success 200 {object} dto.Success[[]dto.UserResponse]
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse query parameters
	query, err := dto.ParseUserQuery(c)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	users, meta, err := h.userService.ListUsers(c.Request.Context(), query)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Transform domain models to response DTOs
	responses := dto.UsersToResponses(users)

	// Send response
	respond.OK(c, responses, meta)
}

// GetUser handles GET /api/v1/users/:id
// @Summary Get a user
// @Description Get a single user by ID (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Success[dto.UserResponse]
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	// Parse user ID from path
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	user, err := h.userService.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Transform domain model to response DTO
	response := dto.UserToResponse(user)

	// Send response
	respond.OK(c, response, nil)
}

// CreateUser handles POST /api/v1/users
// @Summary Create a user
// @Description Create a new user (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body dto.RegisterRequest true "User details"
// @Success 201 {object} dto.Success[dto.UserResponse]
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 409 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	// Parse request body
	var request dto.RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	user, err := h.userService.CreateUser(c.Request.Context(), &request)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Transform domain model to response DTO
	response := dto.UserToResponse(user)

	// Send response
	respond.Created(c, response)
}

// UpdateUser handles PUT /api/v1/users/:id
// @Summary Update a user
// @Description Update an existing user (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param user body dto.UpdateProfileRequest true "User details"
// @Success 200 {object} dto.Success[dto.UserResponse]
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Parse user ID from path
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Parse request body
	var request dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), &request)
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Transform domain model to response DTO
	response := dto.UserToResponse(user)

	// Send response
	respond.OK(c, response, nil)
}

// DeleteUser handles DELETE /api/v1/users/:id
// @Summary Delete a user
// @Description Delete an existing user (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Parse user ID from path
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	err = h.userService.DeleteUser(c.Request.Context(), uint(id))
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Send response
	respond.NoContent(c)
}

// ActivateUser handles POST /api/v1/users/:id/activate
// @Summary Activate a user
// @Description Activate a user account (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Success[string]
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/users/{id}/activate [post]
func (h *UserHandler) ActivateUser(c *gin.Context) {
	// Parse user ID from path
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	err = h.userService.ActivateUser(c.Request.Context(), uint(id))
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Send response
	respond.OK(c, "User activated successfully", nil)
}

// DeactivateUser handles POST /api/v1/users/:id/deactivate
// @Summary Deactivate a user
// @Description Deactivate a user account (admin only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} dto.Success[string]
// @Failure 400 {object} dto.Error
// @Failure 401 {object} dto.Error
// @Failure 403 {object} dto.Error
// @Failure 404 {object} dto.Error
// @Failure 500 {object} dto.Error
// @Router /api/v1/users/{id}/deactivate [post]
func (h *UserHandler) DeactivateUser(c *gin.Context) {
	// Parse user ID from path
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		respond.BadRequest(c, err)
		return
	}

	// Call service
	err = h.userService.DeactivateUser(c.Request.Context(), uint(id))
	if err != nil {
		respond.Error(c, err)
		return
	}

	// Send response
	respond.OK(c, "User deactivated successfully", nil)
}
