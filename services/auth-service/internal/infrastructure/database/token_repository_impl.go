package database

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/dev-mayanktiwari/api-server/services/auth-service/internal/domain/entities"
	"github.com/dev-mayanktiwari/api-server/services/auth-service/internal/domain/repositories"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/database"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
)

// TokenModel represents the database model for refresh tokens
type TokenModel struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"type:text;not null;uniqueIndex"`
	ExpiresAt time.Time `gorm:"type:timestamp;not null;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name
func (TokenModel) TableName() string {
	return "refresh_tokens"
}

// ToEntity converts database model to domain entity
func (m *TokenModel) ToEntity() *entities.RefreshToken {
	return &entities.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		Token:     m.Token,
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// FromEntity converts domain entity to database model
func (m *TokenModel) FromEntity(token *entities.RefreshToken) {
	m.ID = token.ID
	m.UserID = token.UserID
	m.Token = token.Token
	m.ExpiresAt = token.ExpiresAt
	m.CreatedAt = token.CreatedAt
	m.UpdatedAt = token.UpdatedAt
}

// tokenRepository implements the TokenRepository interface
type tokenRepository struct {
	db     database.DB
	logger *logger.Logger
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db database.DB, logger *logger.Logger) repositories.TokenRepository {
	return &tokenRepository{
		db:     db,
		logger: logger,
	}
}

// StoreRefreshToken stores a refresh token
func (r *tokenRepository) StoreRefreshToken(ctx context.Context, token *entities.RefreshToken) error {
	model := &TokenModel{}
	model.FromEntity(token)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		r.logger.WithContext(ctx).WithError(err).Error("Failed to store refresh token")
		return err
	}

	// Update entity with generated ID
	token.ID = model.ID
	token.CreatedAt = model.CreatedAt
	token.UpdatedAt = model.UpdatedAt

	return nil
}

// GetRefreshToken retrieves a refresh token by user ID and token
func (r *tokenRepository) GetRefreshToken(ctx context.Context, userID, token string) (*entities.RefreshToken, error) {
	var model TokenModel

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND token = ? AND expires_at > ?", userID, token, time.Now()).
		First(&model).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.WithContext(ctx).WithError(err).Error("Failed to get refresh token")
		return nil, err
	}

	return model.ToEntity(), nil
}

// IsRefreshTokenValid checks if a refresh token is valid
func (r *tokenRepository) IsRefreshTokenValid(ctx context.Context, userID, token string) (bool, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&TokenModel{}).
		Where("user_id = ? AND token = ? AND expires_at > ?", userID, token, time.Now()).
		Count(&count).Error

	if err != nil {
		r.logger.WithContext(ctx).WithError(err).Error("Failed to validate refresh token")
		return false, err
	}

	return count > 0, nil
}

// RevokeRefreshToken revokes a refresh token
func (r *tokenRepository) RevokeRefreshToken(ctx context.Context, userID, token string) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND token = ?", userID, token).
		Delete(&TokenModel{})

	if result.Error != nil {
		r.logger.WithContext(ctx).WithError(result.Error).Error("Failed to revoke refresh token")
		return result.Error
	}

	r.logger.WithContext(ctx).WithFields(logger.Fields{
		"user_id":        userID,
		"affected_rows":  result.RowsAffected,
	}).Info("Refresh token revoked")

	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (r *tokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&TokenModel{})

	if result.Error != nil {
		r.logger.WithContext(ctx).WithError(result.Error).Error("Failed to revoke all user tokens")
		return result.Error
	}

	r.logger.WithContext(ctx).WithFields(logger.Fields{
		"user_id":       userID,
		"affected_rows": result.RowsAffected,
	}).Info("All user refresh tokens revoked")

	return nil
}

// CleanupExpiredTokens removes expired tokens
func (r *tokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	result := r.db.WithContext(ctx).
		Where("expires_at <= ?", time.Now()).
		Delete(&TokenModel{})

	if result.Error != nil {
		r.logger.WithContext(ctx).WithError(result.Error).Error("Failed to cleanup expired tokens")
		return result.Error
	}

	r.logger.WithContext(ctx).WithFields(logger.Fields{
		"affected_rows": result.RowsAffected,
	}).
		Info("Expired tokens cleaned up")

	return nil
}