package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// APIResponse represents a standard API response structure
type APIResponse struct {
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
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Meta provides additional metadata for responses
type Meta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page  int `json:"page" form:"page" binding:"min=1"`
	Limit int `json:"limit" form:"limit" binding:"min=1,max=100"`
}

// Success sends a successful response
func Success(c *gin.Context, message string, data interface{}) {
	response := APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusOK, response)
}

// Created sends a successful creation response
func Created(c *gin.Context, message string, data interface{}) {
	response := APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusCreated, response)
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(c *gin.Context, message string, data interface{}, meta *Meta) {
	response := APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Meta:      meta,
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusOK, response)
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, code string, message string) {
	response := APIResponse{
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

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, details map[string]string) {
	response := APIResponse{
		Success:   false,
		Message:   "Validation failed",
		Error: &ErrorInfo{
			Code:    "VALIDATION_ERROR",
			Message: "One or more fields are invalid",
			Details: details,
		},
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	}
	c.JSON(http.StatusBadRequest, response)
}

// BadRequest sends a bad request error response
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized sends an unauthorized error response
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a forbidden error response
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a not found error response
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict sends a conflict error response
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, "CONFLICT", message)
}

// InternalServerError sends an internal server error response
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
}

// ServiceUnavailable sends a service unavailable error response
func ServiceUnavailable(c *gin.Context, message string) {
	Error(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}

// getRequestID extracts request ID from context
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// CalculateMeta calculates pagination metadata
func CalculateMeta(page, limit, total int) *Meta {
	totalPages := (total + limit - 1) / limit // Ceiling division
	return &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// GetPaginationParams extracts and validates pagination parameters
func GetPaginationParams(c *gin.Context) PaginationParams {
	var params PaginationParams
	
	// Set defaults
	params.Page = 1
	params.Limit = 10
	
	// Bind query parameters
	if err := c.ShouldBindQuery(&params); err != nil {
		// If binding fails, return defaults
		return PaginationParams{Page: 1, Limit: 10}
	}
	
	// Additional validation
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 10
	}
	
	return params
}