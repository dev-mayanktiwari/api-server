package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for our application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	CORS     CORSConfig     `mapstructure:"cors"`
}

// ServerConfig holds server related configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

// DatabaseConfig holds database related configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

// JWTConfig holds JWT related configuration
type JWTConfig struct {
	Secret string        `mapstructure:"secret"`
	Expiry time.Duration `mapstructure:"expiry"`
}

// LoggerConfig holds logger related configuration
type LoggerConfig struct {
	Level      string `mapstructure:"level"`       // debug, info, warn, error
	Format     string `mapstructure:"format"`      // json, json-pretty, console
	DisableGin bool   `mapstructure:"disable_gin"` // disable gin debug logs
}

// CORSConfig holds CORS related configuration
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

// Load reads configuration from file and environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (for development)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading it: %v", err)
	}

	// Create a new viper instance
	v := viper.New()

	// Set default values
	setDefaults(v)

	// Set up viper to read from environment variables
	v.AutomaticEnv()

	// Set prefix for environment variables
	v.SetEnvPrefix("APP")

	// Replace dots and dashes with underscores in env vars
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Try to read from config file (optional)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")

	// Read config file (it's okay if it doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for configuration
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.mode", "debug")

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.name", "api_server")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.ssl_mode", "disable")

	// JWT defaults
	v.SetDefault("jwt.secret", "change-this-in-production")
	v.SetDefault("jwt.expiry", "24h")

	// Logger defaults
	v.SetDefault("logger.level", "debug")
	v.SetDefault("logger.format", "console")
	v.SetDefault("logger.disable_gin", false)

	// CORS defaults
	v.SetDefault("cors.allowed_origins", []string{"*"})
	v.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allowed_headers", []string{"Content-Type", "Authorization"})
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate server port
	if config.Server.Port == "" {
		return fmt.Errorf("server port cannot be empty")
	}

	// Validate server mode
	validServerModes := map[string]bool{
		"debug": true, "release": true, "test": true,
	}
	if !validServerModes[config.Server.Mode] {
		return fmt.Errorf("invalid server mode: %s (valid options: debug, release, test)", config.Server.Mode)
	}

	// Validate database configuration
	if config.Database.Host == "" {
		return fmt.Errorf("database host cannot be empty")
	}
	if config.Database.Name == "" {
		return fmt.Errorf("database name cannot be empty")
	}

	// Validate JWT secret
	if config.JWT.Secret == "" || config.JWT.Secret == "change-this-in-production" {
		return fmt.Errorf("JWT secret must be set and not be the default value")
	}

	// Validate logger level
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLogLevels[config.Logger.Level] {
		return fmt.Errorf("invalid log level: %s (valid options: debug, info, warn, error)", config.Logger.Level)
	}

	// Validate logger format
	validLogFormats := map[string]bool{
		"console": true, "json": true, "json-pretty": true,
	}
	if !validLogFormats[config.Logger.Format] {
		return fmt.Errorf("invalid log format: %s (valid options: console, json, json-pretty)", config.Logger.Format)
	}

	return nil
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// IsDevelopment returns true if we're in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Mode == "debug"
}

// IsProduction returns true if we're in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Mode == "release"
}