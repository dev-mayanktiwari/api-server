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

type Logger struct {
	*zap.Logger
}

type Config struct {
	Level  string 
	Format string 
}

func New(config Config) (*Logger, error) {
	level, err := parseLogLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

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

	var encoder zapcore.Encoder
	switch strings.ToLower(config.Format) {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "json-pretty":
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
		prettyConfig := encoderConfig
		encoder = newPrettyJSONEncoder(prettyConfig)
	case "console":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = customTimeEncoder
		encoderConfig.EncodeCaller = customCallerEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		return nil, fmt.Errorf("unsupported log format: %s", config.Format)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{Logger: zapLogger}, nil
}

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

func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{Logger: l.With(zap.Any(key, value))}
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return &Logger{Logger: l.With(zapFields...)}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{Logger: l.With(zap.Error(err))}
}

func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.With(zap.String("request_id", requestID))}
}

var globalLogger *Logger

func InitGlobal(config Config) error {
	logger, err := New(config)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

func GetGlobal() *Logger {
	if globalLogger == nil {
		logger, _ := New(Config{Level: "info", Format: "console"})
		return logger
	}
	return globalLogger
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("15:04:05.000"))
}

func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%s:%d", caller.TrimmedPath(), caller.Line))
}

func newPrettyJSONEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	return &prettyJSONEncoder{
		Encoder: zapcore.NewJSONEncoder(cfg),
	}
}

type prettyJSONEncoder struct {
	zapcore.Encoder
}

func (enc *prettyJSONEncoder) Clone() zapcore.Encoder {
	return &prettyJSONEncoder{Encoder: enc.Encoder.Clone()}
}

func (enc *prettyJSONEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf, err := enc.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	var jsonObj map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &jsonObj); err != nil {
		return buf, nil
	}

	prettyBuf := buffer.NewPool().Get()
	prettyBytes, err := json.MarshalIndent(jsonObj, "", "  ")
	if err != nil {
		return buf, nil
	}

	prettyBuf.Write(prettyBytes)
	prettyBuf.AppendString("\n")
	
	return prettyBuf, nil
}