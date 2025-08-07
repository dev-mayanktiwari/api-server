package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set test environment variables
	os.Setenv("APP_SERVER_PORT", "9090")
	os.Setenv("APP_DATABASE_NAME", "test_db")
	os.Setenv("APP_JWT_SECRET", "test-secret")

	defer func() {
		// Clean up
		os.Unsetenv("APP_SERVER_PORT")
		os.Unsetenv("APP_DATABASE_NAME")
		os.Unsetenv("APP_JWT_SECRET")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test if environment variables override defaults
	if config.Server.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", config.Server.Port)
	}

	if config.Database.Name != "test_db" {
		t.Errorf("Expected database name test_db, got %s", config.Database.Name)
	}

	// Test helper methods
	dsn := config.GetDatabaseDSN()
	if dsn == "" {
		t.Error("Database DSN should not be empty")
	}

	address := config.GetServerAddress()
	expected := "localhost:9090"
	if address != expected {
		t.Errorf("Expected address %s, got %s", expected, address)
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid config",
			config: Config{
				Server:   ServerConfig{Port: "8080", Mode: "debug"},
				Database: DatabaseConfig{Host: "localhost", Name: "test"},
				JWT:      JWTConfig{Secret: "valid-secret"},
				Logger:   LoggerConfig{Level: "info", Format: "console"},
			},
			expectError: false,
		},
		{
			name: "invalid JWT secret",
			config: Config{
				Server:   ServerConfig{Port: "8080", Mode: "debug"},
				Database: DatabaseConfig{Host: "localhost", Name: "test"},
				JWT:      JWTConfig{Secret: "change-this-in-production"},
				Logger:   LoggerConfig{Level: "info", Format: "console"},
			},
			expectError: true,
		},
		{
			name: "invalid log level",
			config: Config{
				Server:   ServerConfig{Port: "8080", Mode: "debug"},
				Database: DatabaseConfig{Host: "localhost", Name: "test"},
				JWT:      JWTConfig{Secret: "valid-secret"},
				Logger:   LoggerConfig{Level: "invalid", Format: "console"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
