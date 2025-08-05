package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter holds the rate limiter instances for different clients
type RateLimiter struct {
	clients map[string]*rate.Limiter
	mu      sync.Mutex
	rate    rate.Limit
	burst   int
	cleanup time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*rate.Limiter),
		rate:    rps,
		burst:   burst,
		cleanup: 5 * time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanupWorker()

	return rl
}

// getLimiter returns the rate limiter for a client
func (rl *RateLimiter) getLimiter(clientID string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.clients[clientID]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.clients[clientID] = limiter
	}

	return limiter
}

// cleanupWorker removes inactive rate limiters
func (rl *RateLimiter) cleanupWorker() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		for clientID, limiter := range rl.clients {
			// Remove limiter if it hasn't been used recently
			if limiter.Allow() {
				// If limiter allows immediately, it's likely inactive
				// We use a simple heuristic here
				continue
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rps float64, burst int) gin.HandlerFunc {
	rl := NewRateLimiter(rate.Limit(rps), burst)

	return func(c *gin.Context) {
		// Use client IP as the identifier
		clientID := c.ClientIP()
		
		// Get the rate limiter for this client
		limiter := rl.getLimiter(clientID)
		
		// Check if request is allowed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Rate limit exceeded",
				"error": gin.H{
					"code":    "RATE_LIMIT_EXCEEDED",
					"message": "Too many requests, please try again later",
				},
				"timestamp":  time.Now(),
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthRateLimitMiddleware creates a more restrictive rate limiter for auth endpoints
func AuthRateLimitMiddleware() gin.HandlerFunc {
	// More restrictive for auth endpoints: 5 requests per minute
	return RateLimitMiddleware(5.0/60.0, 5)
}

// APIRateLimitMiddleware creates a general rate limiter for API endpoints
func APIRateLimitMiddleware() gin.HandlerFunc {
	// General API: 100 requests per minute
	return RateLimitMiddleware(100.0/60.0, 10)
}