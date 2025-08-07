package handler

import (
	"net/http"
	"strconv"

	"user-service/internal/model"
	"user-service/internal/service"
	"user-service/pkg/logger"
	"user-service/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
	logger      *logger.Logger
}

func NewUserHandler(userService *service.UserService, logger *logger.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req model.CreateUserRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid registration request")
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request format")
		return
	}
	
	user, err := h.userService.CreateUser(&req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create user")
		
		if err.Error() == "user with email "+req.Email+" already exists" {
			response.Error(c, http.StatusConflict, "USER_ALREADY_EXISTS", "A user with this email already exists")
			return
		}
		
		response.Error(c, http.StatusInternalServerError, "USER_CREATION_FAILED", "Failed to create user")
		return
	}
	
	response.Success(c, "User created successfully", user)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid login request")
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request format")
		return
	}
	
	loginResponse, err := h.userService.Login(&req)
	if err != nil {
		h.logger.WithError(err).Warn("Login failed")
		response.Error(c, http.StatusUnauthorized, "LOGIN_FAILED", "Invalid email or password")
		return
	}
	
	response.Success(c, "Login successful", loginResponse)
}

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

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	
	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid update user request")
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request format")
		return
	}
	
	user, err := h.userService.UpdateUser(userID, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update user")
		
		if err.Error() == "email is already taken" {
			response.Error(c, http.StatusConflict, "EMAIL_ALREADY_TAKEN", "This email is already taken")
			return
		}
		
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update user")
		return
	}
	
	response.Success(c, "User updated successfully", user)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.Param("id")
	
	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid change password request")
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request format")
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

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	
	err := h.userService.DeleteUser(userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete user")
		
		if err.Error() == "user not found" {
			response.Error(c, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}
		
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete user")
		return
	}
	
	response.Success(c, "User deleted successfully", nil)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
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
	
	users, total, err := h.userService.ListUsers(page, limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list users")
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list users")
		return
	}
	
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