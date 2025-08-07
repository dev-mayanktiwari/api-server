// Package handlers contains HTTP handlers for the user service.
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/auth"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/response"
	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/application/dto"
	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/application/services"
	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/domain/entities"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userAppService *services.UserApplicationService
	logger         *logger.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userAppService *services.UserApplicationService, log *logger.Logger) *UserHandler {
	return &UserHandler{
		userAppService: userAppService,
		logger:         log,
	}
}

// RegisterRoutes registers user routes
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// Public routes
	router.POST("/register", h.Register)
	router.POST("/login", h.Login)

	// Protected routes
	protected := router.Group("", authMiddleware)
	{
		// User profile routes
		protected.GET("/profile", h.GetProfile)
		protected.PUT("/profile", h.UpdateProfile)
		protected.POST("/change-password", h.ChangePassword)

		// Admin routes
		admin := protected.Group("", auth.RequireAdmin())
		{
			admin.GET("/users", h.ListUsers)
			admin.GET("/users/:id", h.GetUser)
			admin.PUT("/users/:id", h.UpdateUser)
			admin.DELETE("/users/:id", h.DeleteUser)
			admin.GET("/users/stats", h.GetUserStatistics)
		}
	}
}

// Register creates a new user
// @Summary Register a new user
// @Description Create a new user account
// @Tags Users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User registration data"
// @Success 201 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logger.Fields{
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		}).Warn("Invalid registration request")
		response.ValidationError(c, err)
		return
	}

	user, err := h.userAppService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithFields(logger.Fields{
			"email":      req.Email,
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		}).Error("Failed to create user")

		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		} else {
			response.InternalServerError(c, "Failed to create user")
		}
		return
	}

	response.Created(c, "User created successfully", user)
}

// Login authenticates a user
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags Users
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Login credentials"
// @Success 200 {object} response.Response{data=services.LoginResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logger.Fields{
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		}).Warn("Invalid login request")
		response.ValidationError(c, err)
		return
	}

	loginResponse, err := h.userAppService.LoginUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.WithFields(logger.Fields{
			"email":      req.Email,
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		}).Warn("Login failed")

		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		} else {
			response.Unauthorized(c, "Login failed")
		}
		return
	}

	response.Success(c, "Login successful", loginResponse)
}

// GetProfile gets the current user's profile
// @Summary Get user profile
// @Description Get the current authenticated user's profile
// @Tags Users
// @Produce json
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := h.userAppService.GetUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithFields(logger.Fields{
			"user_id":    userID,
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		}).Error("Failed to get user profile")

		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		} else {
			response.NotFound(c, "User not found")
		}
		return
	}

	response.Success(c, "Profile retrieved successfully", user)
}

// UpdateProfile updates the current user's profile
// @Summary Update user profile
// @Description Update the current authenticated user's profile
// @Tags Users
// @Accept json
// @Produce json
// @Param user body dto.UpdateUserRequest true "User update data"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	userRole := entities.UserRole(c.GetString("user_role"))

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logger.Fields{
			"user_id":    userID,
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		}).Warn("Invalid update profile request")
		response.ValidationError(c, err)
		return
	}

	user, err := h.userAppService.UpdateUser(c.Request.Context(), userID, &req, userID, userRole)
	if err != nil {
		h.logger.WithFields(logger.Fields{
			"user_id":    userID,
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		}).Error("Failed to update user profile")

		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		} else {
			response.InternalServerError(c, "Failed to update profile")
		}
		return
	}

	response.Success(c, "Profile updated successfully", user)
}

// ChangePassword changes the current user's password
// @Summary Change password
// @Description Change the current authenticated user's password
// @Tags Users
// @Accept json
// @Produce json
// @Param password body dto.ChangePasswordRequest true "Password change data"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/change-password [post]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logger.Fields{
			"user_id":    userID,
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		}).Warn("Invalid change password request")
		response.ValidationError(c, err)
		return
	}

	if err := h.userAppService.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		h.logger.WithFields(logger.Fields{
			"user_id":    userID,
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		}).Error("Failed to change password")

		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		} else {
			response.InternalServerError(c, "Failed to change password")
		}
		return
	}

	response.Success(c, "Password changed successfully", nil)
}

