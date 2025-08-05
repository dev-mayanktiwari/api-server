package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-" gorm:"not null"` // Never serialize password
	FirstName string    `json:"first_name" gorm:"not null"`
	LastName  string    `json:"last_name" gorm:"not null"`
	Role      string    `json:"role" gorm:"default:user;not null"`
	IsActive  bool      `json:"is_active" gorm:"default:true;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate is a GORM hook that runs before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Hash password if it's not already hashed
	if len(u.Password) < 60 { // bcrypt hashes are 60 characters
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

// CheckPassword verifies if the provided password matches the user's password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// SetPassword hashes and sets a new password for the user
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// ToSafeUser returns a user struct without sensitive information
func (u *User) ToSafeUser() SafeUser {
	return SafeUser{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
		IsActive:  u.IsActive,
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
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required,min=1"`
	LastName  string `json:"last_name" binding:"required,min=1"`
	Role      string `json:"role,omitempty"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Email     string `json:"email,omitempty" binding:"omitempty,email"`
	FirstName string `json:"first_name,omitempty" binding:"omitempty,min=1"`
	LastName  string `json:"last_name,omitempty" binding:"omitempty,min=1"`
	Role      string `json:"role,omitempty"`
	IsActive  *bool  `json:"is_active,omitempty"`
}

// LoginRequest represents the request payload for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the response payload for user login
type LoginResponse struct {
	User  SafeUser `json:"user"`
	Token string   `json:"token"`
}

// ChangePasswordRequest represents the request payload for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}