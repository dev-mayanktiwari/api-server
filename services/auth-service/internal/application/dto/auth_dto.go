package dto

import (
	"time"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/errors"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int64     `json:"expires_in"` // seconds
	User         *UserInfo `json:"user"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse represents a refresh token response
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"` // seconds
}

// LogoutRequest represents a logout request
type LogoutRequest struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ValidateTokenRequest represents a token validation request
type ValidateTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// ValidateTokenResponse represents a token validation response
type ValidateTokenResponse struct {
	Valid     bool      `json:"valid"`
	User      *UserInfo `json:"user,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

// UserInfo represents user information
type UserInfo struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// Validate validates the login request
func (r *LoginRequest) Validate() error {
	if r.Email == "" {
		return errors.Validation("Email is required")
	}
	if r.Password == "" {
		return errors.Validation("Password is required")
	}
	if len(r.Password) < 8 {
		return errors.Validation("Password must be at least 8 characters")
	}
	return nil
}

// Validate validates the refresh token request
func (r *RefreshTokenRequest) Validate() error {
	if r.RefreshToken == "" {
		return errors.Validation("Refresh token is required")
	}
	return nil
}

// Validate validates the token validation request
func (r *ValidateTokenRequest) Validate() error {
	if r.Token == "" {
		return errors.Validation("Token is required")
	}
	return nil
}