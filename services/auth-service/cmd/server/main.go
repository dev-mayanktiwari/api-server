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
	"github.com/dev-mayanktiwari/api-server/shared/pkg/database"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/middleware"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/auth"
	"github.com/dev-mayanktiwari/api-server/services/auth-service/internal/infrastructure/http/handlers"
	"github.com/dev-mayanktiwari/api-server/services/auth-service/internal/infrastructure/database"
	"github.com/dev-mayanktiwari/api-server/services/auth-service/internal/application/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load("AUTH_SERVICE")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger := logger.New(&cfg.Logging)
	appLogger.Info("Starting Auth Service...")

	// Initialize database
	db, err := database.NewPostgreSQL(&cfg.Database)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to connect to database")
	}
	defer func() {
		if err := database.Close(db); err != nil {
			appLogger.WithError(err).Error("Failed to close database connection")
		}
	}()

	// Auto-migrate tables
	if err := database.AutoMigrate(db, &database.TokenModel{}); err != nil {
		appLogger.WithError(err).Fatal("Failed to migrate database")
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(&cfg.JWT, appLogger)

	// Initialize repositories
	tokenRepo := database.NewTokenRepository(db, appLogger)

	// Initialize services
	authService := services.NewAuthApplicationService(tokenRepo, jwtManager, appLogger)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, appLogger)

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
			"service":   "auth-service",
			"version":   cfg.Version,
			"timestamp": time.Now().UTC(),
		})
	})

	router.GET("/ready", func(c *gin.Context) {
		// Check database connectivity
		if err := database.Ping(db); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"checks": gin.H{
					"database": gin.H{"status": "unhealthy", "error": err.Error()},
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ready",
			"service": "auth-service",
			"checks": gin.H{
				"database": gin.H{"status": "healthy"},
			},
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1/auth")
	{
		v1.POST("/login", authHandler.Login)
		v1.POST("/refresh", authHandler.RefreshToken)
		v1.POST("/logout", authHandler.Logout)
		v1.POST("/validate", authHandler.ValidateToken)
		v1.GET("/me", middleware.JWTAuth(jwtManager), authHandler.GetCurrentUser)
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
		appLogger.WithField("port", cfg.Server.Port).Info("Auth Service started successfully")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down Auth Service...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.WithError(err).Error("Failed to shutdown server gracefully")
	}

	appLogger.Info("Auth Service stopped")
}