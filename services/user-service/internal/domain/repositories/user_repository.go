// Package repositories contains repository interfaces for the user domain.
// These define contracts for data access without implementation details.
package repositories

import (
	"context"

	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/domain/entities"
)

// UserRepository defines the contract for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entities.User) error
	
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id string) (*entities.User, error)
	
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	
	// Update updates an existing user
	Update(ctx context.Context, user *entities.User) error
	
	// Delete soft deletes a user
	Delete(ctx context.Context, id string) error
	
	// List retrieves users with pagination
	List(ctx context.Context, offset, limit int) ([]*entities.User, int64, error)
	
	// ExistsByEmail checks if a user exists by email
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	
	// ExistsByID checks if a user exists by ID
	ExistsByID(ctx context.Context, id string) (bool, error)
	
	// UpdatePassword updates only the user's password
	UpdatePassword(ctx context.Context, id, hashedPassword string) error
	
	// GetActiveUsers retrieves all active users
	GetActiveUsers(ctx context.Context, offset, limit int) ([]*entities.User, int64, error)
	
	// GetUsersByRole retrieves users by role
	GetUsersByRole(ctx context.Context, role entities.UserRole, offset, limit int) ([]*entities.User, int64, error)
	
	// GetUserStats returns user statistics
	GetUserStats(ctx context.Context) (*UserStats, error)
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers     int64 `json:"total_users"`
	ActiveUsers    int64 `json:"active_users"`
	InactiveUsers  int64 `json:"inactive_users"`
	SuspendedUsers int64 `json:"suspended_users"`
	AdminUsers     int64 `json:"admin_users"`
	RegularUsers   int64 `json:"regular_users"`
}

// SearchCriteria represents search criteria for users
type SearchCriteria struct {
	Email     string              `json:"email,omitempty"`
	Role      entities.UserRole   `json:"role,omitempty"`
	Status    entities.UserStatus `json:"status,omitempty"`
	CreatedAfter  *time.Time     `json:"created_after,omitempty"`
	CreatedBefore *time.Time     `json:"created_before,omitempty"`
}