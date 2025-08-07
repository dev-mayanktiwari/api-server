package service

import (
	"fmt"
	"strings"

	"github.com/dev-mayanktiwari/api-server/internal/model"
	"github.com/dev-mayanktiwari/api-server/internal/repository"
	"github.com/dev-mayanktiwari/api-server/pkg/auth"
	"github.com/dev-mayanktiwari/api-server/pkg/logger"
)

// UserService handles user business logic
type UserService struct {
	userRepo   *repository.UserRepository
	jwtManager *auth.JWTManager
	logger     *logger.Logger
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository, jwtManager *auth.JWTManager, logger *logger.Logger) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(req *model.CreateUserRequest) (*model.SafeUser, error) {
	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	
	if exists {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}
	
	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = "user"
	}
	
	// Validate role
	if !isValidRole(role) {
		return nil, fmt.Errorf("invalid role: %s", role)
	}
	
	// Create user model
	user := &model.User{
		Email:     strings.ToLower(req.Email),
		Password:  req.Password, // Will be hashed in BeforeCreate hook
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      role,
		IsActive:  true,
	}
	
	// Create user in database
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	s.logger.LogUserAction("system", "create_user", "user", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
	})
	
	safeUser := user.ToSafeUser()
	return &safeUser, nil
}

// Login authenticates a user and returns a JWT token
func (s *UserService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(strings.ToLower(req.Email))
	if err != nil {
		s.logger.WithFields(map[string]interface{}{
			"email": req.Email,
		}).Warn("Login attempt with non-existent email")
		return nil, fmt.Errorf("invalid email or password")
	}
	
	// Check if user is active
	if !user.IsActive {
		s.logger.WithFields(map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		}).Warn("Login attempt by inactive user")
		return nil, fmt.Errorf("account is deactivated")
	}
	
	// Check password
	if !user.CheckPassword(req.Password) {
		s.logger.WithFields(map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		}).Warn("Login attempt with incorrect password")
		return nil, fmt.Errorf("invalid email or password")
	}
	
	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}
	
	s.logger.LogUserAction(user.ID, "login", "user", map[string]interface{}{
		"email": user.Email,
	})
	
	safeUser := user.ToSafeUser()
	return &model.LoginResponse{
		User:  safeUser,
		Token: token,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(userID string) (*model.SafeUser, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	safeUser := user.ToSafeUser()
	return &safeUser, nil
}

// UpdateUser updates a user's information
func (s *UserService) UpdateUser(userID string, req *model.UpdateUserRequest, currentUserID string, currentUserRole string) (*model.SafeUser, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// Check permissions
	if !s.canUpdateUser(userID, currentUserID, currentUserRole) {
		return nil, fmt.Errorf("insufficient permissions to update user")
	}
	
	// Update fields if provided
	if req.Email != "" {
		// Check if email is already taken by another user
		existingUser, err := s.userRepo.GetByEmail(strings.ToLower(req.Email))
		if err == nil && existingUser.ID != userID {
			return nil, fmt.Errorf("email is already taken")
		}
		user.Email = strings.ToLower(req.Email)
	}
	
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	
	// Only admins can update role and status
	if currentUserRole == "admin" {
		if req.Role != "" {
			if !isValidRole(req.Role) {
				return nil, fmt.Errorf("invalid role: %s", req.Role)
			}
			user.Role = req.Role
		}
		
		if req.IsActive != nil {
			user.IsActive = *req.IsActive
		}
	}
	
	// Update user in database
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	s.logger.LogUserAction(currentUserID, "update_user", "user", map[string]interface{}{
		"target_user_id": userID,
		"email":          user.Email,
	})
	
	safeUser := user.ToSafeUser()
	return &safeUser, nil
}

// ChangePassword changes a user's password
func (s *UserService) ChangePassword(userID string, req *model.ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	
	// Check current password
	if !user.CheckPassword(req.CurrentPassword) {
		return fmt.Errorf("current password is incorrect")
	}
	
	// Update password
	if err := s.userRepo.UpdatePassword(userID, req.NewPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	
	s.logger.LogUserAction(userID, "change_password", "user", map[string]interface{}{
		"user_id": userID,
	})
	
	return nil
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(userID string, currentUserID string, currentUserRole string) error {
	// Check permissions
	if !s.canDeleteUser(userID, currentUserID, currentUserRole) {
		return fmt.Errorf("insufficient permissions to delete user")
	}
	
	// Cannot delete yourself
	if userID == currentUserID {
		return fmt.Errorf("cannot delete your own account")
	}
	
	if err := s.userRepo.Delete(userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	s.logger.LogUserAction(currentUserID, "delete_user", "user", map[string]interface{}{
		"target_user_id": userID,
	})
	
	return nil
}

// ListUsers retrieves users with pagination
func (s *UserService) ListUsers(page, limit int, currentUserRole string) ([]model.SafeUser, int64, error) {
	// Only admins can list all users
	if currentUserRole != "admin" {
		return nil, 0, fmt.Errorf("insufficient permissions to list users")
	}
	
	offset := (page - 1) * limit
	users, total, err := s.userRepo.List(offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	
	// Convert to safe users
	safeUsers := make([]model.SafeUser, len(users))
	for i, user := range users {
		safeUsers[i] = user.ToSafeUser()
	}
	
	return safeUsers, total, nil
}

// canUpdateUser checks if the current user can update the target user
func (s *UserService) canUpdateUser(targetUserID, currentUserID, currentUserRole string) bool {
	// Admins can update anyone
	if currentUserRole == "admin" {
		return true
	}
	
	// Users can only update themselves
	return targetUserID == currentUserID
}

// canDeleteUser checks if the current user can delete the target user
func (s *UserService) canDeleteUser(targetUserID, currentUserID, currentUserRole string) bool {
	// Only admins can delete users
	return currentUserRole == "admin"
}

// isValidRole checks if the role is valid
func isValidRole(role string) bool {
	validRoles := map[string]bool{
		"user":  true,
		"admin": true,
	}
	return validRoles[role]
}