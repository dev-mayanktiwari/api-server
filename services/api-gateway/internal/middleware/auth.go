package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"api-gateway/internal/config"
	"api-gateway/pkg/logger"

	"github.com/gin-gonic/gin"
)

type ValidateTokenResponse struct {
	Valid   bool   `json:"valid"`
	UserID  string `json:"user_id,omitempty"`
	Email   string `json:"email,omitempty"`
	Role    string `json:"role,omitempty"`
	Message string `json:"message,omitempty"`
}

func AuthMiddleware(config *config.Config, logger *logger.Logger) gin.HandlerFunc {
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

		tokenString, err := extractTokenFromHeader(authHeader)
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

		validationResult, err := validateToken(config.Services.AuthService, tokenString)
		if err != nil || !validationResult.Valid {
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

		c.Set("user_id", validationResult.UserID)
		c.Set("user_email", validationResult.Email)
		c.Set("user_role", validationResult.Role)

		logger.WithRequestID(c.GetString("request_id")).
			WithFields(map[string]interface{}{
				"user_id": validationResult.UserID,
				"email":   validationResult.Email,
				"role":    validationResult.Role,
			}).
			Debug("User authenticated successfully")

		c.Next()
	}
}

func extractTokenFromHeader(authHeader string) (string, error) {
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) {
		return "", gin.Error{Err: gin.Error{Err: nil}.Err, Type: gin.ErrorTypePublic}
	}

	if authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", gin.Error{Err: gin.Error{Err: nil}.Err, Type: gin.ErrorTypePublic}
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", gin.Error{Err: gin.Error{Err: nil}.Err, Type: gin.ErrorTypePublic}
	}

	return token, nil
}

func validateToken(authServiceURL, token string) (*ValidateTokenResponse, error) {
	payload := map[string]string{"token": token}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(authServiceURL+"/api/v1/auth/validate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response struct {
		Success bool                   `json:"success"`
		Data    *ValidateTokenResponse `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

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

		for _, allowedRole := range allowedRoles {
			if strings.EqualFold(userRoleStr, allowedRole) {
				c.Next()
				return
			}
		}

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