package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/services/api-gateway/internal/application/services"
)

// GatewayHandler handles gateway-specific HTTP requests
type GatewayHandler struct {
	proxyService *services.ProxyService
	logger       *logger.Logger
}

// NewGatewayHandler creates a new gateway handler
func NewGatewayHandler(proxyService *services.ProxyService, logger *logger.Logger) *GatewayHandler {
	return &GatewayHandler{
		proxyService: proxyService,
		logger:       logger,
	}
}

// ProxyToAuthService proxies requests to the auth service
func (h *GatewayHandler) ProxyToAuthService(c *gin.Context) {
	h.proxyService.ProxyToAuthService(c)
}

// ProxyToUserService proxies requests to the user service  
func (h *GatewayHandler) ProxyToUserService(c *gin.Context) {
	h.proxyService.ProxyToUserService(c)
}