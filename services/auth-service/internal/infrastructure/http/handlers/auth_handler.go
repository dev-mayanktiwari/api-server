package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/auth"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/errors"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/response"
	"github.com/dev-mayanktiwari/api-server/services/auth-service/internal/application/dto"
	"github.com/dev-mayanktiwari/api-server/services/auth-service/internal/application/services"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *services.AuthApplicationService
	logger      *logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthApplicationService, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	result, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Login successful", result)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	result, err := h.authService.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Token refreshed successfully", result)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	// Get access token from Authorization header if not provided in body
	if req.AccessToken == "" {
		authHeader := c.GetHeader("Authorization")
		if token, err := auth.ExtractTokenFromHeader(authHeader); err == nil {
			req.AccessToken = token
		}
	}

	if err := h.authService.Logout(c.Request.Context(), &req); err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Logout successful", nil)
}

// ValidateToken handles token validation
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var req dto.ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	result, err := h.authService.ValidateToken(c.Request.Context(), &req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Token validation result", result)
}

// GetCurrentUser handles getting current user information
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	result, err := h.authService.GetCurrentUser(c.Request.Context(), userID.(string))
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "Current user information", result)
}

// handleError handles application errors
func (h *AuthHandler) handleError(c *gin.Context, err error) {
	if appErr, ok := errors.AsAppError(err); ok {
		response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}

	h.logger.WithContext(c.Request.Context()).WithError(err).Error("Unhandled error in auth handler")
	response.InternalServerError(c, "An internal error occurred")
}