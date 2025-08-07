package middleware

import (
	"time"

	"github.com/dev-mayanktiwari/api-server/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures CORS based on application configuration
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     cfg.CORS.AllowedMethods,
		AllowHeaders:     cfg.CORS.AllowedHeaders,
		ExposeHeaders:    []string{"X-Request-ID", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// In development, allow all origins if "*" is specified
	if cfg.IsDevelopment() {
		for _, origin := range cfg.CORS.AllowedOrigins {
			if origin == "*" {
				corsConfig.AllowAllOrigins = true
				corsConfig.AllowOrigins = nil
				break
			}
		}
	}

	return cors.New(corsConfig)
}
