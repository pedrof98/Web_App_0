package middleware

import (
	"context"
	"net/http"
	"strings"

	"traffic-monitoring-go/internal/domain"
	"traffic-monitoring-go/internal/dto"
	"traffic-monitoring-go/internal/pkg/auth"

	"github.com/gin-gonic/gin"
)

const (
	// AuthorizationHeader is the header key for authorization
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for bearer tokens
	BearerPrefix = "Bearer "
	// UserIDKey is the context key for user ID
	UserIDKey = "user_id"
	// UserRoleKey is the context key for user role
	UserRoleKey = "user_role"
	// UserEmailKey is the context key for user email
	UserEmailKey = "user_email"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			respondWithError(c, "Authorization header is required", "MISSING_TOKEN")
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			var code string
			switch err {
			case auth.ErrExpiredToken:
				code = "EXPIRED_TOKEN"
			case auth.ErrInvalidToken:
				code = "INVALID_TOKEN"
			case auth.ErrTokenClaims:
				code = "INVALID_CLAIMS"
			default:
				code = "TOKEN_ERROR"
			}
			respondWithError(c, err.Error(), code)
			c.Abort()
			return
		}

		// Set user information in context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)
		c.Set(UserEmailKey, claims.Email)

		// Add to request context as well for service layer
		ctx := context.WithValue(c.Request.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequireRole creates a middleware that requires specific roles
func RequireRole(roles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(UserRoleKey)
		if !exists {
			respondWithError(c, "User role not found in context", "MISSING_ROLE")
			c.Abort()
			return
		}

		role, ok := userRole.(domain.UserRole)
		if !ok {
			respondWithError(c, "Invalid user role in context", "INVALID_ROLE")
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, requiredRole := range roles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			respondWithError(c, "Insufficient permissions", "INSUFFICIENT_PERMISSIONS")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin creates a middleware that requires admin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(domain.AdminRole)
}

// RequireAdminOrAnalyst creates a middleware that requires admin or analyst role
func RequireAdminOrAnalyst() gin.HandlerFunc {
	return RequireRole(domain.AdminRole, domain.AnalystRole)
}

// RequireAdminOrAnalystOrOperator creates a middleware that requires admin, analyst, or operator role
func RequireAdminOrAnalystOrOperator() gin.HandlerFunc {
	return RequireRole(domain.AdminRole, domain.AnalystRole, domain.OperatorRole)
}

// OptionalAuth creates an optional authentication middleware
func OptionalAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Set user information in context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)
		c.Set(UserEmailKey, claims.Email)

		// Add to request context as well for service layer
		ctx := context.WithValue(c.Request.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// extractToken extracts the JWT token from the Authorization header
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader(AuthorizationHeader)
	if authHeader == "" {
		return ""
	}

	if !strings.HasPrefix(authHeader, BearerPrefix) {
		return ""
	}

	return strings.TrimPrefix(authHeader, BearerPrefix)
}

// respondWithError sends an error response
func respondWithError(c *gin.Context, message, code string) {
	response := dto.Error{}
	response.Error.Code = code
	response.Error.Message = message

	c.JSON(http.StatusUnauthorized, response)
}

// GetUserIDFromContext retrieves the user ID from Gin context
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}

// GetUserRoleFromContext retrieves the user role from Gin context
func GetUserRoleFromContext(c *gin.Context) (domain.UserRole, bool) {
	userRole, exists := c.Get(UserRoleKey)
	if !exists {
		return "", false
	}

	role, ok := userRole.(domain.UserRole)
	return role, ok
}

// GetUserEmailFromContext retrieves the user email from Gin context
func GetUserEmailFromContext(c *gin.Context) (string, bool) {
	userEmail, exists := c.Get(UserEmailKey)
	if !exists {
		return "", false
	}

	email, ok := userEmail.(string)
	return email, ok
}
