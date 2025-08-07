// Package middleware provides HTTP middleware components for the API server.
// It includes logging, authentication, CORS, rate limiting, and request ID middleware.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/auth"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/config"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/errors"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/response"
)

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		
		// Add to context for logger
		ctx := context.WithValue(c.Request.Context(), logger.RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)
		
		c.Next()
	}
}

// Logger middleware logs HTTP requests with detailed information
func Logger(log *logger.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		start := time.Now()
		path := param.Path
		method := param.Method
		statusCode := param.StatusCode
		latency := param.Latency
		clientIP := param.ClientIP
		userAgent := param.Request.UserAgent()
		requestID := param.Request.Context().Value(logger.RequestIDKey)
		
		fields := logger.Fields{
			"method":      method,
			"path":        path,
			"status_code": statusCode,
			"latency_ms":  latency.Milliseconds(),
			"client_ip":   clientIP,
			"user_agent":  userAgent,
			"timestamp":   start,
		}
		
		if requestID != nil {
			fields["request_id"] = requestID
		}
		
		// Get user ID from context if available
		if userID, exists := param.Keys["user_id"]; exists {
			fields["user_id"] = userID
		}
		
		log.LogHTTPRequest(method, path, statusCode, latency, fields)
		
		return "" // We handle logging ourselves
	})
}

// Recovery middleware handles panics and returns proper error responses
func Recovery(log *logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		_ = fmt.Errorf("panic recovered: %v", recovered)
		
		log.WithFields(logger.Fields{
			"request_id": c.GetString("request_id"),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"panic":      recovered,
		}).Error("Panic recovered in HTTP handler")
		
		response.InternalServerError(c, "An unexpected error occurred")
		c.Abort()
	})
}

// CORS middleware configures Cross-Origin Resource Sharing
func CORS(cfg ...*config.CORSConfig) gin.HandlerFunc {
	var corsConfig *config.CORSConfig
	if len(cfg) > 0 && cfg[0] != nil {
		corsConfig = cfg[0]
	}
	if corsConfig == nil {
		// Default CORS configuration
		corsConfig = &config.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"X-Request-ID"},
			AllowCredentials: false,
			MaxAge:           12 * time.Hour,
		}
	}
	
	corsConfiguration := cors.Config{
		AllowOrigins:     corsConfig.AllowOrigins,
		AllowMethods:     corsConfig.AllowMethods,
		AllowHeaders:     corsConfig.AllowHeaders,
		ExposeHeaders:    corsConfig.ExposeHeaders,
		AllowCredentials: corsConfig.AllowCredentials,
		MaxAge:           corsConfig.MaxAge,
	}
	
	return cors.New(corsConfiguration)
}

// RateLimit middleware implements rate limiting per IP
func RateLimit(cfg ...*config.RateLimitConfig) gin.HandlerFunc {
	var rateLimitConfig *config.RateLimitConfig
	if len(cfg) > 0 && cfg[0] != nil {
		rateLimitConfig = cfg[0]
	}
	if rateLimitConfig == nil {
		rateLimitConfig = &config.RateLimitConfig{
			RequestsPerSecond: 10,
			BurstSize:        20,
			CleanupInterval:  5 * time.Minute,
		}
	}
	
	limiter := NewRateLimiter(rate.Limit(rateLimitConfig.RequestsPerSecond), rateLimitConfig.BurstSize, rateLimitConfig.CleanupInterval)
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		if !limiter.Allow(clientIP) {
			response.TooManyRequests(c, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// Security middleware adds security headers
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		
		c.Next()
	}
}

// Timeout middleware adds request timeout
func Timeout(duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
		defer cancel()
		
		c.Request = c.Request.WithContext(ctx)
		
		finished := make(chan struct{})
		go func() {
			c.Next()
			finished <- struct{}{}
		}()
		
		select {
		case <-ctx.Done():
			response.Error(c, http.StatusRequestTimeout, "TIMEOUT", "Request timeout")
			c.Abort()
		case <-finished:
			return
		}
	}
}

// RateLimiter implements per-client rate limiting
type RateLimiter struct {
	limiters        map[string]*rate.Limiter
	rateLimit       rate.Limit
	burstSize       int
	cleanupInterval time.Duration
	
	// Use sync.Map for better concurrent access
	syncMap *syncMapWrapper
}

type syncMapWrapper struct {
	m map[string]*clientLimiter
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rateLimit rate.Limit, burstSize int, cleanupInterval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limiters:        make(map[string]*rate.Limiter),
		rateLimit:       rateLimit,
		burstSize:       burstSize,
		cleanupInterval: cleanupInterval,
		syncMap:         &syncMapWrapper{m: make(map[string]*clientLimiter)},
	}
	
	// Start cleanup goroutine
	go rl.cleanupRoutine()
	
	return rl
}

// Allow checks if a request is allowed for the given client
func (rl *RateLimiter) Allow(clientID string) bool {
	limiter := rl.getLimiter(clientID)
	return limiter.Allow()
}

func (rl *RateLimiter) getLimiter(clientID string) *rate.Limiter {
	// Try to get existing limiter
	if client, exists := rl.syncMap.m[clientID]; exists {
		client.lastSeen = time.Now()
		return client.limiter
	}
	
	// Create new limiter
	limiter := rate.NewLimiter(rl.rateLimit, rl.burstSize)
	rl.syncMap.m[clientID] = &clientLimiter{
		limiter:  limiter,
		lastSeen: time.Now(),
	}
	
	return limiter
}

func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		for clientID, client := range rl.syncMap.m {
			// Remove clients that haven't been seen for 2x cleanup interval
			if now.Sub(client.lastSeen) > 2*rl.cleanupInterval {
				delete(rl.syncMap.m, clientID)
			}
		}
	}
}

// ErrorHandler middleware handles application errors
func ErrorHandler(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Check for errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			
			log.WithFields(logger.Fields{
				"request_id": c.GetString("request_id"),
				"method":     c.Request.Method,
				"path":       c.Request.URL.Path,
				"error":      err.Error(),
			}).Error("Request error")
			
			// If response was already sent, don't send another
			if !c.Writer.Written() {
				response.InternalServerError(c, "An error occurred processing your request")
			}
		}
	}
}

// HealthCheck middleware for health check endpoints
func HealthCheck(path string, checks func() map[string]interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == path {
			status := "healthy"
			checkResults := make(map[string]interface{})
			
			if checks != nil {
				checkResults = checks()
				// Check if any checks failed
				for _, result := range checkResults {
					if resultMap, ok := result.(map[string]interface{}); ok {
						if status, exists := resultMap["status"]; exists && status != "healthy" {
							status = "unhealthy"
						}
					}
				}
			}
			
			response.HealthCheck(c, status, checkResults)
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// Metrics middleware for collecting request metrics
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		c.Next()
		
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.FullPath()
		
		// Here you would integrate with your metrics system
		// For example, Prometheus metrics
		// requestDuration.WithLabelValues(method, path, fmt.Sprintf("%d", statusCode)).Observe(duration.Seconds())
		// requestCount.WithLabelValues(method, path, fmt.Sprintf("%d", statusCode)).Inc()
		
		_ = duration
		_ = statusCode
		_ = method
		_ = path
	}
}

// JWTAuth middleware validates JWT tokens
func JWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			if appErr, ok := errors.AsAppError(err); ok {
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
		userCtx := &auth.UserContext{
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