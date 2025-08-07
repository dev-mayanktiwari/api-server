package main

import (
	"log"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/config"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
)

// Config represents the application configuration
type Config struct {
	config.BaseConfig `mapstructure:",squash"`
	Server            config.ServerConfig     `mapstructure:"server" json:"server" yaml:"server"`
	JWT               config.JWTConfig        `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Logging           logger.Config           `mapstructure:"logging" json:"logging" yaml:"logging"`
	RateLimit         config.RateLimitConfig  `mapstructure:"rate_limit" json:"rate_limit" yaml:"rate_limit"`
	CORS              config.CORSConfig       `mapstructure:"cors" json:"cors" yaml:"cors"`
	Services          ServiceURLs             `mapstructure:"services" json:"services" yaml:"services"`
}

// ServiceURLs contains URLs for downstream services
type ServiceURLs struct {
	AuthServiceURL string `mapstructure:"auth_service_url" json:"auth_service_url" yaml:"auth_service_url"`
	UserServiceURL string `mapstructure:"user_service_url" json:"user_service_url" yaml:"user_service_url"`
}

func main() {
	cfg := &Config{
		Logging: *logger.DefaultConfig(),
	}
	log.Printf("Logging config: %+v", cfg.Logging)
	
	appLogger, err := logger.New(&cfg.Logging)
	if err \!= nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	appLogger.Info("Test successful\!")
}
