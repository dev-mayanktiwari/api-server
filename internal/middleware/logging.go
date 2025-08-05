package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/dev-mayanktiwari/api-server/pkg/logger"
)

// LoggingMiddleware logs HTTP requests with structured logging
func LoggingMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Get request ID from context
		requestID := ""
		if param.Keys != nil {
			if id, exists := param.Keys["request_id"]; exists {
				requestID = id.(string)
			}
		}

		// Create a contextual logger with request ID
		contextLogger := logger.WithRequestID(requestID)

		// Log the HTTP request
		contextLogger.LogHTTPRequest(
			param.Method,
			param.Path,
			param.Request.UserAgent(),
			param.ClientIP,
			param.StatusCode,
			param.Latency.Nanoseconds()/1e6, // Convert to milliseconds
		)

		// Log errors if status code indicates an error
		if param.StatusCode >= 400 {
			contextLogger.WithFields(map[string]interface{}{
				"method":      param.Method,
				"path":        param.Path,
				"status_code": param.StatusCode,
				"client_ip":   param.ClientIP,
				"user_agent":  param.Request.UserAgent(),
				"error":       param.ErrorMessage,
			}).Error("HTTP request failed")
		}

		// Return empty string as we handle logging ourselves
		return ""
	})
}

// RecoveryMiddleware handles panics and logs them
func RecoveryMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultWriter, func(c *gin.Context, recovered interface{}) {
		// Get request ID
		requestID := ""
		if id, exists := c.Get("request_id"); exists {
			requestID = id.(string)
		}

		contextLogger := logger.WithRequestID(requestID)

		// Log the panic
		contextLogger.WithFields(map[string]interface{}{
			"method":    c.Request.Method,
			"path":      c.Request.URL.Path,
			"client_ip": c.ClientIP(),
			"panic":     recovered,
		}).Error("Panic recovered")

		// Return 500 error
		c.JSON(500, gin.H{
			"success":    false,
			"message":    "Internal server error",
			"error":      gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"message": "An unexpected error occurred",
			},
			"timestamp":  time.Now(),
			"request_id": requestID,
		})
	})
}