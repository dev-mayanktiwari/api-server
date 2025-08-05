package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager manages JWT tokens
type JWTManager struct {
	secretKey string
	tokenTTL  time.Duration
}

// Claims represents the JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, tokenTTL time.Duration) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
		tokenTTL:  tokenTTL,
	}
}

// GenerateToken generates a new JWT token
func (j *JWTManager) GenerateToken(userID, email, role string) (string, error) {
	if j.secretKey == "" {
		return "", errors.New("JWT secret key is not set")
	}

	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "api-server",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	if j.secretKey == "" {
		return nil, errors.New("JWT secret key is not set")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshToken generates a new token if the old one is valid but close to expiry
func (j *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Check if token is close to expiry (within 1 hour)
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", errors.New("token is not close to expiry")
	}

	// Generate new token with same claims
	return j.GenerateToken(claims.UserID, claims.Email, claims.Role)
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) {
		return "", errors.New("invalid authorization header format")
	}

	if authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("authorization header must start with 'Bearer '")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", errors.New("token not found in authorization header")
	}

	return token, nil
}