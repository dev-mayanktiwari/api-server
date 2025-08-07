// Package config provides centralized configuration management for the API server.
// It supports multiple configuration sources (files, environment variables, command-line flags)
// and provides validation and hot-reload capabilities.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Manager handles configuration loading and management
type Manager struct {
	viper    *viper.Viper
	watchers []func(config interface{})
	mutex    sync.RWMutex
}

// Options configures the configuration manager
type Options struct {
	// ConfigName is the name of the configuration file (without extension)
	ConfigName string
	
	// ConfigPaths are the paths to search for configuration files
	ConfigPaths []string
	
	// ConfigType is the type of configuration file (yaml, json, toml, etc.)
	ConfigType string
	
	// EnvPrefix is the prefix for environment variables
	EnvPrefix string
	
	// DefaultConfig is the default configuration to merge with loaded config
	DefaultConfig interface{}
}

// DefaultOptions returns default configuration options
func DefaultOptions() *Options {
	return &Options{
		ConfigName:  "config",
		ConfigPaths: []string{".", "./config", "/etc/app"},
		ConfigType:  "yaml",
		EnvPrefix:   "APP",
	}
}

// New creates a new configuration manager
func New(opts *Options) *Manager {
	if opts == nil {
		opts = DefaultOptions()
	}

	v := viper.New()
	
	// Set configuration file options
	v.SetConfigName(opts.ConfigName)
	v.SetConfigType(opts.ConfigType)
	
	// Add config paths
	for _, path := range opts.ConfigPaths {
		v.AddConfigPath(path)
	}
	
	// Setup environment variable handling
	if opts.EnvPrefix != "" {
		v.SetEnvPrefix(opts.EnvPrefix)
		v.AutomaticEnv()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	}
	
	return &Manager{
		viper:    v,
		watchers: make([]func(config interface{}), 0),
	}
}

// Load loads configuration from all configured sources
func (m *Manager) Load(config interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Try to read config file
	if err := m.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is acceptable, we'll use defaults and env vars
	}
	
	// Unmarshal into the provided struct
	if err := m.viper.Unmarshal(config); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	return nil
}

// LoadWithDefaults loads configuration with default values
func (m *Manager) LoadWithDefaults(config interface{}, defaults interface{}) error {
	// First, apply defaults
	if defaults != nil {
		if err := m.viper.MergeConfigMap(structToMap(defaults)); err != nil {
			return fmt.Errorf("error applying defaults: %w", err)
		}
	}
	
	return m.Load(config)
}

// Watch starts watching for configuration changes and calls the callback when changes occur
func (m *Manager) Watch(config interface{}, callback func(config interface{})) error {
	m.mutex.Lock()
	m.watchers = append(m.watchers, callback)
	m.mutex.Unlock()
	
	m.viper.WatchConfig()
	m.viper.OnConfigChange(func(e fsnotify.Event) {
		m.mutex.RLock()
		defer m.mutex.RUnlock()
		
		// Reload configuration
		if err := m.viper.Unmarshal(config); err != nil {
			// Log error but don't crash
			fmt.Printf("Error reloading config: %v\n", err)
			return
		}
		
		// Notify all watchers
		for _, watcher := range m.watchers {
			go watcher(config)
		}
	})
	
	return nil
}

// Get retrieves a configuration value by key
func (m *Manager) Get(key string) interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.viper.Get(key)
}

// GetString retrieves a string configuration value
func (m *Manager) GetString(key string) string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.viper.GetString(key)
}

// GetInt retrieves an integer configuration value
func (m *Manager) GetInt(key string) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.viper.GetInt(key)
}

// GetBool retrieves a boolean configuration value
func (m *Manager) GetBool(key string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.viper.GetBool(key)
}

// GetDuration retrieves a duration configuration value
func (m *Manager) GetDuration(key string) time.Duration {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.viper.GetDuration(key)
}

// Set sets a configuration value
func (m *Manager) Set(key string, value interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.viper.Set(key, value)
}

// WriteConfig writes the current configuration to file
func (m *Manager) WriteConfig() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.viper.WriteConfig()
}

