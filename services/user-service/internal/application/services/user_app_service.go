// Package services contains application services that orchestrate domain operations.
package services

import (
	"context"
	"strings"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/errors"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/application/dto"
	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/domain/entities"
	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/domain/repositories"
	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/domain/services"
)

// UserApplicationService provides application-level user operations
type UserApplicationService struct {
	userRepo        repositories.UserRepository
	userDomainSvc   *services.UserDomainService
	logger          *logger.Logger
}

// NewUserApplicationService creates a new user application service
func NewUserApplicationService(
	userRepo repositories.UserRepository,
	userDomainSvc *services.UserDomainService,
	log *logger.Logger,
) *UserApplicationService {
	return &UserApplicationService{
		userRepo:      userRepo,
		userDomainSvc: userDomainSvc,
		logger:        log,
	}
}

// CreateUser creates a new user
func (s *UserApplicationService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Validate business rules
	if err := s.userDomainSvc.ValidateUserCreation(ctx, req.Email, req.Password, req.FirstName, req.LastName); err != nil {
		return nil, err
	}

	// Convert DTO to entity
	user, err := req.ToEntity()
	if err != nil {
		return nil, err
	}

	// Create user in repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		s.logger.WithFields(logger.Fields{
			"email": req.Email,
			"error": err.Error(),
		}).Error("Failed to create user")
		return nil, err
	}

	s.logger.WithFields(logger.Fields{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
	}).Info("User created successfully")

	// Convert entity to response DTO
	response := &dto.UserResponse{}
	response.FromEntity(user)

	return response, nil
}

// ValidateCredentials validates user credentials (for auth service)
func (s *UserApplicationService) ValidateCredentials(ctx context.Context, email, password string) (*dto.UserResponse, error) {
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		s.logger.WithFields(logger.Fields{
			"email": email,
		}).Warn("Credential validation attempt with non-existent email")
		return nil, errors.Unauthorized("Invalid email or password")
	}

	// Check if user is active
	if !user.IsActive() {
		s.logger.WithFields(logger.Fields{
			"user_id": user.ID,
			"email":   user.Email,
			"status":  user.Status,
		}).Warn("Credential validation attempt by inactive user")
		return nil, errors.Unauthorized("Account is not active")
	}

	// Verify password
	if !user.VerifyPassword(password) {
		s.logger.WithFields(logger.Fields{
			"user_id": user.ID,
			"email":   user.Email,
		}).Warn("Credential validation attempt with incorrect password")
		return nil, errors.Unauthorized("Invalid email or password")
	}

	s.logger.WithFields(logger.Fields{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("User credentials validated successfully")

	// Convert to response
	userResponse := &dto.UserResponse{}
	userResponse.FromEntity(user)

	return userResponse, nil
}

// GetUser retrieves a user by ID
func (s *UserApplicationService) GetUser(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.NotFound("User not found")
	}

	response := &dto.UserResponse{}
	response.FromEntity(user)

	return response, nil
}

// UpdateUser updates a user
func (s *UserApplicationService) UpdateUser(ctx context.Context, userID string, req *dto.UpdateUserRequest, currentUserID string, currentUserRole entities.UserRole) (*dto.UserResponse, error) {
	// Check permissions
	if err := s.userDomainSvc.CanUpdateUser(ctx, userID, currentUserID, currentUserRole); err != nil {
		return nil, err
	}

	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.NotFound("User not found")
	}

	// Validate email update if provided
	if req.Email != "" {
		if err := s.userDomainSvc.ValidateEmailUpdate(ctx, userID, req.Email); err != nil {
			return nil, err
		}
	}

	// Validate role change if provided
	if req.Role != "" && req.Role != string(user.Role) {
		if err := s.userDomainSvc.CanChangeRole(ctx, userID, currentUserID, currentUserRole, entities.UserRole(req.Role)); err != nil {
			return nil, err
		}
	}

	// Apply updates
	if err := req.ApplyUpdateFields(user); err != nil {
		return nil, err
	}

	// Update in repository
	if err := s.userRepo.Update(ctx, user); err != nil {
		s.logger.WithFields(logger.Fields{
			"user_id":         userID,
			"current_user_id": currentUserID,
			"error":           err.Error(),
		}).Error("Failed to update user")
		return nil, err
	}

	s.logger.WithFields(logger.Fields{
		"user_id":         userID,
		"current_user_id": currentUserID,
		"email":           user.Email,
	}).Info("User updated successfully")

	// Convert to response
	response := &dto.UserResponse{}
	response.FromEntity(user)

	return response, nil
}

