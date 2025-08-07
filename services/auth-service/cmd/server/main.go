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

	"auth-service/internal/config"
	"auth-service/internal/handler"
	"auth-service/internal/service"
	"auth-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logConfig := logger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
	}

	if err := logger.InitGlobal(logConfig); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	appLogger := logger.GetGlobal()

	appLogger.WithFields(map[string]interface{}{
		"service": "auth-service",
		"version": "1.0.0",
		"port":    cfg.Server.Port,
	}).Info("Starting Auth Service")

	authService := service.NewAuthService(cfg, appLogger)
	authHandler := handler.NewAuthHandler(authService, appLogger)

	gin.SetMode(cfg.Server.Mode)
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		v1.POST("/auth/generate", authHandler.GenerateTokens)
		v1.POST("/auth/validate", authHandler.ValidateToken)
		v1.POST("/auth/refresh", authHandler.RefreshTokens)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "auth-service",
			"time":    time.Now(),
		})
	})

	srv := &http.Server{
		Addr:    cfg.GetServerAddress(),
		Handler: router,
	}

	go func() {
		appLogger.WithField("address", cfg.GetServerAddress()).Info("Auth Service started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithError(err).Fatal("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down Auth Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.WithError(err).Fatal("Server forced to shutdown")
	}

	appLogger.Info("Auth Service stopped")
}