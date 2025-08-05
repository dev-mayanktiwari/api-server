package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/dev-mayanktiwari/api-server/internal/config"
	"github.com/dev-mayanktiwari/api-server/internal/handler"
	"github.com/dev-mayanktiwari/api-server/internal/middleware"
	"github.com/dev-mayanktiwari/api-server/pkg/logger"
	"github.com/dev-mayanktiwari/api-server/pkg/response"
)

// Server represents the HTTP server
type Server struct {
	config     *config.Config
	logger     *logger.Logger
	httpServer *http.Server
	router     *gin.Engine
}

// New creates a new HTTP server instance
func New(cfg *config.Config, logger *logger.Logger) *Server {
	// Set Gin mode based on configuration
	gin.SetMode(cfg.Server.Mode)

	// Control GIN debug logging
	if cfg.Logger.DisableGin {
		gin.DefaultWriter = io.Discard
		gin.DefaultErrorWriter = io.Discard
	}

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
		config:     cfg,
		logger:     logger,
		httpServer: httpServer,
		router:     router,
	}

	// Setup middlewares and routes
	server.setupMiddlewares()
	server.setupRoutes()

	return server
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

	// API routes group
	api := s.router.Group("/api")
	{
		// API health check
		api.GET("/health", healthHandler.Health)

		// V1 API routes
		v1 := api.Group("/v1")
		{
			// We'll add more endpoints here later
			v1.GET("/ping", s.pingHandler)
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

	// Attempt graceful shutdown
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to gracefully shutdown server")
		return err
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