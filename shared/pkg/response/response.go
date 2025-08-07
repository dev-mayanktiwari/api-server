// Package response provides standardized HTTP response utilities for the API server.
// It includes consistent response formatting, error handling, and validation error processing.
package response

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Response represents the standard API response structure
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorInfo provides detailed error information
type ErrorInfo struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Fields  []FieldError           `json:"fields,omitempty"`
}

// FieldError represents validation errors for specific fields
type FieldError struct {
	Field   string `json:"field"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Meta provides additional response metadata
type Meta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// Error codes
const (
	// General error codes
	ErrCodeInternal       = "INTERNAL_SERVER_ERROR"
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeForbidden      = "FORBIDDEN"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeRateLimit      = "RATE_LIMIT_EXCEEDED"
	
	// Authentication error codes
	ErrCodeInvalidToken   = "INVALID_TOKEN"
	ErrCodeTokenExpired   = "TOKEN_EXPIRED"
	ErrCodeInvalidLogin   = "INVALID_LOGIN"
	
	// User error codes
	ErrCodeUserNotFound   = "USER_NOT_FOUND"
	ErrCodeUserExists     = "USER_ALREADY_EXISTS"
	ErrCodeEmailTaken     = "EMAIL_ALREADY_TAKEN"
	
	// Database error codes
	ErrCodeDatabaseError  = "DATABASE_ERROR"
	ErrCodeQueryError     = "QUERY_ERROR"
)

// Success sends a successful response
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	response := Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	
	c.JSON(statusCode, response)
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(c *gin.Context, message string, data interface{}, meta *Meta) {
	response := Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	
	c.JSON(http.StatusOK, response)
}

// Created sends a 201 Created response
func Created(c *gin.Context, message string, data interface{}) {
	response := Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	
	c.JSON(http.StatusCreated, response)
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, code, message string) {
	response := Response{
		Success:   false,
		Message:   "Request failed",
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	
	c.JSON(statusCode, response)
}

// ErrorWithDetails sends an error response with additional details
func ErrorWithDetails(c *gin.Context, statusCode int, code, message string, details map[string]interface{}) {
	response := Response{
		Success:   false,
		Message:   "Request failed",
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	
	c.JSON(statusCode, response)
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, err error) {
	fieldErrors := extractValidationErrors(err)
	
	response := Response{
		Success:   false,
		Message:   "Validation failed",
		Error: &ErrorInfo{
			Code:    ErrCodeValidation,
			Message: "One or more fields contain invalid values",
			Fields:  fieldErrors,
		},
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	
	c.JSON(http.StatusBadRequest, response)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string) {
	if message == "" {
		message = "An internal server error occurred"
	}
	
	Error(c, http.StatusInternalServerError, ErrCodeInternal, message)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string, details ...string) {
	if message == "" {
		message = "Bad request"
	}
	
	if len(details) > 0 {
		Error(c, http.StatusBadRequest, ErrCodeBadRequest, message+": "+details[0])
	} else {
		Error(c, http.StatusBadRequest, ErrCodeBadRequest, message)
	}
}

// BadGateway sends a 502 Bad Gateway response
func BadGateway(c *gin.Context, message string) {
	if message == "" {
		message = "Bad gateway"
	}
	
	Error(c, http.StatusBadGateway, "BAD_GATEWAY", message)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	
	Error(c, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	
	Error(c, http.StatusForbidden, ErrCodeForbidden, message)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}
	
	Error(c, http.StatusNotFound, ErrCodeNotFound, message)
}

// Conflict sends a 409 Conflict response
func Conflict(c *gin.Context, message string) {
	if message == "" {
		message = "Resource conflict"
	}
	
	Error(c, http.StatusConflict, ErrCodeConflict, message)
}

// TooManyRequests sends a 429 Too Many Requests response
func TooManyRequests(c *gin.Context, message string) {
	if message == "" {
		message = "Rate limit exceeded"
	}
	
	Error(c, http.StatusTooManyRequests, ErrCodeRateLimit, message)
}

// Pagination creates pagination metadata
func Pagination(page, limit, total int) *Meta {
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}
	
	return &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// extractValidationErrors extracts field validation errors from validator errors
func extractValidationErrors(err error) []FieldError {
	var fieldErrors []FieldError
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, validationError := range validationErrors {
			var value string
			if validationError.Value() != nil {
				value = fmt.Sprintf("%v", validationError.Value())
			}
			fieldError := FieldError{
				Field:   validationError.Field(),
				Value:   value,
				Code:    validationError.Tag(),
				Message: getValidationMessage(validationError),
			}
			fieldErrors = append(fieldErrors, fieldError)
		}
	}
	
	return fieldErrors
}

// getValidationMessage returns a human-readable validation error message
func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Please enter a valid email address"
	case "min":
		return "This field must be at least " + fe.Param() + " characters long"
	case "max":
		return "This field must be at most " + fe.Param() + " characters long"
	case "len":
		return "This field must be exactly " + fe.Param() + " characters long"
	case "alpha":
		return "This field must contain only alphabetic characters"
	case "alphanum":
		return "This field must contain only alphanumeric characters"
	case "numeric":
		return "This field must contain only numeric characters"
	case "url":
		return "Please enter a valid URL"
	case "uuid":
		return "Please enter a valid UUID"
	case "oneof":
		return "This field must be one of: " + fe.Param()
	case "gte":
		return "This field must be greater than or equal to " + fe.Param()
	case "lte":
		return "This field must be less than or equal to " + fe.Param()
	case "gt":
		return "This field must be greater than " + fe.Param()
	case "lt":
		return "This field must be less than " + fe.Param()
	default:
		return "This field is invalid"
	}
}

// getRequestID extracts the request ID from the Gin context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// HealthCheck sends a health check response
func HealthCheck(c *gin.Context, status string, checks map[string]interface{}) {
	statusCode := http.StatusOK
	if status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}
	
	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now(),
		"checks":    checks,
	}
	
	c.JSON(statusCode, response)
}