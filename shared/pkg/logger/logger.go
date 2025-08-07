// Package logger provides structured logging functionality for the API server.
// It wraps zap logger with additional features like correlation IDs, user action logging,
// and configurable log levels.
package logger

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.SugaredLogger with additional functionality
type Logger struct {
	*zap.SugaredLogger
	config    *Config
	zapLogger *zap.Logger
}

// Config holds logger configuration
type Config struct {
	// Level sets the minimum enabled logging level
	Level string `mapstructure:"level" json:"level" yaml:"level"`
	
	// Format sets the log format (json or console)
	Format string `mapstructure:"format" json:"format" yaml:"format"`
	
	// Output sets the output destination (stdout, stderr, or file path)
	Output string `mapstructure:"output" json:"output" yaml:"output"`
	
	// DisableCaller disables adding the calling function as a field
	DisableCaller bool `mapstructure:"disable_caller" json:"disable_caller" yaml:"disable_caller"`
	
	// DisableStacktrace disables automatic stacktrace capturing
	DisableStacktrace bool `mapstructure:"disable_stacktrace" json:"disable_stacktrace" yaml:"disable_stacktrace"`
	
	// ServiceName adds service name to all log entries
	ServiceName string `mapstructure:"service_name" json:"service_name" yaml:"service_name"`
	
	// ServiceVersion adds service version to all log entries
	ServiceVersion string `mapstructure:"service_version" json:"service_version" yaml:"service_version"`
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:             "info",
		Format:            "json",
		Output:            "stdout",
		DisableCaller:     false,
		DisableStacktrace: false,
		ServiceName:       "api-server",
		ServiceVersion:    "v1.0.0",
	}
}

// Fields represents structured logging fields
type Fields map[string]interface{}

// ContextKey represents context keys for logging
type ContextKey string

const (
	// CorrelationIDKey is the context key for correlation ID
	CorrelationIDKey ContextKey = "correlation_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
)

var (
	defaultLogger *Logger
	loggerMutex   sync.RWMutex
)

// New creates a new logger instance with the given configuration
func New(config *Config) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Parse log level
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level %s: %w", config.Level, err)
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.MessageKey = "message"
	encoderConfig.LevelKey = "level"
	encoderConfig.CallerKey = "caller"
	encoderConfig.StacktraceKey = "stacktrace"

	// Create encoder based on format
	var encoder zapcore.Encoder
	switch config.Format {
	case "console":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	case "json":
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	default:
		return nil, fmt.Errorf("invalid log format %s", config.Format)
	}

	// Create writer syncer
	var writeSyncer zapcore.WriteSyncer
	switch config.Output {
	case "stdout":
		writeSyncer = zapcore.AddSync(os.Stdout)
	case "stderr":
		writeSyncer = zapcore.AddSync(os.Stderr)
	default:
		// Assume it's a file path
		file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %s: %w", config.Output, err)
		}
		writeSyncer = zapcore.AddSync(file)
	}

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Create logger options
	opts := []zap.Option{
		zap.AddCallerSkip(1), // Skip the logger wrapper
	}

	if !config.DisableCaller {
		opts = append(opts, zap.AddCaller())
	}

	if !config.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// Create zap logger
	zapLogger := zap.New(core, opts...)

	// Add service information as default fields
	zapLogger = zapLogger.With(
		zap.String("service", config.ServiceName),
		zap.String("version", config.ServiceVersion),
	)

	return &Logger{
		SugaredLogger: zapLogger.Sugar(),
		config:        config,
		zapLogger:     zapLogger,
	}, nil
}

// SetDefault sets the default logger instance
func SetDefault(logger *Logger) {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	defaultLogger = logger
}

// Default returns the default logger instance
func Default() *Logger {
	loggerMutex.RLock()
	defer loggerMutex.RUnlock()
	if defaultLogger == nil {
		// Create a default logger if none exists
		logger, err := New(DefaultConfig())
		if err != nil {
			panic(fmt.Sprintf("failed to create default logger: %v", err))
		}
		defaultLogger = logger
	}
	return defaultLogger
}

