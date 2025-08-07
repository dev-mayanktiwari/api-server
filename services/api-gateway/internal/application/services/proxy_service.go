package services

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/response"
)

// ProxyService handles proxying requests to downstream services
type ProxyService struct {
	logger           *logger.Logger
	authServiceURL   string
	userServiceURL   string
	httpClient       *http.Client
}

// NewProxyService creates a new proxy service
func NewProxyService(logger *logger.Logger) *ProxyService {
	return &ProxyService{
		logger:           logger,
		authServiceURL:   getEnvOrDefault("AUTH_SERVICE_URL", "http://auth-service:8081"),
		userServiceURL:   getEnvOrDefault("USER_SERVICE_URL", "http://user-service:8082"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyToService proxies a request to the specified service
func (s *ProxyService) ProxyToService(c *gin.Context, serviceURL string) {
	// Build target URL
	targetURL := fmt.Sprintf("%s%s", serviceURL, c.Request.URL.Path)
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	// Read request body
	var body []byte
	if c.Request.Body != nil {
		var err error
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			s.logger.WithError(err).Error("Failed to read request body")
			response.InternalServerError(c, "Failed to process request")
			return
		}
	}

	// Create new request
	req, err := http.NewRequestWithContext(
		c.Request.Context(),
		c.Request.Method,
		targetURL,
		bytes.NewReader(body),
	)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create proxy request")
		response.InternalServerError(c, "Failed to create proxy request")
		return
	}

	// Copy headers from original request
	for name, values := range c.Request.Header {
		// Skip hop-by-hop headers
		if isHopByHopHeader(name) {
			continue
		}
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// Add X-Forwarded headers
	req.Header.Set("X-Forwarded-For", c.ClientIP())
	req.Header.Set("X-Forwarded-Proto", "http")
	if c.Request.Host != "" {
		req.Header.Set("X-Forwarded-Host", c.Request.Host)
	}

	// Add request ID if available
	if requestID, exists := c.Get("request_id"); exists {
		req.Header.Set("X-Request-ID", requestID.(string))
	}

	// Add user context headers if available
	if userID, exists := c.Get("user_id"); exists {
		req.Header.Set("X-User-ID", userID.(string))
	}
	if userEmail, exists := c.Get("user_email"); exists {
		req.Header.Set("X-User-Email", userEmail.(string))
	}
	if userRole, exists := c.Get("user_role"); exists {
		req.Header.Set("X-User-Role", userRole.(string))
	}

	// Log the proxy request
	s.logger.WithFields(logger.Fields{
		"method":     req.Method,
		"target_url": targetURL,
		"user_id":    req.Header.Get("X-User-ID"),
		"request_id": req.Header.Get("X-Request-ID"),
	}).Debug("Proxying request to service")

	// Execute the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.WithError(err).WithField("target_url", targetURL).Error("Failed to proxy request")
		response.BadGateway(c, "Service temporarily unavailable")
		return
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.WithError(err).Error("Failed to read response body")
		response.InternalServerError(c, "Failed to process response")
		return
	}

	// Copy response headers
	for name, values := range resp.Header {
		// Skip hop-by-hop headers
		if isHopByHopHeader(name) {
			continue
		}
		for _, value := range values {
			c.Header(name, value)
		}
	}

	// Log successful proxy
	s.logger.WithFields(logger.Fields{
		"method":      req.Method,
		"target_url":  targetURL,
		"status_code": resp.StatusCode,
		"user_id":     req.Header.Get("X-User-ID"),
		"request_id":  req.Header.Get("X-Request-ID"),
	}).Debug("Successfully proxied request")

	// Write response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// ProxyToAuthService proxies requests to the auth service
func (s *ProxyService) ProxyToAuthService(c *gin.Context) {
	s.ProxyToService(c, s.authServiceURL)
}

// ProxyToUserService proxies requests to the user service
func (s *ProxyService) ProxyToUserService(c *gin.Context) {
	s.ProxyToService(c, s.userServiceURL)
}

// isHopByHopHeader checks if a header is hop-by-hop
func isHopByHopHeader(header string) bool {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	header = strings.ToLower(header)
	for _, hopHeader := range hopByHopHeaders {
		if strings.ToLower(hopHeader) == header {
			return true
		}
	}
	return false
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}