// Package dto contains data transfer objects for the application layer.
package dto

import (
	"time"

	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/domain/entities"
)

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email" example:"user@example.com"`
	Password  string `json:"password" binding:"required,min=8" example:"password123"`
	FirstName string `json:"first_name" binding:"required,min=1" example:"John"`
	LastName  string `json:"last_name" binding:"required,min=1" example:"Doe"`
	Role      string `json:"role,omitempty" example:"user"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Email     string `json:"email,omitempty" binding:"omitempty,email" example:"newemail@example.com"`
	FirstName string `json:"first_name,omitempty" binding:"omitempty,min=1" example:"Jane"`
	LastName  string `json:"last_name,omitempty" binding:"omitempty,min=1" example:"Smith"`
	Role      string `json:"role,omitempty" example:"admin"`
	Status    string `json:"status,omitempty" example:"active"`
}

// ChangePasswordRequest represents the request to change password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"oldpassword123"`
	NewPassword     string `json:"new_password" binding:"required,min=8" example:"newpassword123"`
}

// UserResponse represents the user response
type UserResponse struct {
	ID        string    `json:"id" example:"uuid-here"`
	Email     string    `json:"email" example:"user@example.com"`
	FirstName string    `json:"first_name" example:"John"`
	LastName  string    `json:"last_name" example:"Doe"`
	Role      string    `json:"role" example:"user"`
	Status    string    `json:"status" example:"active"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ListUsersRequest represents the request to list users
type ListUsersRequest struct {
	Page   int    `form:"page" binding:"min=1" example:"1"`
	Limit  int    `form:"limit" binding:"min=1,max=100" example:"10"`
	Role   string `form:"role" example:"user"`
	Status string `form:"status" example:"active"`
	Search string `form:"search" example:"john"`
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users      []*UserResponse `json:"users"`
	Pagination *Pagination     `json:"pagination"`
}

// Pagination represents pagination information
type Pagination struct {
	Page       int   `json:"page" example:"1"`
	Limit      int   `json:"limit" example:"10"`
	Total      int64 `json:"total" example:"100"`
	TotalPages int   `json:"total_pages" example:"10"`
}

// UserStatsResponse represents user statistics
type UserStatsResponse struct {
	TotalUsers     int64 `json:"total_users" example:"1000"`
	ActiveUsers    int64 `json:"active_users" example:"950"`
	InactiveUsers  int64 `json:"inactive_users" example:"30"`
	SuspendedUsers int64 `json:"suspended_users" example:"20"`
	AdminUsers     int64 `json:"admin_users" example:"5"`
	RegularUsers   int64 `json:"regular_users" example:"995"`
}

// Helper functions to convert between DTOs and entities

// ToEntity converts CreateUserRequest to User entity
func (r *CreateUserRequest) ToEntity() (*entities.User, error) {
	role := entities.RoleUser
	if r.Role != "" {
		role = entities.UserRole(r.Role)
	}

	user, err := entities.NewUser(r.Email, r.Password, r.FirstName, r.LastName)
	if err != nil {
		return nil, err
	}

	if r.Role != "" {
		if err := user.ChangeRole(role); err != nil {
			return nil, err
		}
	}

	return user, nil
}

// FromEntity converts User entity to UserResponse
func (r *UserResponse) FromEntity(user *entities.User) {
	r.ID = user.ID
	r.Email = user.Email
	r.FirstName = user.FirstName
	r.LastName = user.LastName
	r.Role = string(user.Role)
	r.Status = string(user.Status)
	r.CreatedAt = user.CreatedAt
	r.UpdatedAt = user.UpdatedAt
}

// FromSafeUser converts SafeUser to UserResponse
func (r *UserResponse) FromSafeUser(safeUser *entities.SafeUser) {
	r.ID = safeUser.ID
	r.Email = safeUser.Email
	r.FirstName = safeUser.FirstName
	r.LastName = safeUser.LastName
	r.Role = safeUser.Role
	r.Status = safeUser.Status
	r.CreatedAt = safeUser.CreatedAt
	r.UpdatedAt = safeUser.UpdatedAt
}

// CalculatePagination calculates pagination information
func CalculatePagination(page, limit int, total int64) *Pagination {
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	if totalPages < 1 {
		totalPages = 1
	}

	return &Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// ApplyUpdateFields applies update fields from UpdateUserRequest to User entity
func (r *UpdateUserRequest) ApplyUpdateFields(user *entities.User) error {
	if r.FirstName != "" || r.LastName != "" {
		firstName := user.FirstName
		lastName := user.LastName
		
		if r.FirstName != "" {
			firstName = r.FirstName
		}
		if r.LastName != "" {
			lastName = r.LastName
		}
		
		if err := user.UpdateProfile(firstName, lastName); err != nil {
			return err
		}
	}

	if r.Email != "" {
		if err := user.UpdateEmail(r.Email); err != nil {
			return err
		}
	}

	if r.Role != "" {
		if err := user.ChangeRole(entities.UserRole(r.Role)); err != nil {
			return err
		}
	}

	if r.Status != "" {
		if err := user.ChangeStatus(entities.UserStatus(r.Status)); err != nil {
			return err
		}
	}

	return nil
}