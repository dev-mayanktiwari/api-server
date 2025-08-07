// Package database contains database implementations of repository interfaces.
package database

import (
	"context"
	"errors"
	"time"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/database"
	sharedErrors "github.com/dev-mayanktiwari/api-server/shared/pkg/errors"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/domain/entities"
	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/domain/repositories"
	"gorm.io/gorm"
)

// UserModel represents the user database model
type UserModel struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Password  string    `gorm:"not null"`
	FirstName string    `gorm:"not null"`
	LastName  string    `gorm:"not null"`
	Role      string    `gorm:"default:user;not null"`
	Status    string    `gorm:"default:active;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName returns the table name for the UserModel
func (UserModel) TableName() string {
	return "users"
}

// ToEntity converts UserModel to domain entity
func (m *UserModel) ToEntity() *entities.User {
	return &entities.User{
		ID:        m.ID,
		Email:     m.Email,
		Password:  m.Password,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		Role:      entities.UserRole(m.Role),
		Status:    entities.UserStatus(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// FromEntity converts domain entity to UserModel
func (m *UserModel) FromEntity(user *entities.User) {
	m.ID = user.ID
	m.Email = user.Email
	m.Password = user.Password
	m.FirstName = user.FirstName
	m.LastName = user.LastName
	m.Role = string(user.Role)
	m.Status = string(user.Status)
	m.CreatedAt = user.CreatedAt
	m.UpdatedAt = user.UpdatedAt
}

// UserRepositoryImpl implements the UserRepository interface using GORM
type UserRepositoryImpl struct {
	db     *database.DB
	logger *logger.Logger
}

// NewUserRepository creates a new user repository implementation
func NewUserRepository(db *database.DB, log *logger.Logger) repositories.UserRepository {
	return &UserRepositoryImpl{
		db:     db,
		logger: log,
	}
}

// Create creates a new user
func (r *UserRepositoryImpl) Create(ctx context.Context, user *entities.User) error {
	start := time.Now()
	defer func() {
		r.logger.LogDatabaseQuery("INSERT INTO users", time.Since(start), nil)
	}()

	model := &UserModel{}
	model.FromEntity(user)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.LogDatabaseQuery("INSERT INTO users", time.Since(start), err)
		return sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to create user", 500)
	}

	// Update the user ID from the database
	user.ID = model.ID
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id string) (*entities.User, error) {
	start := time.Now()
	
	var model UserModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	r.logger.LogDatabaseQuery("SELECT * FROM users WHERE id = ?", time.Since(start), err)
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil for not found
		}
		return nil, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to get user by ID", 500)
	}

	return model.ToEntity(), nil
}

// GetByEmail retrieves a user by email
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	start := time.Now()
	
	var model UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
	r.logger.LogDatabaseQuery("SELECT * FROM users WHERE email = ?", time.Since(start), err)
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil for not found
		}
		return nil, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to get user by email", 500)
	}

	return model.ToEntity(), nil
}

// Update updates an existing user
func (r *UserRepositoryImpl) Update(ctx context.Context, user *entities.User) error {
	start := time.Now()
	
	model := &UserModel{}
	model.FromEntity(user)
	model.UpdatedAt = time.Now().UTC()

	err := r.db.WithContext(ctx).Where("id = ?", user.ID).Updates(model).Error
	r.logger.LogDatabaseQuery("UPDATE users SET ... WHERE id = ?", time.Since(start), err)
	
	if err != nil {
		return sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to update user", 500)
	}

	return nil
}

// Delete soft deletes a user
func (r *UserRepositoryImpl) Delete(ctx context.Context, id string) error {
	start := time.Now()
	
	err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&UserModel{}).Error
	r.logger.LogDatabaseQuery("UPDATE users SET deleted_at = ? WHERE id = ?", time.Since(start), err)
	
	if err != nil {
		return sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to delete user", 500)
	}

	return nil
}

// List retrieves users with pagination
func (r *UserRepositoryImpl) List(ctx context.Context, offset, limit int) ([]*entities.User, int64, error) {
	start := time.Now()
	
	var models []UserModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Count(&total).Error; err != nil {
		r.logger.LogDatabaseQuery("SELECT COUNT(*) FROM users", time.Since(start), err)
		return nil, 0, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to count users", 500)
	}

	// Get paginated results
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&models).Error
	r.logger.LogDatabaseQuery("SELECT * FROM users LIMIT ? OFFSET ?", time.Since(start), err)
	
	if err != nil {
		return nil, 0, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to list users", 500)
	}

	// Convert to entities
	users := make([]*entities.User, len(models))
	for i, model := range models {
		users[i] = model.ToEntity()
	}

	return users, total, nil
}

// ExistsByEmail checks if a user exists by email
func (r *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	start := time.Now()
	
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Where("email = ?", email).Count(&count).Error
	r.logger.LogDatabaseQuery("SELECT COUNT(*) FROM users WHERE email = ?", time.Since(start), err)
	
	if err != nil {
		return false, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to check user existence", 500)
	}

	return count > 0, nil
}

// ExistsByID checks if a user exists by ID
func (r *UserRepositoryImpl) ExistsByID(ctx context.Context, id string) (bool, error) {
	start := time.Now()
	
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Count(&count).Error
	r.logger.LogDatabaseQuery("SELECT COUNT(*) FROM users WHERE id = ?", time.Since(start), err)
	
	if err != nil {
		return false, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to check user existence", 500)
	}

	return count > 0, nil
}

// UpdatePassword updates only the user's password
func (r *UserRepositoryImpl) UpdatePassword(ctx context.Context, id, hashedPassword string) error {
	start := time.Now()
	
	err := r.db.WithContext(ctx).Model(&UserModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"password":   hashedPassword,
			"updated_at": time.Now().UTC(),
		}).Error
	
	r.logger.LogDatabaseQuery("UPDATE users SET password = ?, updated_at = ? WHERE id = ?", time.Since(start), err)
	
	if err != nil {
		return sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to update password", 500)
	}

	return nil
}

// GetActiveUsers retrieves all active users
func (r *UserRepositoryImpl) GetActiveUsers(ctx context.Context, offset, limit int) ([]*entities.User, int64, error) {
	start := time.Now()
	
	var models []UserModel
	var total int64

	// Get total count of active users
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Where("status = ?", "active").Count(&total).Error; err != nil {
		r.logger.LogDatabaseQuery("SELECT COUNT(*) FROM users WHERE status = 'active'", time.Since(start), err)
		return nil, 0, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to count active users", 500)
	}

	// Get paginated results
	err := r.db.WithContext(ctx).Where("status = ?", "active").Offset(offset).Limit(limit).Find(&models).Error
	r.logger.LogDatabaseQuery("SELECT * FROM users WHERE status = 'active' LIMIT ? OFFSET ?", time.Since(start), err)
	
	if err != nil {
		return nil, 0, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to list active users", 500)
	}

	// Convert to entities
	users := make([]*entities.User, len(models))
	for i, model := range models {
		users[i] = model.ToEntity()
	}

	return users, total, nil
}

// GetUsersByRole retrieves users by role
func (r *UserRepositoryImpl) GetUsersByRole(ctx context.Context, role entities.UserRole, offset, limit int) ([]*entities.User, int64, error) {
	start := time.Now()
	
	var models []UserModel
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Where("role = ?", string(role)).Count(&total).Error; err != nil {
		r.logger.LogDatabaseQuery("SELECT COUNT(*) FROM users WHERE role = ?", time.Since(start), err)
		return nil, 0, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to count users by role", 500)
	}

	// Get paginated results
	err := r.db.WithContext(ctx).Where("role = ?", string(role)).Offset(offset).Limit(limit).Find(&models).Error
	r.logger.LogDatabaseQuery("SELECT * FROM users WHERE role = ? LIMIT ? OFFSET ?", time.Since(start), err)
	
	if err != nil {
		return nil, 0, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to list users by role", 500)
	}

	// Convert to entities
	users := make([]*entities.User, len(models))
	for i, model := range models {
		users[i] = model.ToEntity()
	}

	return users, total, nil
}

// GetUserStats returns user statistics
func (r *UserRepositoryImpl) GetUserStats(ctx context.Context) (*repositories.UserStats, error) {
	start := time.Now()
	
	stats := &repositories.UserStats{}

	// Total users
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Count(&stats.TotalUsers).Error; err != nil {
		r.logger.LogDatabaseQuery("SELECT COUNT(*) FROM users", time.Since(start), err)
		return nil, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to count total users", 500)
	}

	// Active users
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Where("status = ?", "active").Count(&stats.ActiveUsers).Error; err != nil {
		return nil, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to count active users", 500)
	}

	// Inactive users
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Where("status = ?", "inactive").Count(&stats.InactiveUsers).Error; err != nil {
		return nil, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to count inactive users", 500)
	}

	// Suspended users
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Where("status = ?", "suspended").Count(&stats.SuspendedUsers).Error; err != nil {
		return nil, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to count suspended users", 500)
	}

	// Admin users
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Where("role = ?", "admin").Count(&stats.AdminUsers).Error; err != nil {
		return nil, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to count admin users", 500)
	}

	// Regular users
	if err := r.db.WithContext(ctx).Model(&UserModel{}).Where("role = ?", "user").Count(&stats.RegularUsers).Error; err != nil {
		r.logger.LogDatabaseQuery("User statistics query completed", time.Since(start), err)
		return nil, sharedErrors.Wrap(err, sharedErrors.CodeDatabaseError, "Failed to count regular users", 500)
	}

	r.logger.LogDatabaseQuery("User statistics query completed", time.Since(start), nil)
	return stats, nil
}