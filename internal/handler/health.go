package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/dev-mayanktiwari/api-server/pkg/response"
)

// HealthHandler handles health-related endpoints
type HealthHandler struct {
	startTime time.Time
	version   string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(version string) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		version:   version,
	}
}

// HealthResponse represents the health check response structure
type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	System    SystemInfo        `json:"system"`
	Checks    map[string]string `json:"checks"`
}

// SystemInfo represents system information
type SystemInfo struct {
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	GoVersion  string `json:"go_version"`
	NumCPU     int    `json:"num_cpu"`
	Goroutines int    `json:"goroutines"`
}

// Health returns the basic health status of the application
// @Summary Health check
// @Description Get the health status of the application
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse{data=HealthResponse}
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	uptime := time.Since(h.startTime)
	
	healthResp := HealthResponse{
		Status:    "healthy",
		Version:   h.version,
		Timestamp: time.Now(),
		Uptime:    uptime.String(),
		System: SystemInfo{
			OS:         runtime.GOOS,
			Arch:       runtime.GOARCH,
			GoVersion:  runtime.Version(),
			NumCPU:     runtime.NumCPU(),
			Goroutines: runtime.NumGoroutine(),
		},
		Checks: map[string]string{
			"api": "healthy",
			// We'll add database check later
		},
	}

	response.Success(c, "Application is healthy", healthResp)
}

// Ready returns the readiness status of the application
// @Summary Readiness check
// @Description Get the readiness status of the application
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse
// @Service 503 {object} response.APIResponse
// @Router /ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	// Check if all dependencies are ready
	checks := make(map[string]string)
	allReady := true

	// API is always ready if we reach this point
	checks["api"] = "ready"

	// TODO: Add database connectivity check
	// if db connection fails {
	//     checks["database"] = "not ready"
	//     allReady = false
	// } else {
	//     checks["database"] = "ready"
	// }

	// TODO: Add other dependency checks (Redis, external APIs, etc.)

	if allReady {
		response.Success(c, "Application is ready", gin.H{
			"status": "ready",
			"checks": checks,
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success":    false,
			"message":    "Application is not ready",
			"error": gin.H{
				"code":    "SERVICE_UNAVAILABLE",
				"message": "One or more dependencies are not ready",
			},
			"data": gin.H{
				"status": "not ready",
				"checks": checks,
			},
			"timestamp": time.Now(),
		})
	}
}

// Liveness returns the liveness status of the application
// @Summary Liveness check
// @Description Get the liveness status of the application (for Kubernetes)
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /live [get]
func (h *HealthHandler) Liveness(c *gin.Context) {
	// This endpoint should only fail if the application is completely broken
	// Keep it simple - if we can respond, we're alive
	response.Success(c, "Application is alive", gin.H{
		"status":    "alive",
		"timestamp": time.Now(),
	})
}

// Version returns the application version information
// @Summary Version information
// @Description Get the version information of the application
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /version [get]
func (h *HealthHandler) Version(c *gin.Context) {
	versionInfo := gin.H{
		"version":    h.version,
		"build_time": h.startTime.Format(time.RFC3339),
		"go_version": runtime.Version(),
		"git_commit": "dev", // We'll make this dynamic later
	}

	response.Success(c, "Version information", versionInfo)
}