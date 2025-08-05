package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationMiddleware handles request validation
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// HandleValidationErrors converts validation errors to user-friendly format
func HandleValidationErrors(err error) gin.H {
	var validationErrors []gin.H
	
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range errs {
			validationErrors = append(validationErrors, gin.H{
				"field":   strings.ToLower(fieldErr.Field()),
				"message": getValidationMessage(fieldErr),
			})
		}
	} else {
		// Handle other types of errors
		validationErrors = append(validationErrors, gin.H{
			"field":   "unknown",
			"message": err.Error(),
		})
	}
	
	return gin.H{
		"success": false,
		"message": "Validation failed",
		"error": gin.H{
			"code":    "VALIDATION_ERROR",
			"message": "The request contains invalid data",
			"details": validationErrors,
		},
		"timestamp": time.Now(),
	}
}

// getValidationMessage returns user-friendly validation messages
func getValidationMessage(fieldErr validator.FieldError) string {
	field := strings.ToLower(fieldErr.Field())
	
	switch fieldErr.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		if fieldErr.Kind().String() == "string" {
			return field + " must be at least " + fieldErr.Param() + " characters long"
		}
		return field + " must be at least " + fieldErr.Param()
	case "max":
		if fieldErr.Kind().String() == "string" {
			return field + " must be at most " + fieldErr.Param() + " characters long"
		}
		return field + " must be at most " + fieldErr.Param()
	case "len":
		return field + " must be exactly " + fieldErr.Param() + " characters long"
	case "alpha":
		return field + " must contain only alphabetic characters"
	case "alphanum":
		return field + " must contain only alphanumeric characters"
	case "numeric":
		return field + " must be a number"
	case "url":
		return field + " must be a valid URL"
	case "uuid":
		return field + " must be a valid UUID"
	case "oneof":
		return field + " must be one of: " + fieldErr.Param()
	case "gte":
		return field + " must be greater than or equal to " + fieldErr.Param()
	case "lte":
		return field + " must be less than or equal to " + fieldErr.Param()
	case "gt":
		return field + " must be greater than " + fieldErr.Param()
	case "lt":
		return field + " must be less than " + fieldErr.Param()
	default:
		return field + " is invalid"
	}
}

// ValidateJSON middleware for JSON payload validation
func ValidateJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if content-type is JSON for POST, PUT, PATCH requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Invalid content type",
					"error": gin.H{
						"code":    "INVALID_CONTENT_TYPE",
						"message": "Content-Type must be application/json",
					},
					"timestamp":  time.Now(),
					"request_id": c.GetString("request_id"),
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

// ValidatePagination middleware for pagination parameters
func ValidatePagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get page parameter
		page := c.DefaultQuery("page", "1")
		if page == "" || page == "0" {
			page = "1"
		}
		
		// Get limit parameter
		limit := c.DefaultQuery("limit", "10")
		if limit == "" {
			limit = "10"
		}
		
		// Validate page is a positive integer
		if !isPositiveInteger(page) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid page parameter",
				"error": gin.H{
					"code":    "INVALID_PAGE",
					"message": "Page must be a positive integer",
				},
				"timestamp":  time.Now(),
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}
		
		// Validate limit is a positive integer and not too large
		if !isPositiveInteger(limit) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid limit parameter",
				"error": gin.H{
					"code":    "INVALID_LIMIT",
					"message": "Limit must be a positive integer",
				},
				"timestamp":  time.Now(),
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}
		
		// Set maximum limit
		if parseToInt(limit) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Limit too large",
				"error": gin.H{
					"code":    "LIMIT_TOO_LARGE",
					"message": "Limit cannot be greater than 100",
				},
				"timestamp":  time.Now(),
				"request_id": c.GetString("request_id"),
			})
			c.Abort()
			return
		}
		
		// Set validated values in context
		c.Set("page", parseToInt(page))
		c.Set("limit", parseToInt(limit))
		
		c.Next()
	}
}

// Helper function to check if string is a positive integer
func isPositiveInteger(s string) bool {
	if s == "" || s == "0" {
		return false
	}
	
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	
	return true
}

// Helper function to parse string to int
func parseToInt(s string) int {
	result := 0
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		}
	}
	return result
}