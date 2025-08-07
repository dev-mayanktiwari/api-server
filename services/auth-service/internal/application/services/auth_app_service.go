package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/auth"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/errors"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/services/auth-service/internal/domain/entities"
	"github.com/dev-mayanktiwari/api-server/services/auth-service/internal/domain/repositories"
	"github.com/dev-mayanktiwari/api-server/services/auth-service/internal/application/dto"
)

// AuthApplicationService provides authentication operations
type AuthApplicationService struct {
	tokenRepo  repositories.TokenRepository
	jwtManager *auth.JWTManager
	logger     *logger.Logger
}

// NewAuthApplicationService creates a new auth application service
func NewAuthApplicationService(
	tokenRepo repositories.TokenRepository,
	jwtManager *auth.JWTManager,
	logger *logger.Logger,
) *AuthApplicationService {
	return &AuthApplicationService{
		tokenRepo:  tokenRepo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// Login authenticates a user (calls user service internally)
func (s *AuthApplicationService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// Normalize email
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// TODO: Call User Service to validate credentials
	// For now, we'll simulate this by checking a hardcoded admin user
	// In real implementation, make HTTP call to user-service
	
	var userID, email, role string
	var isValid bool
	
	// Simulate user service call
	if req.Email == "admin@api-server.com" && req.Password == "admin123" {
		userID = "admin-user-id"
		email = "admin@api-server.com"
		role = "admin"
		isValid = true
	} else if req.Email == "user@example.com" && req.Password == "password123" {
		userID = "regular-user-id"
		email = "user@example.com"
		role = "user"
		isValid = true
	}

	if !isValid {
		s.logger.WithField("email", req.Email).Warn("Invalid login attempt")
		return nil, errors.Unauthorized("Invalid email or password")
	}

	// Generate access token
	accessToken, err := s.jwtManager.GenerateToken(userID, email, role)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate access token")
		return nil, errors.Internal("Failed to generate token")
	}

	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(userID, email, role)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate refresh token")
		return nil, errors.Internal("Failed to generate refresh token")
	}

	// Store refresh token in database
	token := entities.NewRefreshToken(userID, refreshToken, time.Now().Add(7*24*time.Hour))
	if err := s.tokenRepo.StoreRefreshToken(ctx, token); err != nil {
		s.logger.WithError(err).Error("Failed to store refresh token")
		return nil, errors.Internal("Failed to store refresh token")
	}

	s.logger.WithFields(logger.Fields{
		"user_id": userID,
		"email":   email,
		"role":    role,
	}).Info("User logged in successfully")

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(24 * time.Hour / time.Second), // 24 hours in seconds
		User: &dto.UserInfo{
			UserID: userID,
			Email:  email,
			Role:   role,
		},
	}, nil
}

// RefreshToken generates a new access token using refresh token
func (s *AuthApplicationService) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		s.logger.WithError(err).Warn("Invalid refresh token")
		return nil, errors.Unauthorized("Invalid refresh token")
	}

	// Check if refresh token exists in database
	exists, err := s.tokenRepo.IsRefreshTokenValid(ctx, claims.UserID, req.RefreshToken)
	if err != nil {
		s.logger.WithError(err).Error("Failed to validate refresh token")
		return nil, errors.Internal("Failed to validate refresh token")
	}

	if !exists {
		s.logger.WithField("user_id", claims.UserID).Warn("Refresh token not found in database")
		return nil, errors.Unauthorized("Invalid refresh token")
	}

	// Generate new access token
	newAccessToken, err := s.jwtManager.GenerateToken(claims.UserID, claims.Email, claims.Role)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate new access token")
		return nil, errors.Internal("Failed to generate token")
	}

	s.logger.WithFields(logger.Fields{
		"user_id": claims.UserID,
		"email":   claims.Email,
	}).Info("Token refreshed successfully")

	return &dto.RefreshTokenResponse{
		AccessToken: newAccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(24 * time.Hour / time.Second), // 24 hours in seconds
	}, nil
}

// Logout invalidates the refresh token
func (s *AuthApplicationService) Logout(ctx context.Context, req *dto.LogoutRequest) error {
	// Validate access token to get user ID
	claims, err := s.jwtManager.ValidateToken(req.AccessToken)
	if err != nil {
		// Even if access token is invalid, we should try to logout using refresh token
		if req.RefreshToken != "" {
			refreshClaims, refreshErr := s.jwtManager.ValidateToken(req.RefreshToken)
			if refreshErr != nil {
				return errors.Unauthorized("Invalid tokens")
			}
			claims = refreshClaims
		} else {
			return errors.Unauthorized("Invalid access token")
		}
	}

	// Remove refresh token from database
	if req.RefreshToken != "" {
		if err := s.tokenRepo.RevokeRefreshToken(ctx, claims.UserID, req.RefreshToken); err != nil {
			s.logger.WithError(err).Error("Failed to revoke refresh token")
			return errors.Internal("Failed to logout")
		}
	}

	s.logger.WithField("user_id", claims.UserID).Info("User logged out successfully")
	return nil
}

// ValidateToken validates an access token
func (s *AuthApplicationService) ValidateToken(ctx context.Context, req *dto.ValidateTokenRequest) (*dto.ValidateTokenResponse, error) {
	claims, err := s.jwtManager.ValidateToken(req.Token)
	if err != nil {
		return nil, err
	}

	return &dto.ValidateTokenResponse{
		Valid: true,
		User: &dto.UserInfo{
			UserID: claims.UserID,
			Email:  claims.Email,
			Role:   claims.Role,
		},
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

// GetCurrentUser returns current user information from token
func (s *AuthApplicationService) GetCurrentUser(ctx context.Context, userID string) (*dto.UserInfo, error) {
	// TODO: In real implementation, fetch user details from user-service
	// For now, return basic info from token
	
	// This would typically make an HTTP call to user-service
	// userServiceClient := NewUserServiceClient()
	// user, err := userServiceClient.GetUser(ctx, userID)
	
	// Simulated response
	if userID == "admin-user-id" {
		return &dto.UserInfo{
			UserID:    "admin-user-id",
			Email:     "admin@api-server.com",
			Role:      "admin",
			FirstName: "System",
			LastName:  "Administrator",
		}, nil
	}

	return &dto.UserInfo{
		UserID:    userID,
		Email:     "user@example.com",
		Role:      "user",
		FirstName: "Test",
		LastName:  "User",
	}, nil
}