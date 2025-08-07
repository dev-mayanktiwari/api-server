package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dev-mayanktiwari/api-server/internal/config"
	"github.com/dev-mayanktiwari/api-server/internal/handler"
	"github.com/dev-mayanktiwari/api-server/internal/middleware"
	"github.com/dev-mayanktiwari/api-server/internal/model"
	"github.com/dev-mayanktiwari/api-server/internal/repository"
	"github.com/dev-mayanktiwari/api-server/internal/service"
	"github.com/dev-mayanktiwari/api-server/pkg/auth"
	"github.com/dev-mayanktiwari/api-server/pkg/database"
	"github.com/dev-mayanktiwari/api-server/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/pkg/response"
	"github.com/gin-gonic/gin"
)

// Server represents the HTTP server
type Server struct {
	config      *config.Config
	logger      *logger.Logger
	db          *database.Database
	jwtManager  *auth.JWTManager
	httpServer  *http.Server
	router      *gin.Engine
	userHandler *handler.UserHandler
}

// New creates a new HTTP server instance
func New(cfg *config.Config, logger *logger.Logger) (*Server, error) {
	// Set Gin mode based on configuration
	gin.SetMode(cfg.Server.Mode)

	// Control GIN debug logging
	if cfg.Logger.DisableGin {
		gin.DefaultWriter = io.Discard
		gin.DefaultErrorWriter = io.Discard
	}

	// Initialize database
	db, err := database.New(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run database migrations
	if err := db.Migrate(&model.User{}); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expiry)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB, logger)

	// Initialize services
	userService := service.NewUserService(userRepo, jwtManager, logger)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService, logger)

	// Create Gin router
	router := gin.New()

	// Create HTTP server
	httpServer := &http.Server{
		Addr:           cfg.GetServerAddress(),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	server := &Server{
		config:      cfg,
		logger:      logger,
		db:          db,
		jwtManager:  jwtManager,
		httpServer:  httpServer,
		router:      router,
		userHandler: userHandler,
	}

	// Setup middlewares and routes
	server.setupMiddlewares()
	server.setupRoutes()

	return server, nil
}

// setupMiddlewares configures all middlewares
func (s *Server) setupMiddlewares() {
	// Recovery middleware (must be first)
	s.router.Use(middleware.RecoveryMiddleware(s.logger))

	// Request ID middleware
	s.router.Use(middleware.RequestIDMiddleware())

	// CORS middleware
	s.router.Use(middleware.CORSMiddleware(s.config))

	// Logging middleware
	s.router.Use(middleware.LoggingMiddleware(s.logger))

	// JSON validation middleware
	s.router.Use(middleware.ValidateJSON())

	// Security headers middleware
	s.router.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	})
}

// setupRoutes configures all routes
func (s *Server) setupRoutes() {
	// Create handlers
	healthHandler := handler.NewHealthHandler("1.0.0")

	// Health check routes (no authentication required)
	s.router.GET("/health", healthHandler.Health)
	s.router.GET("/ready", healthHandler.Ready)
	s.router.GET("/live", healthHandler.Liveness)
	s.router.GET("/version", healthHandler.Version)

	// API routes group with rate limiting
	api := s.router.Group("/api")
	api.Use(middleware.APIRateLimitMiddleware())
	{
		// API health check
		api.GET("/health", healthHandler.Health)

		// V1 API routes
		v1 := api.Group("/v1")
		{
			// Public endpoints (no authentication required)
			v1.GET("/ping", s.pingHandler)

			// Auth endpoints with strict rate limiting
			auth := v1.Group("/auth")
			auth.Use(middleware.AuthRateLimitMiddleware())
			{
				auth.POST("/register", s.userHandler.Register)
				auth.POST("/login", s.userHandler.Login)
			}

			// Protected endpoints (authentication required)
			protected := v1.Group("/")
			protected.Use(middleware.AuthMiddleware(s.jwtManager, s.logger))
			{
				// User profile endpoints
				profile := protected.Group("/profile")
				{
					profile.GET("", s.userHandler.GetProfile)
					profile.PUT("", s.userHandler.UpdateProfile)
					profile.POST("/change-password", s.userHandler.ChangePassword)
				}

				// Admin endpoints (admin role required)
				admin := protected.Group("/admin")
				admin.Use(middleware.AdminMiddleware())
				{
					// User management
					users := admin.Group("/users")
					users.Use(middleware.ValidatePagination())
					{
						users.GET("", s.userHandler.ListUsers)
						users.GET("/:id", s.userHandler.GetUser)
						users.PUT("/:id", s.userHandler.UpdateUser)
						users.DELETE("/:id", s.userHandler.DeleteUser)
					}
				}
			}
		}
	}

	// Handle 404 for unknown routes
	s.router.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "The requested endpoint was not found")
	})

	// Handle 405 for wrong methods
	s.router.NoMethod(func(c *gin.Context) {
		response.Error(c, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "The requested method is not allowed for this endpoint")
	})
}

// pingHandler is a simple ping endpoint for testing
func (s *Server) pingHandler(c *gin.Context) {
	response.Success(c, "pong", gin.H{
		"timestamp": time.Now(),
		"server":    "api-server",
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.WithFields(map[string]interface{}{
		"address": s.config.GetServerAddress(),
		"mode":    s.config.Server.Mode,
	}).Info("Starting HTTP server")

	// Start server in a goroutine
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")

	// Attempt graceful shutdown of HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to gracefully shutdown server")
		return err
	}

	// Close database connection
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			s.logger.WithError(err).Error("Failed to close database connection")
			return err
		}
	}

	s.logger.Info("HTTP server stopped")
	return nil
}

// Router returns the Gin router instance
func (s *Server) Router() *gin.Engine {
	return s.router
}

// GetRoutes returns a list of registered routes (useful for debugging)
func (s *Server) GetRoutes() []gin.RouteInfo {
	return s.router.Routes()
}
