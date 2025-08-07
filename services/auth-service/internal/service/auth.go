package service

import (
	"errors"
	"fmt"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/model"
	"auth-service/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	config *config.Config
	logger *logger.Logger
}

func NewAuthService(config *config.Config, logger *logger.Logger) *AuthService {
	return &AuthService{
		config: config,
		logger: logger,
	}
}

func (s *AuthService) GenerateTokens(userID, email, role string) (*model.TokenPair, error) {
	accessToken, err := s.generateAccessToken(userID, email, role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.generateRefreshToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(s.config.JWT.AccessTokenDuration) * time.Minute)

	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*model.ValidateTokenResponse, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return &model.ValidateTokenResponse{
			Valid:   false,
			Message: err.Error(),
		}, nil
	}

	return &model.ValidateTokenResponse{
		Valid:  true,
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}

func (s *AuthService) RefreshTokens(refreshToken string) (*model.TokenPair, error) {
	claims, err := s.parseRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	return s.GenerateTokens(claims.UserID, claims.Email, claims.Role)
}

func (s *AuthService) generateAccessToken(userID, email, role string) (string, error) {
	now := time.Now()
	claims := &model.Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.config.JWT.AccessTokenDuration) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "auth-service",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.SecretKey))
}

func (s *AuthService) generateRefreshToken(userID string) (string, error) {
	now := time.Now()
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.config.JWT.RefreshTokenDuration) * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    "auth-service",
		Subject:   userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.SecretKey))
}

func (s *AuthService) parseToken(tokenString string) (*model.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*model.Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *AuthService) parseRefreshToken(tokenString string) (*model.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, errors.New("invalid refresh token claims")
	}

	return &model.Claims{
		UserID: claims.Subject,
	}, nil
}