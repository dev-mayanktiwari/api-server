// Package entities contains domain entities for the user service.
// These represent core business objects with business logic.
package entities

import (
	"time"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/auth"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/errors"
	"github.com/google/uuid"
)

// User represents the user domain entity
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never serialize password
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      UserRole  `json:"role"`
	Status    UserStatus `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRole represents user roles in the system
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// UserStatus represents user account status
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusSuspended UserStatus = "suspended"
)

// NewUser creates a new user with default values
func NewUser(email, password, firstName, lastName string) (*User, error) {
	if email == "" {
		return nil, errors.Validation("Email is required")
	}
	
	if password == "" {
		return nil, errors.Validation("Password is required")
	}
	
	if len(password) < 8 {
		return nil, errors.Validation("Password must be at least 8 characters long")
	}
	
	if firstName == "" {
		return nil, errors.Validation("First name is required")
	}
	
	if lastName == "" {
		return nil, errors.Validation("Last name is required")
	}
	
	// Hash the password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, errors.Wrap(err, errors.CodeInternal, "Failed to hash password", 500)
	}
	
	return &User{
		ID:        uuid.New().String(),
		Email:     email,
		Password:  hashedPassword,
		FirstName: firstName,
		LastName:  lastName,
		Role:      RoleUser,
		Status:    StatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}

// ChangePassword changes the user's password
func (u *User) ChangePassword(currentPassword, newPassword string) error {
	// Verify current password
	if err := auth.VerifyPassword(u.Password, currentPassword); err != nil {
		return errors.Unauthorized("Current password is incorrect")
	}
	
	// Validate new password
	if len(newPassword) < 8 {
		return errors.Validation("New password must be at least 8 characters long")
	}
	
	// Hash new password
	hashedPassword, err := auth.HashPassword(newPassword)
	if err != nil {
		return errors.Wrap(err, errors.CodeInternal, "Failed to hash new password", 500)
	}
	
	u.Password = hashedPassword
	u.UpdatedAt = time.Now().UTC()
	
	return nil
}

// UpdateProfile updates user profile information
func (u *User) UpdateProfile(firstName, lastName string) error {
	if firstName == "" {
		return errors.Validation("First name is required")
	}
	
	if lastName == "" {
		return errors.Validation("Last name is required")
	}
	
	u.FirstName = firstName
	u.LastName = lastName
	u.UpdatedAt = time.Now().UTC()
	
	return nil
}

// UpdateEmail updates the user's email
func (u *User) UpdateEmail(email string) error {
	if email == "" {
		return errors.Validation("Email is required")
	}
	
	u.Email = email
	u.UpdatedAt = time.Now().UTC()
	
	return nil
}

// ChangeRole changes the user's role (admin operation)
func (u *User) ChangeRole(role UserRole) error {
	if role != RoleUser && role != RoleAdmin {
		return errors.Validation("Invalid role")
	}
	
	u.Role = role
	u.UpdatedAt = time.Now().UTC()
	
	return nil
}

// ChangeStatus changes the user's status (admin operation)
func (u *User) ChangeStatus(status UserStatus) error {
	if status != StatusActive && status != StatusInactive && status != StatusSuspended {
		return errors.Validation("Invalid status")
	}
	
	u.Status = status
	u.UpdatedAt = time.Now().UTC()
	
	return nil
}

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsActive returns true if the user account is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// VerifyPassword verifies if the provided password matches the user's password
func (u *User) VerifyPassword(password string) bool {
	return auth.VerifyPassword(u.Password, password) == nil
}

// ToSafeUser returns a user struct without sensitive information
func (u *User) ToSafeUser() *SafeUser {
	return &SafeUser{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      string(u.Role),
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// SafeUser represents user data safe for API responses
type SafeUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Domain events

// UserCreatedEvent represents a user creation event
type UserCreatedEvent struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// UserUpdatedEvent represents a user update event  
type UserUpdatedEvent struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserDeletedEvent represents a user deletion event
type UserDeletedEvent struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	DeletedAt time.Time `json:"deleted_at"`
}