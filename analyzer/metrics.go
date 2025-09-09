package analyzer

import (
	"sync"
	"time"
)

// MetricsManager handles performance metrics collection and reporting
type MetricsManager struct {
	mu             sync.RWMutex
	TotalRequests  int64
	ActiveRequests int64
	TotalDuration  time.Duration
	AvgDuration    time.Duration
	CacheHits      int64
	CacheMisses    int64
}

// NewMetricsManager creates a new metrics manager
func NewMetricsManager() *MetricsManager {
	return &MetricsManager{}
}

// GetMetrics returns a copy of current metrics
func (mm *MetricsManager) GetMetrics() MetricsManager {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	return MetricsManager{
		TotalRequests:  mm.TotalRequests,
		ActiveRequests: mm.ActiveRequests,
		TotalDuration:  mm.TotalDuration,
		AvgDuration:    mm.AvgDuration,
		CacheHits:      mm.CacheHits,
		CacheMisses:    mm.CacheMisses,
	}
}

// updateMetrics updates metrics with a new request duration
func (mm *MetricsManager) updateMetrics(duration time.Duration) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.TotalRequests++
	mm.TotalDuration += duration

	// Calculate running average
	if mm.TotalRequests > 0 {
		mm.AvgDuration = mm.TotalDuration / time.Duration(mm.TotalRequests)
	}
}

// incrementActiveRequests increments the active requests counter
func (mm *MetricsManager) incrementActiveRequests() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.ActiveRequests++
}

// decrementActiveRequests decrements the active requests counter
func (mm *MetricsManager) decrementActiveRequests() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.ActiveRequests--
}

// RecordCacheHit records a cache hit
func (mm *MetricsManager) RecordCacheHit() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.CacheHits++
}

// RecordCacheMiss records a cache miss
func (mm *MetricsManager) RecordCacheMiss() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.CacheMisses++
}

// Resets all metrics to zero
func (mm *MetricsManager) Reset() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.TotalRequests = 0
	mm.ActiveRequests = 0
	mm.TotalDuration = 0
	mm.AvgDuration = 0
	mm.CacheHits = 0
	mm.CacheMisses = 0
}