// WithContext creates a new logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := make([]interface{}, 0)

	if correlationID := ctx.Value(CorrelationIDKey); correlationID != nil {
		fields = append(fields, "correlation_id", correlationID)
	}

	if userID := ctx.Value(UserIDKey); userID != nil {
		fields = append(fields, "user_id", userID)
	}

	if requestID := ctx.Value(RequestIDKey); requestID != nil {
		fields = append(fields, "request_id", requestID)
	}

	if len(fields) == 0 {
		return l
	}

	return &Logger{
		SugaredLogger: l.SugaredLogger.With(fields...),
		config:        l.config,
		zapLogger:     l.zapLogger,
	}
}

// WithFields creates a new logger with additional fields
func (l *Logger) WithFields(fields Fields) *Logger {
	if len(fields) == 0 {
		return l
	}

	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	return &Logger{
		SugaredLogger: l.SugaredLogger.With(args...),
		config:        l.config,
		zapLogger:     l.zapLogger,
	}
}

// WithError creates a new logger with an error field
func (l *Logger) WithError(err error) *Logger {
	return l.WithFields(Fields{"error": err.Error()})
}

// LogUserAction logs a user action with standardized fields
func (l *Logger) LogUserAction(userID, action, resource string, metadata Fields) {
	fields := Fields{
		"user_id":   userID,
		"action":    action,
		"resource":  resource,
		"timestamp": time.Now(),
	}

	// Merge with provided metadata
	for k, v := range metadata {
		fields[k] = v
	}

	l.WithFields(fields).Info("user action")
}

// LogHTTPRequest logs an HTTP request with standardized fields
func (l *Logger) LogHTTPRequest(method, path string, statusCode int, duration time.Duration, fields Fields) {
	logFields := Fields{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
		"timestamp":   time.Now(),
	}

	// Merge with provided fields
	for k, v := range fields {
		logFields[k] = v
	}

	logger := l.WithFields(logFields)

	// Log at appropriate level based on status code
	if statusCode >= 500 {
		logger.Error("http request")
	} else if statusCode >= 400 {
		logger.Warn("http request")
	} else {
		logger.Info("http request")
	}
}

// LogDatabaseQuery logs a database query with standardized fields
func (l *Logger) LogDatabaseQuery(query string, duration time.Duration, err error) {
	fields := Fields{
		"query":       query,
		"duration_ms": duration.Milliseconds(),
		"timestamp":   time.Now(),
	}

	if err != nil {
		fields["error"] = err.Error()
		l.WithFields(fields).Error("database query failed")
	} else {
		l.WithFields(fields).Debug("database query")
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}

// Close closes the logger and flushes any buffered entries
func (l *Logger) Close() error {
	return l.Sync()
}

// Package-level convenience functions using default logger

// Info logs at info level using default logger
func Info(args ...interface{}) {
	Default().Info(args...)
}

// Infof logs at info level with formatting using default logger
func Infof(format string, args ...interface{}) {
	Default().Infof(format, args...)
}

// Warn logs at warn level using default logger
func Warn(args ...interface{}) {
	Default().Warn(args...)
}

// Warnf logs at warn level with formatting using default logger
func Warnf(format string, args ...interface{}) {
	Default().Warnf(format, args...)
}

// Error logs at error level using default logger
func Error(args ...interface{}) {
	Default().Error(args...)
}

// Errorf logs at error level with formatting using default logger
func Errorf(format string, args ...interface{}) {
	Default().Errorf(format, args...)
}

// Debug logs at debug level using default logger
func Debug(args ...interface{}) {
	Default().Debug(args...)
}

// Debugf logs at debug level with formatting using default logger
func Debugf(format string, args ...interface{}) {
	Default().Debugf(format, args...)
}

// WithFields creates a logger with fields using default logger
func WithFields(fields Fields) *Logger {
	return Default().WithFields(fields)
}

// WithContext creates a logger with context using default logger
func WithContext(ctx context.Context) *Logger {
	return Default().WithContext(ctx)
}

// WithError creates a logger with error using default logger
func WithError(err error) *Logger {
	return Default().WithError(err)
}