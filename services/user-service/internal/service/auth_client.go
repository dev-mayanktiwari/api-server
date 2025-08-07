package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"user-service/internal/config"
	"user-service/pkg/logger"
)

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

type AuthClient struct {
	baseURL string
	client  *http.Client
	logger  *logger.Logger
}

func NewAuthClient(config *config.Config, logger *logger.Logger) *AuthClient {
	return &AuthClient{
		baseURL: config.Services.AuthService,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (c *AuthClient) GenerateTokens(userID, email, role string) (*TokenPair, error) {
	payload := map[string]string{
		"user_id": userID,
		"email":   email,
		"role":    role,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.client.Post(c.baseURL+"/api/v1/auth/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call auth service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status: %d", resp.StatusCode)
	}

	var response struct {
		Success bool       `json:"success"`
		Data    *TokenPair `json:"data"`
		Message string     `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("auth service error: %s", response.Message)
	}

	return response.Data, nil
}