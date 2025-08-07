package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dev-mayanktiwari/api-server/internal/config"
	"github.com/dev-mayanktiwari/api-server/internal/server"
	"github.com/dev-mayanktiwari/api-server/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logConfig := logger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
	}

	if err := logger.InitGlobal(logConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Get logger instance
	appLogger := logger.GetGlobal()

	// Log application startup
	appLogger.WithFields(map[string]interface{}{
		"version":     "1.0.0",
		"environment": cfg.Server.Mode,
		"port":        cfg.Server.Port,
	}).Info("Starting API Server")

	// Create HTTP server
	httpServer, err := server.New(cfg, appLogger)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to create HTTP server")
	}

	// Start server in a goroutine
	go func() {
		if err := httpServer.Start(); err != nil {
			appLogger.WithError(err).Fatal("Failed to start HTTP server")
		}
	}()

	appLogger.WithField("address", cfg.GetServerAddress()).Info("HTTP server started successfully")

	// Print available routes (helpful for development)
	if cfg.IsDevelopment() {
		routes := httpServer.GetRoutes()
		appLogger.WithField("route_count", len(routes)).Info("Registered routes")
		for _, route := range routes {
			appLogger.WithFields(map[string]interface{}{
				"method": route.Method,
				"path":   route.Path,
			}).Debug("Route registered")
		}
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutdown signal received, shutting down gracefully...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := httpServer.Stop(ctx); err != nil {
		appLogger.WithError(err).Error("Failed to shutdown server gracefully")
		os.Exit(1)
	}

	appLogger.Info("Server stopped successfully")
}
