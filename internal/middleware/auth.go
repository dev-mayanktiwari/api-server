package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/dev-mayanktiwari/api-server/pkg/auth"
	"github.com/dev-mayanktiwari/api-server/pkg/logger"
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(jwtManager *auth.JWTManager, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.WithRequestID(c.GetString("request_id")).
				Warn("Missing authorization header")
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header is required",
				"error": gin.H{
					"code":    "MISSING_AUTH_HEADER",
					"message": "Authorization header is required",
				},
				"timestamp":  time.Now(),
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		// Extract token from header
		tokenString, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			logger.WithRequestID(c.GetString("request_id")).
				WithError(err).
				Warn("Invalid authorization header format")
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization header format",
				"error": gin.H{
					"code":    "INVALID_AUTH_HEADER",
					"message": "Authorization header must be in format: Bearer <token>",
				},
				"timestamp":  time.Now(),
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			logger.WithRequestID(c.GetString("request_id")).
				WithError(err).
				Warn("Invalid JWT token")
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid or expired token",
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "The provided token is invalid or has expired",
				},
				"timestamp":  time.Now(),
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("jwt_claims", claims)

		// Log successful authentication
		logger.WithRequestID(c.GetString("request_id")).
			WithFields(map[string]interface{}{
				"user_id": claims.UserID,
				"email":   claims.Email,
				"role":    claims.Role,
			}).
			Debug("User authenticated successfully")

		c.Next()
	}
}

// OptionalAuthMiddleware creates an optional JWT authentication middleware
// This middleware will parse the token if present but won't fail if missing
func OptionalAuthMiddleware(jwtManager *auth.JWTManager, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			c.Next()
			return
		}

		// Extract token from header
		tokenString, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("jwt_claims", claims)

		c.Next()
	}
}

// RoleMiddleware creates role-based authorization middleware
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authentication required",
				"error": gin.H{
					"code":    "AUTHENTICATION_REQUIRED",
					"message": "This endpoint requires authentication",
				},
				"timestamp":  time.Now(),
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		userRoleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Internal server error",
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Invalid user role in token",
				},
				"timestamp":  time.Now(),
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		// Check if user role is in allowed roles
		for _, allowedRole := range allowedRoles {
			if strings.EqualFold(userRoleStr, allowedRole) {
				c.Next()
				return
			}
		}

		// User doesn't have required role
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Insufficient permissions",
			"error": gin.H{
				"code":    "INSUFFICIENT_PERMISSIONS",
				"message": "You don't have permission to access this resource",
			},
			"timestamp":  time.Now(),
			"request_id": c.GetString("request_id"),
		})
		c.Abort()
	}
}

// AdminMiddleware creates admin-only authorization middleware
func AdminMiddleware() gin.HandlerFunc {
	return RoleMiddleware("admin")
}

// UserMiddleware creates user+ authorization middleware (user, admin)
func UserMiddleware() gin.HandlerFunc {
	return RoleMiddleware("user", "admin")
}