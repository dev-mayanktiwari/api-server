package entities

import (
	"time"
)

// RefreshToken represents a refresh token entity
type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewRefreshToken creates a new refresh token
func NewRefreshToken(userID, token string, expiresAt time.Time) *RefreshToken {
	now := time.Now()
	return &RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsExpired checks if the refresh token is expired
func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsValid checks if the refresh token is valid
func (t *RefreshToken) IsValid() bool {
	return !t.IsExpired()
}