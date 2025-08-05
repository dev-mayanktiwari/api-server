package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger to provide additional functionality
type Logger struct {
	*zap.Logger
}

// Config holds logger configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, console
}

// New creates a new logger instance
func New(config Config) (*Logger, error) {
	// Parse log level
	level, err := parseLogLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Choose encoder based on format
	var encoder zapcore.Encoder
	switch strings.ToLower(config.Format) {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "json-pretty":
		// Pretty JSON with indentation for development
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
		prettyConfig := encoderConfig
		encoder = newPrettyJSONEncoder(prettyConfig)
	case "console":
		// Beautiful console output with colors
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = customTimeEncoder
		encoderConfig.EncodeCaller = customCallerEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return nil, fmt.Errorf("unsupported log format: %s (supported: json, json-pretty, console)", config.Format)
	}

	// Create core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	// Create logger with caller info
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{Logger: zapLogger}, nil
}

// parseLogLevel converts string to zapcore.Level
func parseLogLevel(level string) (zapcore.Level, error) {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown level: %s", level)
	}
}

// WithField adds a single field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{Logger: l.With(zap.Any(key, value))}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return &Logger{Logger: l.With(zapFields...)}
}

// WithError adds an error field to the logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.With(zap.Error(err))}
}

// WithRequestID adds request ID to logger context
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.With(zap.String("request_id", requestID))}
}

// HTTP logging helpers

// LogHTTPRequest logs HTTP request details
func (l *Logger) LogHTTPRequest(method, path, userAgent, clientIP string, statusCode int, duration int64) {
	l.Info("HTTP Request",
		zap.String("method", method),
		zap.String("path", path),
		zap.String("user_agent", userAgent),
		zap.String("client_ip", clientIP),
		zap.Int("status_code", statusCode),
		zap.Int64("duration_ms", duration),
	)
}

// LogError logs error with context
func (l *Logger) LogError(message string, err error, fields ...zap.Field) {
	allFields := append(fields, zap.Error(err))
	l.Error(message, allFields...)
}

// Business logic helpers

// LogUserAction logs user actions for audit
func (l *Logger) LogUserAction(userID, action, resource string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.String("resource", resource),
	}

	for key, value := range metadata {
		fields = append(fields, zap.Any(key, value))
	}

	l.Info("User Action", fields...)
}

// LogDatabaseOperation logs database operations
func (l *Logger) LogDatabaseOperation(operation, table string, duration int64, err error) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Int64("duration_ms", duration),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Error("Database Operation Failed", fields...)
	} else {
		l.Debug("Database Operation", fields...)
	}
}

// Global logger instance (for convenience)
var globalLogger *Logger

// InitGlobal initializes the global logger
func InitGlobal(config Config) error {
	logger, err := New(config)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// GetGlobal returns the global logger instance
func GetGlobal() *Logger {
	if globalLogger == nil {
		// Fallback to a basic logger if not initialized
		logger, _ := New(Config{Level: "info", Format: "console"})
		return logger
	}
	return globalLogger
}

// Convenience functions for global logger
func Debug(msg string, fields ...zap.Field) {
	GetGlobal().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	GetGlobal().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	GetGlobal().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	GetGlobal().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	GetGlobal().Fatal(msg, fields...)
}

// Custom encoders for better formatting

// customTimeEncoder formats time in a more readable way
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("15:04:05.000"))
}

// customCallerEncoder formats caller information nicely
func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%s:%d", caller.TrimmedPath(), caller.Line))
}

// prettyJSONEncoder creates a JSON encoder with pretty printing
func newPrettyJSONEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	return &prettyJSONEncoder{
		Encoder: zapcore.NewJSONEncoder(cfg),
	}
}

// prettyJSONEncoder wraps the default JSON encoder to add pretty printing
type prettyJSONEncoder struct {
	zapcore.Encoder
}

// Clone implements zapcore.Encoder
func (enc *prettyJSONEncoder) Clone() zapcore.Encoder {
	return &prettyJSONEncoder{Encoder: enc.Encoder.Clone()}
}

// EncodeEntry implements zapcore.Encoder with pretty JSON formatting
func (enc *prettyJSONEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// Use the base encoder to get the JSON
	buf, err := enc.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	// Parse and re-format as pretty JSON
	var jsonObj map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &jsonObj); err != nil {
		// If parsing fails, return original
		return buf, nil
	}

	// Create a new buffer for pretty JSON
	prettyBuf := buffer.NewPool().Get()
	prettyBytes, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		// If pretty formatting fails, return original
		return buf, nil
	}

	prettyBuf.Write(prettyBytes)
	prettyBuf.AppendString("\n")
	
	return prettyBuf, nil
}