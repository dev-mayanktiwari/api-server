// Package main is the entry point for the user service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/config"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/database"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/middleware"
	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/application/services"
	domainServices "github.com/dev-mayanktiwari/api-server/services/user-service/internal/domain/services"
	dbImpl "github.com/dev-mayanktiwari/api-server/services/user-service/internal/infrastructure/database"
	"github.com/dev-mayanktiwari/api-server/services/user-service/internal/infrastructure/http/handlers"
)

// ServiceConfig represents the user service configuration
type ServiceConfig struct {
	config.BaseConfig `mapstructure:",squash"`
	Server            config.ServerConfig   `mapstructure:"server"`
	Database          config.DatabaseConfig `mapstructure:"database"`
	RateLimit         config.RateLimitConfig `mapstructure:"rate_limit"`
	CORS              config.CORSConfig      `mapstructure:"cors"`
	Logging           *logger.Config         `mapstructure:"logging"`
}

// DefaultConfig returns default service configuration
func DefaultConfig() *ServiceConfig {
	return &ServiceConfig{
		BaseConfig: config.BaseConfig{
			Environment: "development",
			ServiceName: "user-service",
			Version:     "v1.0.0",
		},
		Server: config.ServerConfig{
			Host:            "localhost",
			Port:            8082,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
		Database: *database.DefaultConfig(),
		JWT: config.JWTConfig{
			Secret:         "your-secret-key",
			Issuer:         "user-service",
			ExpirationTime: 24 * time.Hour,
			RefreshTime:    7 * 24 * time.Hour,
			Algorithm:      "HS256",
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerSecond: 10.0,
			BurstSize:         20,
			CleanupInterval:   5 * time.Minute,
		},
		CORS: config.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"X-Request-ID"},
			AllowCredentials: false,
			MaxAge:           12 * time.Hour,
		},
		Logging: logger.DefaultConfig(),
	}
}

func main() {
	// Load configuration
	cfg := DefaultConfig()
	configManager := config.New(&config.Options{
		ConfigName:  "config",
		ConfigPaths: []string{".", "./configs/" + config.GetEnvironment()},
		ConfigType:  "yaml",
		EnvPrefix:   "USER_SERVICE",
	})

	if err := configManager.Load(cfg); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := logger.New(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// Set global logger
	logger.SetDefault(logger)

	logger.WithFields(logger.Fields{
		"service": cfg.ServiceName,
		"version": cfg.Version,
		"env":     cfg.Environment,
	}).Info("Starting user service")

	// Initialize database
	db, err := database.Connect(&cfg.Database, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Auto-migrate database schema
	if err := db.Migrate(&dbImpl.UserModel{}); err != nil {
		logger.WithError(err).Fatal("Failed to migrate database")
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(&cfg.JWT, logger)

	// Initialize repositories
	userRepo := dbImpl.NewUserRepository(db, logger)

	// Initialize domain services
	userDomainService := domainServices.NewUserDomainService(userRepo, logger)

	// Initialize application services
	userAppService := services.NewUserApplicationService(userRepo, userDomainService, jwtManager, logger)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userAppService, logger)

	// Initialize HTTP server
	server := initializeServer(cfg, logger, userHandler, jwtManager, db)

	// Start server
	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		logger.WithFields(logger.Fields{
			"address": addr,
		}).Info("Starting HTTP server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Shutdown server gracefully
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("User service stopped")
}

// initializeServer initializes the HTTP server with middleware and routes
func initializeServer(cfg *ServiceConfig, log *logger.Logger, userHandler *handlers.UserHandler, jwtManager *auth.JWTManager, db *database.DB) *http.Server {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Logging(log))
	router.Use(middleware.Recovery(log))
	router.Use(middleware.CORS(&cfg.CORS))
	router.Use(middleware.RateLimit(&cfg.RateLimit))
	router.Use(middleware.Security())

	// Health check routes
	router.GET("/health", healthCheck(log))
	router.GET("/ready", readinessCheck(log, db))

	// API routes
	api := router.Group("/api/v1")
	{
		// User routes
		users := api.Group("/users")
		userHandler.RegisterRoutes(users, auth.AuthMiddleware(jwtManager))
	}

	// Create HTTP server
	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}
}

// healthCheck returns a simple health check handler
func healthCheck(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "user-service",
			"timestamp": time.Now().UTC(),
		})
	}
}

// readinessCheck returns a readiness check handler
func readinessCheck(log *logger.Logger, db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		checks := make(map[string]interface{})

		// Check database connection if provided
		if db != nil {
			if err := db.HealthCheck(c.Request.Context()); err != nil {
				checks["database"] = map[string]interface{}{
					"status": "unhealthy",
					"error":  err.Error(),
				}
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status":    "not ready",
					"service":   "user-service",
					"checks":    checks,
					"timestamp": time.Now().UTC(),
				})
				return
			}
			checks["database"] = map[string]interface{}{
				"status": "healthy",
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"service":   "user-service",
			"checks":    checks,
			"timestamp": time.Now().UTC(),
		})
	}
}