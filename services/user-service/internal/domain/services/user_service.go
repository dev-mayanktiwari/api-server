// Package services contains domain services for business logic that doesn't naturally fit into entities.
package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/errors"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/domain/entities"
	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/domain/repositories"
)

// UserDomainService provides domain-level user operations
type UserDomainService struct {
	userRepo repositories.UserRepository
	logger   *logger.Logger
}

// NewUserDomainService creates a new user domain service
func NewUserDomainService(userRepo repositories.UserRepository, log *logger.Logger) *UserDomainService {
	return &UserDomainService{
		userRepo: userRepo,
		logger:   log,
	}
}

// ValidateUserCreation validates user creation business rules
func (s *UserDomainService) ValidateUserCreation(ctx context.Context, email, password, firstName, lastName string) error {
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))
	
	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return errors.Wrap(err, errors.CodeDatabaseError, "Failed to check user existence", 500)
	}
	
	if exists {
		return errors.Conflict(fmt.Sprintf("User with email %s already exists", email))
	}
	
	// Additional business validations can be added here
	// e.g., email domain restrictions, password complexity rules, etc.
	
	return nil
}

// CanUpdateUser checks if a user can be updated by another user
func (s *UserDomainService) CanUpdateUser(ctx context.Context, targetUserID, currentUserID string, currentUserRole entities.UserRole) error {
	// Admins can update anyone
	if currentUserRole == entities.RoleAdmin {
		return nil
	}
	
	// Users can only update themselves
	if targetUserID == currentUserID {
		return nil
	}
	
	return errors.Forbidden("Insufficient permissions to update user")
}

// CanDeleteUser checks if a user can be deleted by another user
func (s *UserDomainService) CanDeleteUser(ctx context.Context, targetUserID, currentUserID string, currentUserRole entities.UserRole) error {
	// Only admins can delete users
	if currentUserRole != entities.RoleAdmin {
		return errors.Forbidden("Only administrators can delete users")
	}
	
	// Cannot delete yourself
	if targetUserID == currentUserID {
		return errors.BadRequest("You cannot delete your own account")
	}
	
	return nil
}

// CanChangeRole checks if a user's role can be changed
func (s *UserDomainService) CanChangeRole(ctx context.Context, targetUserID, currentUserID string, currentUserRole entities.UserRole, newRole entities.UserRole) error {
	// Only admins can change roles
	if currentUserRole != entities.RoleAdmin {
		return errors.Forbidden("Only administrators can change user roles")
	}
	
	// Cannot change your own role
	if targetUserID == currentUserID {
		return errors.BadRequest("You cannot change your own role")
	}
	
	// Additional business rules can be added here
	// e.g., minimum number of admin users, role hierarchy, etc.
	
	return nil
}

// ValidatePasswordChange validates password change business rules
func (s *UserDomainService) ValidatePasswordChange(ctx context.Context, userID string, currentPassword, newPassword string) (*entities.User, error) {
	// Get the user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeDatabaseError, "Failed to get user", 500)
	}
	
	if user == nil {
		return nil, errors.NotFound("User not found")
	}
	
	// Verify current password
	if !user.VerifyPassword(currentPassword) {
		s.logger.WithFields(logger.Fields{
			"user_id": userID,
			"action":  "password_change_attempt",
		}).Warn("Invalid current password provided")
		
		return nil, errors.Unauthorized("Current password is incorrect")
	}
	
	// Validate new password (additional business rules)
	if err := s.validatePasswordStrength(newPassword); err != nil {
		return nil, err
	}
	
	return user, nil
}

// validatePasswordStrength validates password strength
func (s *UserDomainService) validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.Validation("Password must be at least 8 characters long")
	}
	
	// Add more password strength rules as needed
	// e.g., uppercase, lowercase, numbers, special characters
	
	return nil
}

// ValidateEmailUpdate validates email update business rules
func (s *UserDomainService) ValidateEmailUpdate(ctx context.Context, userID, newEmail string) error {
	// Normalize email
	newEmail = strings.ToLower(strings.TrimSpace(newEmail))
	
	// Check if email is already taken by another user
	existingUser, err := s.userRepo.GetByEmail(ctx, newEmail)
	if err != nil {
		return errors.Wrap(err, errors.CodeDatabaseError, "Failed to check email availability", 500)
	}
	
	if existingUser != nil && existingUser.ID != userID {
		return errors.Conflict("Email is already taken by another user")
	}
	
	return nil
}

// GetUserStatistics retrieves user statistics
func (s *UserDomainService) GetUserStatistics(ctx context.Context) (*repositories.UserStats, error) {
	stats, err := s.userRepo.GetUserStats(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeDatabaseError, "Failed to get user statistics", 500)
	}
	
	return stats, nil
}

// ArchiveInactiveUsers archives users that have been inactive for a specified period
func (s *UserDomainService) ArchiveInactiveUsers(ctx context.Context, inactiveDays int) (int, error) {
	// This is a complex business operation that might involve multiple repositories
	// For now, we'll just return a placeholder
	
	s.logger.WithFields(logger.Fields{
		"inactive_days": inactiveDays,
		"action":        "archive_inactive_users",
	}).Info("Archiving inactive users")
	
	// Implementation would go here
	// 1. Find users inactive for X days
	// 2. Change their status to archived
	// 3. Possibly move data to archive tables
	// 4. Send notifications
	
	return 0, nil
}