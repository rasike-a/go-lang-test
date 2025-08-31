package analyzer

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"sync"
	"time"
)

// CacheManager handles caching operations for analysis results
type CacheManager struct {
	cache      map[string]*CacheEntry
	mutex      sync.RWMutex
	ttl        time.Duration
	logger     *log.Logger
	cleanupTicker *time.Ticker
	stopChan   chan struct{}
}

// NewCacheManager creates a new cache manager
func NewCacheManager(ttl time.Duration, logger *log.Logger) *CacheManager {
	cm := &CacheManager{
		cache:      make(map[string]*CacheEntry),
		ttl:        ttl,
		logger:     logger,
		stopChan:   make(chan struct{}),
	}
	
	// Start cache cleanup goroutine
	cm.startCleanup()
	
	return cm
}

// startCleanup starts the background cache cleanup process
func (cm *CacheManager) startCleanup() {
	cm.cleanupTicker = time.NewTicker(1 * time.Minute)
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
	
	cm.logger.Printf("analyzer cache_hit url=%q", url)
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
	
	cm.logger.Printf("analyzer cache_set url=%q", url)
}

// clearExpired removes expired cache entries
func (cm *CacheManager) clearExpired() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	now := time.Now()
	
	for key, entry := range cm.cache {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(cm.cache, key)
		}
	}
	
	remainingCount := len(cm.cache)
	cm.logger.Printf("analyzer cache_cleanup_completed entries_remaining=%d", remainingCount)
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
