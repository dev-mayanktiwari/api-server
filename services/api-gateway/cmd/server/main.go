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
	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/middleware"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/auth"
	"github.com/dev-mayanktiwari/api-server/services/api-gateway/internal/infrastructure/http/handlers"
	"github.com/dev-mayanktiwari/api-server/services/api-gateway/internal/application/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load("API_GATEWAY")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.New(&cfg.Logging)
	appLogger.Info("Starting API Gateway...")

	// Initialize JWT manager (for auth validation)
	jwtManager := auth.NewJWTManager(&cfg.JWT, appLogger)

	// Initialize services
	proxyService := services.NewProxyService(appLogger)

	// Initialize handlers
	gatewayHandler := handlers.NewGatewayHandler(proxyService, appLogger)

	// Setup router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(appLogger))
	router.Use(middleware.RateLimit())

	// Health endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "api-gateway",
			"version":   cfg.Version,
			"timestamp": time.Now().UTC(),
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		// Check if downstream services are available
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"service":   "api-gateway",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes with service routing
	api := router.Group("/api/v1")
	{
		// Auth service routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/login", gatewayHandler.ProxyToAuthService)
			auth.POST("/refresh", gatewayHandler.ProxyToAuthService)
			auth.POST("/logout", gatewayHandler.ProxyToAuthService)
			auth.POST("/validate", gatewayHandler.ProxyToAuthService)
			auth.GET("/me", middleware.JWTAuth(jwtManager), gatewayHandler.ProxyToAuthService)
		}

		// User service routes
		users := api.Group("/users")
		{
			// Public routes
			users.POST("/register", gatewayHandler.ProxyToUserService)
			users.POST("/login", gatewayHandler.ProxyToAuthService) // Login goes to auth service

			// Protected routes
			protected := users.Use(middleware.JWTAuth(jwtManager))
			{
				protected.GET("/profile", gatewayHandler.ProxyToUserService)
				protected.PUT("/profile", gatewayHandler.ProxyToUserService)
				protected.POST("/change-password", gatewayHandler.ProxyToUserService)
			}

			// Admin routes
			admin := users.Use(middleware.JWTAuth(jwtManager), middleware.RequireRole("admin"))
			{
				admin.GET("", gatewayHandler.ProxyToUserService)           // List users
				admin.GET("/:id", gatewayHandler.ProxyToUserService)       // Get user by ID
				admin.PUT("/:id", gatewayHandler.ProxyToUserService)       // Update user
				admin.DELETE("/:id", gatewayHandler.ProxyToUserService)    // Delete user
				admin.GET("/statistics", gatewayHandler.ProxyToUserService) // Get statistics
			}
		}
	}

	// Start server
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		IdleTimeout:    cfg.Server.IdleTimeout,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Graceful shutdown
	go func() {
		appLogger.WithField("port", cfg.Server.Port).Info("API Gateway started successfully")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down API Gateway...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.WithError(err).Error("Failed to shutdown server gracefully")
	}

	appLogger.Info("API Gateway stopped")
}