// Common configuration structures

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host" json:"host" yaml:"host"`
	Port            int           `mapstructure:"port" json:"port" yaml:"port"`
	Username        string        `mapstructure:"username" json:"username" yaml:"username"`
	Password        string        `mapstructure:"password" json:"-" yaml:"password"` // Hidden in JSON
	Database        string        `mapstructure:"database" json:"database" yaml:"database"`
	SSLMode         string        `mapstructure:"ssl_mode" json:"ssl_mode" yaml:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime" json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time" json:"conn_max_idle_time" yaml:"conn_max_idle_time"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Host            string        `mapstructure:"host" json:"host" yaml:"host"`
	Port            int           `mapstructure:"port" json:"port" yaml:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" json:"write_timeout" yaml:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout" json:"idle_timeout" yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" json:"shutdown_timeout" yaml:"shutdown_timeout"`
	TLSCertFile     string        `mapstructure:"tls_cert_file" json:"tls_cert_file" yaml:"tls_cert_file"`
	TLSKeyFile      string        `mapstructure:"tls_key_file" json:"tls_key_file" yaml:"tls_key_file"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret          string        `mapstructure:"secret" json:"-" yaml:"secret"` // Hidden in JSON
	Issuer          string        `mapstructure:"issuer" json:"issuer" yaml:"issuer"`
	ExpirationTime  time.Duration `mapstructure:"expiration_time" json:"expiration_time" yaml:"expiration_time"`
	RefreshTime     time.Duration `mapstructure:"refresh_time" json:"refresh_time" yaml:"refresh_time"`
	Algorithm       string        `mapstructure:"algorithm" json:"algorithm" yaml:"algorithm"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64       `mapstructure:"requests_per_second" json:"requests_per_second" yaml:"requests_per_second"`
	BurstSize         int           `mapstructure:"burst_size" json:"burst_size" yaml:"burst_size"`
	CleanupInterval   time.Duration `mapstructure:"cleanup_interval" json:"cleanup_interval" yaml:"cleanup_interval"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowOrigins     []string      `mapstructure:"allow_origins" json:"allow_origins" yaml:"allow_origins"`
	AllowMethods     []string      `mapstructure:"allow_methods" json:"allow_methods" yaml:"allow_methods"`
	AllowHeaders     []string      `mapstructure:"allow_headers" json:"allow_headers" yaml:"allow_headers"`
	ExposeHeaders    []string      `mapstructure:"expose_headers" json:"expose_headers" yaml:"expose_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials" json:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age" json:"max_age" yaml:"max_age"`
}

// BaseConfig represents common configuration for all services
type BaseConfig struct {
	Environment string `mapstructure:"environment" json:"environment" yaml:"environment"`
	ServiceName string `mapstructure:"service_name" json:"service_name" yaml:"service_name"`
	Version     string `mapstructure:"version" json:"version" yaml:"version"`
}

// Helper functions

// LoadFromFile loads configuration from a specific file
func LoadFromFile(filename string, config interface{}) error {
	v := viper.New()
	
	// Set the file name and path
	v.SetConfigFile(filename)
	
	// Read the file
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file %s: %w", filename, err)
	}
	
	// Unmarshal into struct
	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	return nil
}

// LoadFromEnv loads configuration from environment variables with the given prefix
func LoadFromEnv(prefix string, config interface{}) error {
	v := viper.New()
	v.SetEnvPrefix(prefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// Unmarshal into struct
	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("error unmarshaling config from env: %w", err)
	}
	
	return nil
}

// ValidateRequired validates that required fields are set
func ValidateRequired(config interface{}, requiredFields []string) error {
	v := viper.New()
	if err := v.MergeConfigMap(structToMap(config)); err != nil {
		return err
	}
	
	var missing []string
	for _, field := range requiredFields {
		if !v.IsSet(field) || v.GetString(field) == "" {
			missing = append(missing, field)
		}
	}
	
	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration fields: %v", missing)
	}
	
	return nil
}

// GetEnvironment returns the current environment (development, staging, production)
func GetEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("APP_ENVIRONMENT")
	}
	if env == "" {
		return "development"
	}
	return strings.ToLower(env)
}

// IsProduction returns true if running in production environment
func IsProduction() bool {
	return GetEnvironment() == "production"
}

// IsDevelopment returns true if running in development environment
func IsDevelopment() bool {
	return GetEnvironment() == "development"
}

// GetConfigPath returns the configuration file path for the current environment
func GetConfigPath(serviceName string) string {
	env := GetEnvironment()
	filename := fmt.Sprintf("config.%s.yaml", env)
	
	// Check common locations
	paths := []string{
		filepath.Join(".", filename),
		filepath.Join(".", "config", filename),
		filepath.Join(".", "configs", filename),
		filepath.Join("/etc", serviceName, filename),
		filepath.Join(os.Getenv("HOME"), ".config", serviceName, filename),
	}
	
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	// Return default path if none found
	return filepath.Join(".", "config", filename)
}

// structToMap converts a struct to a map for viper
func structToMap(obj interface{}) map[string]interface{} {
	// This is a simplified implementation
	// In a real implementation, you'd use reflection or a library like mapstructure
	return make(map[string]interface{})
}