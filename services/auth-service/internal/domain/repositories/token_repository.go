package repositories

import (
	"context"

	"github.com/dev-mayanktiwari/api-server/services/auth-service/internal/domain/entities"
)

// TokenRepository defines the interface for token storage operations
type TokenRepository interface {
	// StoreRefreshToken stores a refresh token
	StoreRefreshToken(ctx context.Context, token *entities.RefreshToken) error
	
	// GetRefreshToken retrieves a refresh token by user ID and token
	GetRefreshToken(ctx context.Context, userID, token string) (*entities.RefreshToken, error)
	
	// IsRefreshTokenValid checks if a refresh token is valid
	IsRefreshTokenValid(ctx context.Context, userID, token string) (bool, error)
	
	// RevokeRefreshToken revokes a refresh token
	RevokeRefreshToken(ctx context.Context, userID, token string) error
	
	// RevokeAllUserTokens revokes all refresh tokens for a user
	RevokeAllUserTokens(ctx context.Context, userID string) error
	
	// CleanupExpiredTokens removes expired tokens
	CleanupExpiredTokens(ctx context.Context) error
}