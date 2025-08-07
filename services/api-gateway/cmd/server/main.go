package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/internal/config"
	"api-gateway/internal/handler"
	"api-gateway/internal/middleware"
	"api-gateway/pkg/logger"

	"github.com/gin-contrib/cors"
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
		"service": "api-gateway",
		"version": "1.0.0",
		"port":    cfg.Server.Port,
	}).Info("Starting API Gateway")

	gin.SetMode(cfg.Server.Mode)
	router := gin.New()

	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		ExposeHeaders:    cfg.CORS.ExposeHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           time.Duration(cfg.CORS.MaxAge) * time.Second,
	}
	router.Use(cors.New(corsConfig))

	router.Use(middleware.RequestID())
	router.Use(middleware.LoggingMiddleware(appLogger))
	router.Use(middleware.RateLimitMiddleware())
	router.Use(gin.Recovery())

	healthHandler := handler.NewHealthHandler(cfg, appLogger)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", middleware.ProxyToService(cfg.Services.UserService, appLogger))
			auth.POST("/login", middleware.ProxyToService(cfg.Services.UserService, appLogger))
			auth.POST("/refresh", middleware.ProxyToService(cfg.Services.AuthService, appLogger))
			auth.POST("/validate", middleware.ProxyToService(cfg.Services.AuthService, appLogger))
		}

		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg, appLogger))
		{
			users.GET("", middleware.RoleMiddleware("admin"), middleware.ProxyToService(cfg.Services.UserService, appLogger))
			users.GET("/:id", middleware.ProxyToService(cfg.Services.UserService, appLogger))
			users.PUT("/:id", middleware.ProxyToService(cfg.Services.UserService, appLogger))
			users.DELETE("/:id", middleware.RoleMiddleware("admin"), middleware.ProxyToService(cfg.Services.UserService, appLogger))
			users.POST("/:id/change-password", middleware.ProxyToService(cfg.Services.UserService, appLogger))
		}
	}

	srv := &http.Server{
		Addr:    cfg.GetServerAddress(),
		Handler: router,
	}

	go func() {
		appLogger.WithField("address", cfg.GetServerAddress()).Info("API Gateway started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithError(err).Fatal("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down API Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.WithError(err).Fatal("Server forced to shutdown")
	}

	appLogger.Info("API Gateway stopped")
}