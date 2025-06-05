package auth

import (
	"errors"
	"fmt"
	"time"

	"traffic-monitoring-go/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
	ErrTokenClaims  = errors.New("invalid token claims")
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID  uint            `json:"user_id"`
	Email   string          `json:"email"`
	Role    domain.UserRole `json:"role"`
	TokenID string          `json:"token_id"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

// GenerateToken generates a new JWT token for a user
func (m *JWTManager) GenerateToken(user *domain.User) (string, error) {
	now := time.Now()

	claims := &Claims{
		UserID:  user.ID,
		Email:   user.Email,
		Role:    user.Role,
		TokenID: fmt.Sprintf("%d_%d", user.ID, now.Unix()),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.tokenDuration)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "v2x-siem",
			Subject:   fmt.Sprintf("user_%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenClaims
	}

	return claims, nil
}

// RefreshToken generates a new token if the current one is close to expiring
func (m *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Check if token is close to expiring (within 15 minutes)
	if time.Until(claims.ExpiresAt.Time) > 15*time.Minute {
		return "", errors.New("token does not need refresh yet")
	}

	// Create new token with updated expiration
	user := &domain.User{
		ID:    claims.UserID,
		Email: claims.Email,
		Role:  claims.Role,
	}

	return m.GenerateToken(user)
}
