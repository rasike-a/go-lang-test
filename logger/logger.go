package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger is the global logger instance
	Logger *zap.Logger
	// Sugar is the sugared logger for easier usage
	Sugar *zap.SugaredLogger
)

// Init initializes the global logger
func Init() {
	// Check if we're in development mode
	isDevelopment := os.Getenv("ENV") == "development"

	var config zap.Config
	if isDevelopment {
		// Development: Human-readable console output
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		// Production: JSON output
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		config.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	// Create logger
	var err error
	Logger, err = config.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// Create sugared logger for easier usage
	Sugar = Logger.Sugar()

	// Log initialization
	format := "json"
	if isDevelopment {
		format = "console"
	}
	Sugar.Info("Logger initialized",
		"environment", os.Getenv("ENV"),
		"format", format,
	)
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		if err := Logger.Sync(); err != nil {
			// Logger sync errors are typically not critical and can be ignored
			// in most cases, but we could log them if needed
			_ = err
		}
	}
}

// WithFields creates a logger with predefined fields
func WithFields(fields map[string]interface{}) *zap.SugaredLogger {
	if Sugar == nil {
		Init()
	}

	// Convert map to key-value pairs for Sugar.With()
	var args []interface{}
	for k, v := range fields {
		args = append(args, k, v)
	}

	return Sugar.With(args...)
}

// WithComponent creates a logger with component field
func WithComponent(component string) *zap.SugaredLogger {
	return WithFields(map[string]interface{}{
		"component": component,
	})
}

// WithRequest creates a logger with HTTP request fields
func WithRequest(method, path, remoteAddr, userAgent string) *zap.SugaredLogger {
	return WithFields(map[string]interface{}{
		"method":      method,
		"path":        path,
		"remote_addr": remoteAddr,
		"user_agent":  userAgent,
	})
}

// WithAnalysis creates a logger with analysis-specific fields
func WithAnalysis(url string) *zap.SugaredLogger {
	return WithFields(map[string]interface{}{
		"component": "analyzer",
		"url":       url,
	})
}

// WithCache creates a logger with cache-specific fields
func WithCache(operation, url string) *zap.SugaredLogger {
	return WithFields(map[string]interface{}{
		"component":       "cache",
		"cache_operation": operation,
		"url":             url,
	})
}

// WithMetrics creates a logger with metrics-specific fields
func WithMetrics(operation string) *zap.SugaredLogger {
	return WithFields(map[string]interface{}{
		"component":         "metrics",
		"metrics_operation": operation,
	})
}

// WithCircuitBreaker creates a logger with circuit breaker fields
func WithCircuitBreaker(state, operation string) *zap.SugaredLogger {
	return WithFields(map[string]interface{}{
		"component":             "circuit_breaker",
		"circuit_breaker_state": state,
		"operation":             operation,
	})
}
