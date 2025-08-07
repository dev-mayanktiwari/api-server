// Package errors provides custom error types and error handling utilities for the API server.
// It includes domain-specific errors, error codes, and error wrapping functionality.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Error codes
const (
	// General error codes
	CodeInternal       = "INTERNAL_ERROR"
	CodeBadRequest     = "BAD_REQUEST"
	CodeUnauthorized   = "UNAUTHORIZED"
	CodeForbidden      = "FORBIDDEN"
	CodeNotFound       = "NOT_FOUND"
	CodeConflict       = "CONFLICT"
	CodeValidation     = "VALIDATION_ERROR"
	CodeRateLimit      = "RATE_LIMIT_EXCEEDED"
	
	// Authentication error codes
	CodeInvalidToken   = "INVALID_TOKEN"
	CodeTokenExpired   = "TOKEN_EXPIRED"
	CodeInvalidLogin   = "INVALID_LOGIN"
	CodePasswordTooWeak = "PASSWORD_TOO_WEAK"
	
	// User error codes
	CodeUserNotFound   = "USER_NOT_FOUND"
	CodeUserExists     = "USER_ALREADY_EXISTS"
	CodeEmailTaken     = "EMAIL_ALREADY_TAKEN"
	CodeInvalidEmail   = "INVALID_EMAIL"
	
	// Database error codes
	CodeDatabaseError  = "DATABASE_ERROR"
	CodeQueryError     = "QUERY_ERROR"
	CodeConnectionError = "CONNECTION_ERROR"
	CodeMigrationError = "MIGRATION_ERROR"
	
	// Business logic error codes
	CodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
	CodeResourceLocked = "RESOURCE_LOCKED"
	CodeQuotaExceeded  = "QUOTA_EXCEEDED"
)

