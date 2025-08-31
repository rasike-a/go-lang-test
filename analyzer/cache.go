package analyzer

import (
	"crypto/md5"
	"encoding/hex"
	"sync"
	"time"

	"web-page-analyzer/logger"
)

// CacheManager handles caching operations for analysis results
type CacheManager struct {
	cache         map[string]*CacheEntry
	mutex         sync.RWMutex
	ttl           time.Duration
	cleanupTicker *time.Ticker
	stopChan      chan struct{}
	verbose       bool // Control logging verbosity
}

// NewCacheManager creates a new cache manager
func NewCacheManager(ttl time.Duration) *CacheManager {
	cm := &CacheManager{
		cache:    make(map[string]*CacheEntry),
		ttl:      ttl,
		stopChan: make(chan struct{}),
		verbose:  false, // Default to quiet logging
	}
	cm.startCleanup()
	return cm
}

// startCleanup starts the background cache cleanup process
func (cm *CacheManager) startCleanup() {
	// Run cleanup every 5 minutes instead of every minute to reduce log noise
	cm.cleanupTicker = time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-cm.cleanupTicker.C:
				cm.clearExpired()
			case <-cm.stopChan:
				cm.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// Stop stops the cache manager and cleanup processes
func (cm *CacheManager) Stop() {
	close(cm.stopChan)
	if cm.cleanupTicker != nil {
		cm.cleanupTicker.Stop()
	}
}

// generateCacheKey generates an MD5 hash for the URL to use as cache key
func (cm *CacheManager) generateCacheKey(url string) string {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

// SetVerbose enables or disables verbose logging
func (cm *CacheManager) SetVerbose(verbose bool) {
	cm.verbose = verbose
}

// Get retrieves a result from cache if it exists and is not expired
func (cm *CacheManager) Get(url string) (*AnalysisResult, bool) {
	key := cm.generateCacheKey(url)
	
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	entry, exists := cm.cache[key]
	if !exists {
		return nil, false
	}
	
	// Check if entry has expired
	if time.Since(entry.Timestamp) > entry.TTL {
		// Entry expired, remove it
		cm.mutex.RUnlock()
		cm.mutex.Lock()
		delete(cm.cache, key)
		cm.mutex.Unlock()
		cm.mutex.RLock()
		return nil, false
	}
	
	if cm.verbose {
		logger.WithCache("hit", url).Info("Cache hit")
	}
	return entry.Result, true
}

// Set stores a result in the cache
func (cm *CacheManager) Set(url string, result *AnalysisResult) {
	key := cm.generateCacheKey(url)
	
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.cache[key] = &CacheEntry{
		Result:    result,
		Timestamp: time.Now(),
		TTL:       cm.ttl,
	}
	
	if cm.verbose {
		logger.WithCache("set", url).Info("Cache set")
	}
}

// clearExpired removes expired cache entries
func (cm *CacheManager) clearExpired() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	now := time.Now()
	expiredCount := 0
	
	for key, entry := range cm.cache {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(cm.cache, key)
			expiredCount++
		}
	}
	
	remainingCount := len(cm.cache)
	
	// Only log if we actually removed expired entries or if cache is getting large
	if expiredCount > 0 {
		logger.WithComponent("cache").Infow("Cache cleanup completed",
			"expired_removed", expiredCount,
			"entries_remaining", remainingCount,
		)
	} else if cm.verbose && remainingCount > 10 {
		// Log occasionally when cache is large but no cleanup needed (only in verbose mode)
		logger.WithComponent("cache").Infow("Cache status",
			"entries", remainingCount,
			"status", "no_expired",
		)
	}
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats() (int, int) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	total := len(cm.cache)
	expired := 0
	now := time.Now()
	
	for _, entry := range cm.cache {
		if now.Sub(entry.Timestamp) > entry.TTL {
			expired++
		}
	}
	
	return total, expired
}
