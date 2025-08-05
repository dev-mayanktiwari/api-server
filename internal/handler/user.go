package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/dev-mayanktiwari/api-server/internal/middleware"
	"github.com/dev-mayanktiwari/api-server/internal/model"
	"github.com/dev-mayanktiwari/api-server/internal/service"
	"github.com/dev-mayanktiwari/api-server/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/pkg/response"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *service.UserService
	logger      *logger.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *service.UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// Register creates a new user
func (h *UserHandler) Register(c *gin.Context) {
	var req model.CreateUserRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid registration request")
		c.JSON(http.StatusBadRequest, middleware.HandleValidationErrors(err))
		return
	}
	
	user, err := h.userService.CreateUser(&req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create user")
		
		// Check for specific errors
		if err.Error() == "user with email "+req.Email+" already exists" {
			response.Error(c, http.StatusConflict, "USER_ALREADY_EXISTS", "A user with this email already exists")
			return
		}
		
		response.Error(c, http.StatusInternalServerError, "USER_CREATION_FAILED", "Failed to create user")
		return
	}
	
	response.Success(c, "User created successfully", user)
}

// Login authenticates a user
func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid login request")
		c.JSON(http.StatusBadRequest, middleware.HandleValidationErrors(err))
		return
	}
	
	loginResponse, err := h.userService.Login(&req)
	if err != nil {
		h.logger.WithError(err).Warn("Login failed")
		
		// Don't expose specific error details for security
		response.Error(c, http.StatusUnauthorized, "LOGIN_FAILED", "Invalid email or password")
		return
	}
	
	response.Success(c, "Login successful", loginResponse)
}

// GetProfile gets the current user's profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user profile")
		response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}
	
	response.Success(c, "Profile retrieved successfully", user)
}

// UpdateProfile updates the current user's profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	userRole := c.GetString("user_role")
	
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid update profile request")
		c.JSON(http.StatusBadRequest, middleware.HandleValidationErrors(err))
		return
	}
	
	user, err := h.userService.UpdateUser(userID, &req, userID, userRole)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update user profile")
		
		if err.Error() == "email is already taken" {
			response.Error(c, http.StatusConflict, "EMAIL_ALREADY_TAKEN", "This email is already taken")
			return
		}
		
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update profile")
		return
	}
	
	response.Success(c, "Profile updated successfully", user)
}

// ChangePassword changes the current user's password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.GetString("user_id")
	
	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid change password request")
		c.JSON(http.StatusBadRequest, middleware.HandleValidationErrors(err))
		return
	}
	
	err := h.userService.ChangePassword(userID, &req)
	if err != nil {
		h.logger.WithError(err).Warn("Failed to change password")
		
		if err.Error() == "current password is incorrect" {
			response.Error(c, http.StatusBadRequest, "INCORRECT_PASSWORD", "Current password is incorrect")
			return
		}
		
		response.Error(c, http.StatusInternalServerError, "PASSWORD_CHANGE_FAILED", "Failed to change password")
		return
	}
	
	response.Success(c, "Password changed successfully", nil)
}

// Admin endpoints

// GetUser gets a user by ID (admin only)
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user")
		response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
		return
	}
	
	response.Success(c, "User retrieved successfully", user)
}

// UpdateUser updates a user (admin only)
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	currentUserID := c.GetString("user_id")
	currentUserRole := c.GetString("user_role")
	
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid update user request")
		c.JSON(http.StatusBadRequest, middleware.HandleValidationErrors(err))
		return
	}
	
	user, err := h.userService.UpdateUser(userID, &req, currentUserID, currentUserRole)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update user")
		
		if err.Error() == "insufficient permissions to update user" {
			response.Error(c, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", "You don't have permission to update this user")
			return
		}
		
		if err.Error() == "email is already taken" {
			response.Error(c, http.StatusConflict, "EMAIL_ALREADY_TAKEN", "This email is already taken")
			return
		}
		
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update user")
		return
	}
	
	response.Success(c, "User updated successfully", user)
}

// DeleteUser deletes a user (admin only)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	currentUserID := c.GetString("user_id")
	currentUserRole := c.GetString("user_role")
	
	err := h.userService.DeleteUser(userID, currentUserID, currentUserRole)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete user")
		
		if err.Error() == "insufficient permissions to delete user" {
			response.Error(c, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", "You don't have permission to delete users")
			return
		}
		
		if err.Error() == "cannot delete your own account" {
			response.Error(c, http.StatusBadRequest, "CANNOT_DELETE_SELF", "You cannot delete your own account")
			return
		}
		
		if err.Error() == "user not found" {
			response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}
		
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete user")
		return
	}
	
	response.Success(c, "User deleted successfully", nil)
}

// ListUsers lists all users with pagination (admin only)
func (h *UserHandler) ListUsers(c *gin.Context) {
	currentUserRole := c.GetString("user_role")
	
	// Get pagination parameters
	page := 1
	limit := 10
	
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	users, total, err := h.userService.ListUsers(page, limit, currentUserRole)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list users")
		
		if err.Error() == "insufficient permissions to list users" {
			response.Error(c, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", "You don't have permission to list users")
			return
		}
		
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list users")
		return
	}
	
	// Calculate pagination info
	totalPages := (int(total) + limit - 1) / limit
	
	response.Success(c, "Users retrieved successfully", gin.H{
		"users": users,
		"pagination": gin.H{
			"current_page": page,
			"total_pages":  totalPages,
			"per_page":     limit,
			"total_items":  total,
		},
	})
}