// AppError represents an application error with code, message, and HTTP status
type AppError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	HTTPStatus int                    `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Cause      error                  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithCause adds a cause to the error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// New creates a new AppError
func New(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Newf creates a new AppError with formatted message
func Newf(code string, httpStatus int, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		HTTPStatus: httpStatus,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Cause:      err,
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, code string, httpStatus int, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		HTTPStatus: httpStatus,
		Cause:      err,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError converts an error to AppError if possible
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}

// GetHTTPStatus returns the HTTP status code for an error
func GetHTTPStatus(err error) int {
	if appErr, ok := AsAppError(err); ok {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// GetErrorCode returns the error code for an error
func GetErrorCode(err error) string {
	if appErr, ok := AsAppError(err); ok {
		return appErr.Code
	}
	return CodeInternal
}

// Predefined errors for common use cases

// Internal server errors
func Internal(message string) *AppError {
	return New(CodeInternal, message, http.StatusInternalServerError)
}

func Internalf(format string, args ...interface{}) *AppError {
	return Newf(CodeInternal, http.StatusInternalServerError, format, args...)
}

// Bad request errors
func BadRequest(message string) *AppError {
	return New(CodeBadRequest, message, http.StatusBadRequest)
}

func BadRequestf(format string, args ...interface{}) *AppError {
	return Newf(CodeBadRequest, http.StatusBadRequest, format, args...)
}

// Unauthorized errors
func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message, http.StatusUnauthorized)
}

func Unauthorizedf(format string, args ...interface{}) *AppError {
	return Newf(CodeUnauthorized, http.StatusUnauthorized, format, args...)
}

// Forbidden errors
func Forbidden(message string) *AppError {
	return New(CodeForbidden, message, http.StatusForbidden)
}

func Forbiddenf(format string, args ...interface{}) *AppError {
	return Newf(CodeForbidden, http.StatusForbidden, format, args...)
}

// Not found errors
func NotFound(message string) *AppError {
	return New(CodeNotFound, message, http.StatusNotFound)
}

func NotFoundf(format string, args ...interface{}) *AppError {
	return Newf(CodeNotFound, http.StatusNotFound, format, args...)
}

// Conflict errors
func Conflict(message string) *AppError {
	return New(CodeConflict, message, http.StatusConflict)
}

func Conflictf(format string, args ...interface{}) *AppError {
	return Newf(CodeConflict, http.StatusConflict, format, args...)
}

// Validation errors
func Validation(message string) *AppError {
	return New(CodeValidation, message, http.StatusBadRequest)
}

func Validationf(format string, args ...interface{}) *AppError {
	return Newf(CodeValidation, http.StatusBadRequest, format, args...)
}

// Authentication errors
func InvalidToken(message string) *AppError {
	return New(CodeInvalidToken, message, http.StatusUnauthorized)
}

func TokenExpired(message string) *AppError {
	return New(CodeTokenExpired, message, http.StatusUnauthorized)
}

func InvalidLogin(message string) *AppError {
	return New(CodeInvalidLogin, message, http.StatusUnauthorized)
}

// User errors
func UserNotFound(message string) *AppError {
	if message == "" {
		message = "User not found"
	}
	return New(CodeUserNotFound, message, http.StatusNotFound)
}

func UserExists(message string) *AppError {
	if message == "" {
		message = "User already exists"
	}
	return New(CodeUserExists, message, http.StatusConflict)
}

func EmailTaken(message string) *AppError {
	if message == "" {
		message = "Email address is already taken"
	}
	return New(CodeEmailTaken, message, http.StatusConflict)
}

// Database errors
func DatabaseError(message string) *AppError {
	return New(CodeDatabaseError, message, http.StatusInternalServerError)
}

func DatabaseErrorf(format string, args ...interface{}) *AppError {
	return Newf(CodeDatabaseError, http.StatusInternalServerError, format, args...)
}

// Business logic errors
func InsufficientPermissions(message string) *AppError {
	if message == "" {
		message = "Insufficient permissions to perform this action"
	}
	return New(CodeInsufficientPermissions, message, http.StatusForbidden)
}

func ResourceLocked(message string) *AppError {
	if message == "" {
		message = "Resource is currently locked"
	}
	return New(CodeResourceLocked, message, http.StatusConflict)
}

// Error context for better debugging
type ErrorContext struct {
	UserID       string                 `json:"user_id,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	Operation    string                 `json:"operation,omitempty"`
	Resource     string                 `json:"resource,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	StackTrace   string                 `json:"stack_trace,omitempty"`
}

// ContextualError represents an error with additional context
type ContextualError struct {
	*AppError
	Context *ErrorContext `json:"context,omitempty"`
}

// Error implements the error interface
func (e *ContextualError) Error() string {
	return e.AppError.Error()
}

// WithContext adds context to an error
func WithContext(err *AppError, ctx *ErrorContext) *ContextualError {
	return &ContextualError{
		AppError: err,
		Context:  ctx,
	}
}

// Error handler middleware helper
type ErrorHandler struct {
	logger interface {
		WithFields(fields map[string]interface{}) interface{ Error(args ...interface{}) }
	}
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger interface {
	WithFields(fields map[string]interface{}) interface{ Error(args ...interface{}) }
}) *ErrorHandler {
	return &ErrorHandler{logger: logger}
}

// LogError logs an error with appropriate context
func (h *ErrorHandler) LogError(err error, ctx *ErrorContext) {
	fields := make(map[string]interface{})
	
	if ctx != nil {
		if ctx.UserID != "" {
			fields["user_id"] = ctx.UserID
		}
		if ctx.RequestID != "" {
			fields["request_id"] = ctx.RequestID
		}
		if ctx.Operation != "" {
			fields["operation"] = ctx.Operation
		}
		if ctx.Resource != "" {
			fields["resource"] = ctx.Resource
		}
		for k, v := range ctx.Metadata {
			fields[k] = v
		}
	}
	
	if appErr, ok := AsAppError(err); ok {
		fields["error_code"] = appErr.Code
		fields["http_status"] = appErr.HTTPStatus
		if len(appErr.Details) > 0 {
			fields["error_details"] = appErr.Details
		}
	}
	
	h.logger.WithFields(fields).Error(err.Error())
}