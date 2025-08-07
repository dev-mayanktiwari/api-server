// Package auth provides authentication and authorization utilities for the API server.
// It includes JWT token management, password hashing, and authentication middleware.
package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/config"
	sharedErrors "github.com/dev-mayanktiwari/api-server/shared/pkg/errors"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/response"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	config *config.JWTConfig
	logger *logger.Logger
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// UserContext represents authenticated user context
type UserContext struct {
	UserID string
	Email  string
	Role   string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(cfg *config.JWTConfig, log *logger.Logger) *JWTManager {
	if cfg == nil {
		cfg = &config.JWTConfig{
			Secret:         "your-secret-key", // Should be loaded from environment
			Issuer:         "api-server",
			ExpirationTime: 24 * time.Hour,
			RefreshTime:    7 * 24 * time.Hour,
			Algorithm:      "HS256",
		}
	}
	
	if log == nil {
		log = logger.Default()
	}
	
	return &JWTManager{
		config: cfg,
		logger: log,
	}
}

// GenerateToken generates a new JWT token for a user
func (m *JWTManager) GenerateToken(userID, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.ExpirationTime)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	
	m.logger.WithFields(logger.Fields{
		"user_id": userID,
		"email":   email,
		"role":    role,
	}).Debug("Generated JWT token")
	
	return tokenString, nil
}

// GenerateRefreshToken generates a refresh token
func (m *JWTManager) GenerateRefreshToken(userID, email, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.config.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.RefreshTime)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	
	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.config.Secret), nil
	})
	
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, sharedErrors.TokenExpired("Token has expired")
		}
		return nil, sharedErrors.InvalidToken("Invalid token")
	}
	
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, sharedErrors.InvalidToken("Invalid token claims")
	}
	
	return claims, nil
}

// RefreshToken refreshes an existing token
func (m *JWTManager) RefreshToken(refreshTokenString string) (string, error) {
	claims, err := m.ValidateToken(refreshTokenString)
	if err != nil {
		return "", err
	}
	
	// Generate new access token
	return m.GenerateToken(claims.UserID, claims.Email, claims.Role)
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", sharedErrors.Unauthorized("Missing authorization header")
	}
	
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", sharedErrors.Unauthorized("Invalid authorization header format")
	}
	
	return parts[1], nil
}

// Password hashing utilities

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", sharedErrors.Validation("Password must be at least 8 characters long")
	}
	
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	
	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Middleware functions

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}
		
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			if appErr, ok := sharedErrors.AsAppError(err); ok {
				response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
			} else {
				response.Unauthorized(c, "Invalid token")
			}
			c.Abort()
			return
		}
		
		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		
		// Add user context to request context
		userCtx := &UserContext{
			UserID: claims.UserID,
			Email:  claims.Email,
			Role:   claims.Role,
		}
		ctx := context.WithValue(c.Request.Context(), "user", userCtx)
		ctx = context.WithValue(ctx, logger.UserIDKey, claims.UserID)
		c.Request = c.Request.WithContext(ctx)
		
		jwtManager.logger.WithFields(logger.Fields{
			"user_id": claims.UserID,
			"email":   claims.Email,
			"role":    claims.Role,
		}).Debug("User authenticated")
		
		c.Next()
	}
}

// RequireRole creates role-based authorization middleware
func RequireRole(roles ...string) gin.HandlerFunc {
	roleMap := make(map[string]bool)
	for _, role := range roles {
		roleMap[role] = true
	}
	
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			response.Unauthorized(c, "User not authenticated")
			c.Abort()
			return
		}
		
		if !roleMap[userRole.(string)] {
			response.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequireAdmin middleware requires admin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// OptionalAuth middleware that doesn't require authentication but extracts user info if present
func OptionalAuth(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}
		
		tokenString, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.Next()
			return
		}
		
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.Next()
			return
		}
		
		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		
		// Add user context to request context
		userCtx := &UserContext{
			UserID: claims.UserID,
			Email:  claims.Email,
			Role:   claims.Role,
		}
		ctx := context.WithValue(c.Request.Context(), "user", userCtx)
		ctx = context.WithValue(ctx, logger.UserIDKey, claims.UserID)
		c.Request = c.Request.WithContext(ctx)
		
		c.Next()
	}
}

// GetUserFromContext extracts user information from Gin context
func GetUserFromContext(c *gin.Context) (*UserContext, bool) {
	userID, hasUserID := c.Get("user_id")
	email, hasEmail := c.Get("user_email")
	role, hasRole := c.Get("user_role")
	
	if !hasUserID || !hasEmail || !hasRole {
		return nil, false
	}
	
	return &UserContext{
		UserID: userID.(string),
		Email:  email.(string),
		Role:   role.(string),
	}, true
}

// GetUserFromRequestContext extracts user information from request context
func GetUserFromRequestContext(ctx context.Context) (*UserContext, bool) {
	user, ok := ctx.Value("user").(*UserContext)
	return user, ok
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// HasRole checks if the current user has the specified role
func HasRole(c *gin.Context, role string) bool {
	userRole, exists := c.Get("user_role")
	if !exists {
		return false
	}
	return userRole.(string) == role
}

// IsAdmin checks if the current user is an admin
func IsAdmin(c *gin.Context) bool {
	return HasRole(c, "admin")
}

// CanAccessResource checks if the current user can access a resource
func CanAccessResource(c *gin.Context, resourceUserID string) bool {
	userID, exists := c.Get("user_id")
	if !exists {
		return false
	}
	
	// Users can access their own resources, admins can access any resource
	return userID.(string) == resourceUserID || IsAdmin(c)
}