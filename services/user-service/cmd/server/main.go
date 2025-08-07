package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"user-service/internal/config"
	"user-service/internal/handler"
	"user-service/internal/model"
	"user-service/internal/repository"
	"user-service/internal/service"
	"user-service/pkg/database"
	"user-service/pkg/logger"

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
		"service": "user-service",
		"version": "1.0.0",
		"port":    cfg.Server.Port,
	}).Info("Starting User Service")

	db, err := database.New(cfg, appLogger)
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to connect to database")
	}

	if err := db.Migrate(&model.User{}); err != nil {
		appLogger.WithError(err).Fatal("Failed to run database migrations")
	}

	userRepo := repository.NewUserRepository(db)
	authClient := service.NewAuthClient(cfg, appLogger)
	userService := service.NewUserService(userRepo, authClient, appLogger)
	userHandler := handler.NewUserHandler(userService, appLogger)

	gin.SetMode(cfg.Server.Mode)
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		v1.POST("/auth/register", userHandler.Register)
		v1.POST("/auth/login", userHandler.Login)
		
		users := v1.Group("/users")
		{
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.POST("/:id/change-password", userHandler.ChangePassword)
		}
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "user-service",
			"time":    time.Now(),
		})
	})

	srv := &http.Server{
		Addr:    cfg.GetServerAddress(),
		Handler: router,
	}

	go func() {
		appLogger.WithField("address", cfg.GetServerAddress()).Info("User Service started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.WithError(err).Fatal("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down User Service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.WithError(err).Fatal("Server forced to shutdown")
	}

	appLogger.Info("User Service stopped")
}