// ChangePassword changes a user's password
func (s *UserApplicationService) ChangePassword(ctx context.Context, userID string, req *dto.ChangePasswordRequest) error {
	// Validate and get user
	user, err := s.userDomainSvc.ValidatePasswordChange(ctx, userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return err
	}

	// Change password
	if err := user.ChangePassword(req.CurrentPassword, req.NewPassword); err != nil {
		return err
	}

	// Update password in repository
	if err := s.userRepo.UpdatePassword(ctx, userID, user.Password); err != nil {
		s.logger.WithFields(logger.Fields{
			"user_id": userID,
			"error":   err.Error(),
		}).Error("Failed to update password")
		return err
	}

	s.logger.WithFields(logger.Fields{
		"user_id": userID,
	}).Info("Password changed successfully")

	return nil
}

// DeleteUser deletes a user
func (s *UserApplicationService) DeleteUser(ctx context.Context, userID, currentUserID string, currentUserRole entities.UserRole) error {
	// Check permissions
	if err := s.userDomainSvc.CanDeleteUser(ctx, userID, currentUserID, currentUserRole); err != nil {
		return err
	}

	// Check if user exists
	exists, err := s.userRepo.ExistsByID(ctx, userID)
	if err != nil {
		return err
	}

	if !exists {
		return errors.NotFound("User not found")
	}

	// Delete user
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		s.logger.WithFields(logger.Fields{
			"user_id":         userID,
			"current_user_id": currentUserID,
			"error":           err.Error(),
		}).Error("Failed to delete user")
		return err
	}

	s.logger.WithFields(logger.Fields{
		"user_id":         userID,
		"current_user_id": currentUserID,
	}).Info("User deleted successfully")

	return nil
}

// ListUsers lists users with pagination and filtering
func (s *UserApplicationService) ListUsers(ctx context.Context, req *dto.ListUsersRequest, currentUserRole entities.UserRole) (*dto.ListUsersResponse, error) {
	// Only admins can list users
	if currentUserRole != entities.RoleAdmin {
		return nil, errors.Forbidden("Insufficient permissions to list users")
	}

	// Calculate offset
	offset := (req.Page - 1) * req.Limit

	var users []*entities.User
	var total int64
	var err error

	// Apply filters
	if req.Role != "" {
		users, total, err = s.userRepo.GetUsersByRole(ctx, entities.UserRole(req.Role), offset, req.Limit)
	} else if req.Status == "active" {
		users, total, err = s.userRepo.GetActiveUsers(ctx, offset, req.Limit)
	} else {
		users, total, err = s.userRepo.List(ctx, offset, req.Limit)
	}

	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	userResponses := make([]*dto.UserResponse, len(users))
	for i, user := range users {
		response := &dto.UserResponse{}
		response.FromEntity(user)
		userResponses[i] = response
	}

	// Calculate pagination
	pagination := dto.CalculatePagination(req.Page, req.Limit, total)

	return &dto.ListUsersResponse{
		Users:      userResponses,
		Pagination: pagination,
	}, nil
}

// GetUserStatistics retrieves user statistics
func (s *UserApplicationService) GetUserStatistics(ctx context.Context, currentUserRole entities.UserRole) (*dto.UserStatsResponse, error) {
	// Only admins can view statistics
	if currentUserRole != entities.RoleAdmin {
		return nil, errors.Forbidden("Insufficient permissions to view user statistics")
	}

	stats, err := s.userDomainSvc.GetUserStatistics(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.UserStatsResponse{
		TotalUsers:     stats.TotalUsers,
		ActiveUsers:    stats.ActiveUsers,
		InactiveUsers:  stats.InactiveUsers,
		SuspendedUsers: stats.SuspendedUsers,
		AdminUsers:     stats.AdminUsers,
		RegularUsers:   stats.RegularUsers,
	}, nil
}