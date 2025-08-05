package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"github.com/dev-mayanktiwari/api-server/internal/model"
	"github.com/dev-mayanktiwari/api-server/pkg/logger"
)

// UserRepository handles user data operations
type UserRepository struct {
	db     *gorm.DB
	logger *logger.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB, logger *logger.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new user
func (r *UserRepository) Create(user *model.User) error {
	if err := r.db.Create(user).Error; err != nil {
		r.logger.LogError("Failed to create user", err)
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	r.logger.WithFields(map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("User created successfully")
	
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id string) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", id).First(&user).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.LogError("Failed to get user by ID", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		r.logger.LogError("Failed to get user by email", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *model.User) error {
	if err := r.db.Save(user).Error; err != nil {
		r.logger.LogError("Failed to update user", err)
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	r.logger.WithFields(map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
	}).Info("User updated successfully")
	
	return nil
}

// Delete soft deletes a user
func (r *UserRepository) Delete(id string) error {
	result := r.db.Where("id = ?", id).Delete(&model.User{})
	
	if result.Error != nil {
		r.logger.LogError("Failed to delete user", result.Error)
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	r.logger.WithField("user_id", id).Info("User deleted successfully")
	return nil
}

// List retrieves users with pagination
func (r *UserRepository) List(offset, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	
	// Get total count
	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		r.logger.LogError("Failed to count users", err)
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}
	
	// Get users with pagination
	if err := r.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		r.logger.LogError("Failed to list users", err)
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	
	return users, total, nil
}

// ExistsByEmail checks if a user exists with the given email
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	
	if err != nil {
		r.logger.LogError("Failed to check user existence by email", err)
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	
	return count > 0, nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(userID, newPassword string) error {
	user := &model.User{ID: userID}
	if err := user.SetPassword(newPassword); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	
	result := r.db.Model(&model.User{}).Where("id = ?", userID).Update("password", user.Password)
	
	if result.Error != nil {
		r.logger.LogError("Failed to update user password", result.Error)
		return fmt.Errorf("failed to update password: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	r.logger.WithField("user_id", userID).Info("User password updated successfully")
	return nil
}

// GetActiveUsers retrieves all active users
func (r *UserRepository) GetActiveUsers() ([]model.User, error) {
	var users []model.User
	
	if err := r.db.Where("is_active = ?", true).Find(&users).Error; err != nil {
		r.logger.LogError("Failed to get active users", err)
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	
	return users, nil
}

// SetUserStatus updates a user's active status
func (r *UserRepository) SetUserStatus(userID string, isActive bool) error {
	result := r.db.Model(&model.User{}).Where("id = ?", userID).Update("is_active", isActive)
	
	if result.Error != nil {
		r.logger.LogError("Failed to update user status", result.Error)
		return fmt.Errorf("failed to update user status: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	
	status := "deactivated"
	if isActive {
		status = "activated"
	}
	
	r.logger.WithFields(map[string]interface{}{
		"user_id": userID,
		"status":  status,
	}).Info("User status updated successfully")
	
	return nil
}