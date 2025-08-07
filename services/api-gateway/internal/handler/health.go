package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"api-gateway/internal/config"
	"api-gateway/pkg/logger"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	config *config.Config
	logger *logger.Logger
}

func NewHealthHandler(config *config.Config, logger *logger.Logger) *HealthHandler {
	return &HealthHandler{
		config: config,
		logger: logger,
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "api-gateway",
		"time":    time.Now(),
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	services := map[string]string{
		"auth_service": h.config.Services.AuthService + "/health",
		"user_service": h.config.Services.UserService + "/health",
	}

	serviceStatus := make(map[string]interface{})
	allHealthy := true

	for name, url := range services {
		status := h.checkServiceHealth(url)
		serviceStatus[name] = status
		if !status["healthy"].(bool) {
			allHealthy = false
		}
	}

	statusCode := http.StatusOK
	if !allHealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status":   map[string]bool{"ready": allHealthy},
		"services": serviceStatus,
		"time":     time.Now(),
	})
}

func (h *HealthHandler) checkServiceHealth(url string) map[string]interface{} {
	client := &http.Client{Timeout: 5 * time.Second}
	
	resp, err := client.Get(url)
	if err != nil {
		return map[string]interface{}{
			"healthy": false,
			"error":   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{
			"healthy":     false,
			"status_code": resp.StatusCode,
		}
	}

	var healthResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&healthResponse); err != nil {
		return map[string]interface{}{
			"healthy": false,
			"error":   "failed to decode health response",
		}
	}

	return map[string]interface{}{
		"healthy": true,
		"response": healthResponse,
	}
}