// GetUser gets a user by ID (admin only)
// @Summary Get user by ID
// @Description Get user information by ID (admin only)
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := h.userAppService.GetUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithFields(logger.Fields{
			"target_user_id": userID,
			"error":          err.Error(),
			"request_id":     c.GetString("request_id"),
		}).Error("Failed to get user")

		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		} else {
			response.NotFound(c, "User not found")
		}
		return
	}

	response.Success(c, "User retrieved successfully", user)
}

// UpdateUser updates a user (admin only)
// @Summary Update user
// @Description Update user information (admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body dto.UpdateUserRequest true "User update data"
// @Success 200 {object} response.Response{data=dto.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	targetUserID := c.Param("id")
	currentUserID := c.GetString("user_id")
	currentUserRole := entities.UserRole(c.GetString("user_role"))

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logger.Fields{
			"target_user_id": targetUserID,
			"current_user_id": currentUserID,
			"error":          err.Error(),
			"request_id":     c.GetString("request_id"),
		}).Warn("Invalid update user request")
		response.ValidationError(c, err)
		return
	}

	user, err := h.userAppService.UpdateUser(c.Request.Context(), targetUserID, &req, currentUserID, currentUserRole)
	if err != nil {
		h.logger.WithFields(logger.Fields{
			"target_user_id":  targetUserID,
			"current_user_id": currentUserID,
			"error":           err.Error(),
			"request_id":      c.GetString("request_id"),
		}).Error("Failed to update user")

		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		} else {
			response.InternalServerError(c, "Failed to update user")
		}
		return
	}

	response.Success(c, "User updated successfully", user)
}

// DeleteUser deletes a user (admin only)
// @Summary Delete user
// @Description Delete a user (admin only)
// @Tags Users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	targetUserID := c.Param("id")
	currentUserID := c.GetString("user_id")
	currentUserRole := entities.UserRole(c.GetString("user_role"))

	if err := h.userAppService.DeleteUser(c.Request.Context(), targetUserID, currentUserID, currentUserRole); err != nil {
		h.logger.WithFields(logger.Fields{
			"target_user_id":  targetUserID,
			"current_user_id": currentUserID,
			"error":           err.Error(),
			"request_id":      c.GetString("request_id"),
		}).Error("Failed to delete user")

		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		} else {
			response.InternalServerError(c, "Failed to delete user")
		}
		return
	}

	response.Success(c, "User deleted successfully", nil)
}

// ListUsers lists all users with pagination (admin only)
// @Summary List users
// @Description List all users with pagination and filtering (admin only)
// @Tags Users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size" default(10)
// @Param role query string false "Filter by role"
// @Param status query string false "Filter by status"
// @Param search query string false "Search term"
// @Success 200 {object} response.Response{data=dto.ListUsersResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	currentUserRole := entities.UserRole(c.GetString("user_role"))

	// Parse query parameters
	req := &dto.ListUsersRequest{
		Page:  1,
		Limit: 10,
	}

	if err := c.ShouldBindQuery(req); err != nil {
		h.logger.WithFields(logger.Fields{
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		}).Warn("Invalid list users request")
		response.ValidationError(c, err)
		return
	}

	// Validate pagination parameters
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 10
	}

	users, err := h.userAppService.ListUsers(c.Request.Context(), req, currentUserRole)
	if err != nil {
		h.logger.WithFields(logger.Fields{
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		}).Error("Failed to list users")

		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		} else {
			response.InternalServerError(c, "Failed to list users")
		}
		return
	}

	response.Success(c, "Users retrieved successfully", users)
}

// GetUserStatistics gets user statistics (admin only)
// @Summary Get user statistics
// @Description Get user statistics (admin only)
// @Tags Users
// @Produce json
// @Success 200 {object} response.Response{data=dto.UserStatsResponse}
// @Failure 401 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/users/stats [get]
func (h *UserHandler) GetUserStatistics(c *gin.Context) {
	currentUserRole := entities.UserRole(c.GetString("user_role"))

	stats, err := h.userAppService.GetUserStatistics(c.Request.Context(), currentUserRole)
	if err != nil {
		h.logger.WithFields(logger.Fields{
			"error":      err.Error(),
			"request_id": c.GetString("request_id"),
		}).Error("Failed to get user statistics")

		if appErr, ok := err.(*errors.AppError); ok {
			response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)
		} else {
			response.InternalServerError(c, "Failed to get user statistics")
		}
		return
	}

	response.Success(c, "User statistics retrieved successfully", stats)
}

// Add missing DTOs for completeness

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}