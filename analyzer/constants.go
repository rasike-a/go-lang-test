package analyzer

import "time"

// Timeout constants
const (
	DefaultTimeout        = 60 * time.Second
	LinkCheckTimeout      = 3 * time.Second
	HTMLAnalysisTimeout   = 10 * time.Second
	CircuitBreakerTimeout = 60 * time.Second
	CacheCleanupInterval  = 5 * time.Minute
	CacheDefaultTTL       = 5 * time.Minute
)

// HTTP constants
const (
	MaxHeaderBytes = 1 << 20 // 1MB
	ReadTimeout    = 15 * time.Second
	WriteTimeout   = 15 * time.Second
	IdleTimeout    = 60 * time.Second
)

// Worker pool constants
const (
	BufferMultiplier = 4
	MinWorkers       = 4
	MaxWorkers       = 100
)

// Circuit breaker constants
const (
	DefaultFailureThreshold = 5
	DefaultSuccessThreshold = 2
)

// Cache constants
const (
	CacheCleanupIntervalMinutes = 5
	CacheVerboseThreshold       = 10
)
