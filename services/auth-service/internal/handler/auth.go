package handler

import (
	"net/http"

	"auth-service/internal/model"
	"auth-service/internal/service"
	"auth-service/pkg/logger"
	"auth-service/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
	logger      *logger.Logger
}

func NewAuthHandler(authService *service.AuthService, logger *logger.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) GenerateTokens(c *gin.Context) {
	userID := c.PostForm("user_id")
	email := c.PostForm("email")
	role := c.PostForm("role")

	if userID == "" || email == "" || role == "" {
		response.Error(c, http.StatusBadRequest, "MISSING_FIELDS", "user_id, email, and role are required")
		return
	}

	tokens, err := h.authService.GenerateTokens(userID, email, role)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate tokens")
		response.Error(c, http.StatusInternalServerError, "TOKEN_GENERATION_FAILED", "Failed to generate tokens")
		return
	}

	response.Success(c, "Tokens generated successfully", tokens)
}

func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var req model.ValidateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format")
		return
	}

	result, err := h.authService.ValidateToken(req.Token)
	if err != nil {
		h.logger.WithError(err).Error("Failed to validate token")
		response.Error(c, http.StatusInternalServerError, "VALIDATION_FAILED", "Failed to validate token")
		return
	}

	response.Success(c, "Token validation completed", result)
}

func (h *AuthHandler) RefreshTokens(c *gin.Context) {
	var req model.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format")
		return
	}

	tokens, err := h.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to refresh tokens")
		response.Error(c, http.StatusUnauthorized, "REFRESH_FAILED", "Invalid or expired refresh token")
		return
	}

	response.Success(c, "Tokens refreshed successfully", tokens)
}