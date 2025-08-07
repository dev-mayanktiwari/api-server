package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"api-gateway/pkg/logger"

	"github.com/gin-gonic/gin"
)

func ProxyToService(serviceURL string, logger *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		var body []byte
		if c.Request.Body != nil {
			var err error
			body, err = io.ReadAll(c.Request.Body)
			if err != nil {
				logger.WithError(err).Error("Failed to read request body")
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Internal server error",
					"error": gin.H{
						"code":    "PROXY_ERROR",
						"message": "Failed to process request",
					},
				})
				return
			}
			c.Request.Body.Close()
		}

		targetURL := serviceURL + c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}

		req, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(body))
		if err != nil {
			logger.WithError(err).Error("Failed to create proxy request")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Internal server error",
				"error": gin.H{
					"code":    "PROXY_ERROR",
					"message": "Failed to create request",
				},
			})
			return
		}

		for key, values := range c.Request.Header {
			if strings.ToLower(key) == "host" {
				continue
			}
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}

		if userID, exists := c.Get("user_id"); exists {
			req.Header.Set("X-User-ID", userID.(string))
		}
		if email, exists := c.Get("user_email"); exists {
			req.Header.Set("X-User-Email", email.(string))
		}
		if role, exists := c.Get("user_role"); exists {
			req.Header.Set("X-User-Role", role.(string))
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.WithError(err).Error("Failed to proxy request")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"message": "Service unavailable",
				"error": gin.H{
					"code":    "SERVICE_UNAVAILABLE",
					"message": "The requested service is currently unavailable",
				},
			})
			return
		}
		defer resp.Body.Close()

		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		c.Status(resp.StatusCode)

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.WithError(err).Error("Failed to read service response")
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Internal server error",
			})
			return
		}

		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), responseBody)
	}
}