package analyzer

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type AnalysisResult struct {
	URL               string            `json:"url"`
	HTMLVersion       string            `json:"html_version"`
	PageTitle         string            `json:"page_title"`
	HeadingCounts     map[string]int    `json:"heading_counts"`
	InternalLinks     int               `json:"internal_links"`
	ExternalLinks     int               `json:"external_links"`
	InaccessibleLinks int               `json:"inaccessible_links"`
	HasLoginForm      bool              `json:"has_login_form"`
	Error             *AnalysisError    `json:"error,omitempty"`
	StatusCode        int               `json:"status_code,omitempty"`
}

// CacheEntry represents a cached analysis result
type CacheEntry struct {
	Result    *AnalysisResult
	Timestamp time.Time
	TTL       time.Duration
}

// Analyzer struct with caching
type Analyzer struct {
	httpClient     *http.Client
	timeout        time.Duration
	logger         *log.Logger
	circuitBreaker *CircuitBreaker
	// Performance optimization fields
	httpClientPool *sync.Pool
	metrics        *AnalyzerMetrics
	// Caching
	cache          map[string]*CacheEntry
	cacheMutex     sync.RWMutex
	cacheTTL       time.Duration
}

// AnalyzerMetrics tracks performance metrics
type AnalyzerMetrics struct {
	mu              sync.RWMutex
	TotalRequests   int64
	ActiveRequests  int64
	TotalDuration   time.Duration
	AvgDuration     time.Duration
	CacheHits       int64
	CacheMisses     int64
}

func NewAnalyzer(timeout time.Duration) *Analyzer {
	// Create optimized transport
	transport := &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    false, // Enable gzip compression
		ForceAttemptHTTP2:     true,  // Force HTTP/2 when possible
		// Connection pooling optimizations
		MaxConnsPerHost:       100,
		DisableKeepAlives:     false,
	}

	// Create HTTP client with optimized transport
	httpClient := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	// Create HTTP client pool for concurrent operations
	httpClientPool := &sync.Pool{
		New: func() interface{} {
			return &http.Client{
				Timeout:   timeout,
				Transport: transport,
			}
		},
	}

	analyzer := &Analyzer{
		httpClient:     httpClient,
		timeout:        timeout,
		logger:         log.New(os.Stdout, "analyzer ", log.LstdFlags),
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second, 2),
		httpClientPool: httpClientPool,
		metrics:        &AnalyzerMetrics{},
		cache:          make(map[string]*CacheEntry),
		cacheTTL:       5 * time.Minute, // Cache for 5 minutes
	}

	// Start background cache cleanup goroutine
	go analyzer.startCacheCleanup()

	return analyzer
}

// startCacheCleanup starts a background goroutine to clean expired cache entries
func (a *Analyzer) startCacheCleanup() {
	ticker := time.NewTicker(1 * time.Minute) // Clean every minute
	defer ticker.Stop()

	for range ticker.C {
		a.clearExpiredCache()
		if a.logger != nil {
			a.logger.Printf("cache_cleanup_completed entries_remaining=%d", len(a.cache))
		}
	}
}

// generateCacheKey generates a cache key for a URL
func (a *Analyzer) generateCacheKey(url string) string {
	hash := md5.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

// getFromCache retrieves a result from cache if valid
func (a *Analyzer) getFromCache(url string) (*AnalysisResult, bool) {
	key := a.generateCacheKey(url)
	
	a.cacheMutex.RLock()
	defer a.cacheMutex.RUnlock()
	
	entry, exists := a.cache[key]
	if !exists {
		return nil, false
	}
	
	// Check if cache entry is still valid
	if time.Since(entry.Timestamp) > entry.TTL {
		// Entry expired, remove it
		a.cacheMutex.RUnlock()
		a.cacheMutex.Lock()
		delete(a.cache, key)
		a.cacheMutex.Unlock()
		a.cacheMutex.RLock()
		return nil, false
	}
	
	// Update cache hit metrics
	a.metrics.mu.Lock()
	a.metrics.CacheHits++
	a.metrics.mu.Unlock()
	
	return entry.Result, true
}

// setCache stores a result in cache
func (a *Analyzer) setCache(url string, result *AnalysisResult) {
	key := a.generateCacheKey(url)
	
	a.cacheMutex.Lock()
	defer a.cacheMutex.Unlock()
	
	a.cache[key] = &CacheEntry{
		Result:    result,
		Timestamp: time.Now(),
		TTL:       a.cacheTTL,
	}
	
	// Update cache miss metrics
	a.metrics.mu.Lock()
	a.metrics.CacheMisses++
	a.metrics.mu.Unlock()
}

// clearExpiredCache removes expired cache entries
func (a *Analyzer) clearExpiredCache() {
	a.cacheMutex.Lock()
	defer a.cacheMutex.Unlock()
	
	now := time.Now()
	for key, entry := range a.cache {
		if now.Sub(entry.Timestamp) > entry.TTL {
			delete(a.cache, key)
		}
	}
}

// getHTTPClient gets an HTTP client from the pool
func (a *Analyzer) getHTTPClient() *http.Client {
	return a.httpClientPool.Get().(*http.Client)
}

// putHTTPClient returns an HTTP client to the pool
func (a *Analyzer) putHTTPClient(client *http.Client) {
	a.httpClientPool.Put(client)
}

// SetLogger allows tests or callers to provide a custom logger
func (a *Analyzer) SetLogger(logger *log.Logger) {
	if logger != nil {
		a.logger = logger
	}
}

// GetMetrics returns a copy of current metrics
func (a *Analyzer) GetMetrics() AnalyzerMetrics {
	a.metrics.mu.RLock()
	defer a.metrics.mu.RUnlock()
	
	return AnalyzerMetrics{
		TotalRequests:  a.metrics.TotalRequests,
		ActiveRequests: a.metrics.ActiveRequests,
		TotalDuration:  a.metrics.TotalDuration,
		AvgDuration:    a.metrics.AvgDuration,
		CacheHits:      a.metrics.CacheHits,
		CacheMisses:    a.metrics.CacheMisses,
	}
}

// updateMetrics updates performance metrics
func (a *Analyzer) updateMetrics(duration time.Duration) {
	a.metrics.mu.Lock()
	defer a.metrics.mu.Unlock()
	
	a.metrics.TotalRequests++
	a.metrics.TotalDuration += duration
	a.metrics.AvgDuration = a.metrics.TotalDuration / time.Duration(a.metrics.TotalRequests)
}

// incrementActiveRequests increments the active request counter
func (a *Analyzer) incrementActiveRequests() {
	a.metrics.mu.Lock()
	defer a.metrics.mu.Unlock()
	a.metrics.ActiveRequests++
}

// decrementActiveRequests decrements the active request counter
func (a *Analyzer) decrementActiveRequests() {
	a.metrics.mu.Lock()
	defer a.metrics.mu.Unlock()
	a.metrics.ActiveRequests--
}

func (a *Analyzer) AnalyzeURL(targetURL string) *AnalysisResult {
	return a.AnalyzeURLWithContext(context.Background(), targetURL)
}

func (a *Analyzer) AnalyzeURLWithContext(ctx context.Context, targetURL string) *AnalysisResult {
	// Track active requests and overall metrics
	a.incrementActiveRequests()
	defer a.decrementActiveRequests()
	
	overallStart := time.Now()
	defer func() {
		a.updateMetrics(time.Since(overallStart))
	}()

	result := &AnalysisResult{
		URL:           targetURL,
		HeadingCounts: make(map[string]int),
	}

	if a.logger != nil {
		a.logger.Printf("analyze_start url=%q", targetURL)
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		// If parsing failed and there's no scheme, try normalizing with https://
		if !strings.Contains(targetURL, "://") {
			normalized := "https://" + targetURL
			parsedURL, err = url.Parse(normalized)
			if err != nil {
				result.Error = NewInvalidURLError(targetURL, err)
				if a.logger != nil {
					a.logger.Printf("invalid_url url=%q err=%v", targetURL, err)
				}
				return result
			}
			if a.logger != nil {
				a.logger.Printf("url_normalized url=%q", normalized)
			}
			targetURL = normalized
		} else {
			result.Error = NewInvalidURLError(targetURL, err)
			if a.logger != nil {
				a.logger.Printf("invalid_url url=%q err=%v", targetURL, err)
			}
			return result
		}
	}

	if parsedURL.Scheme == "" {
		targetURL = "https://" + targetURL
		parsedURL, err = url.Parse(targetURL)
		if err != nil {
			result.Error = NewInvalidURLError(targetURL, err)
			if a.logger != nil {
				a.logger.Printf("invalid_url url=%q err=%v", targetURL, err)
			}
			return result
		}
	}

	// Check cache first
	if cachedResult, found := a.getFromCache(targetURL); found {
		if a.logger != nil {
			a.logger.Printf("analyze_cache_hit url=%q", targetURL)
		}
		return cachedResult
	}

	// Create child context with shorter timeout for HTTP operations
	httpCtx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	// Check context cancellation before making request
	select {
	case <-httpCtx.Done():
		result.Error = NewTimeoutError(targetURL, a.timeout)
		return result
	default:
	}

	// Use circuit breaker with context
	err = a.circuitBreaker.Execute(func() error {
		req, err := http.NewRequestWithContext(httpCtx, "GET", targetURL, nil)
		if err != nil {
			return err
		}

		// Use client from pool for concurrent operations
		client := a.getHTTPClient()
		defer a.putHTTPClient(client)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		result.StatusCode = resp.StatusCode

		// Check for HTTP errors
		if resp.StatusCode >= 400 {
			return NewHTTPError(resp.StatusCode, "HTTP request failed")
		}

		// Read body with context cancellation check
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		// Check context cancellation before parsing
		select {
		case <-httpCtx.Done():
			return httpCtx.Err()
		default:
		}

		// Parse HTML with context
		doc, err := html.Parse(strings.NewReader(string(body)))
		if err != nil {
			return NewParseError("Failed to parse HTML", err)
		}

		// Analyze document with context
		a.analyzeDocumentWithContext(httpCtx, doc, result, parsedURL, string(body))
		return nil
	})

	if err != nil {
		// Convert error to AnalysisError if needed
		if analysisErr, ok := err.(*AnalysisError); ok {
			result.Error = analysisErr
		} else {
			result.Error = NewAnalysisError(ErrCodeInternalError, "Analysis failed").WithCause(err)
		}
		if a.logger != nil {
			a.logger.Printf("analyze_error url=%q err=%v", targetURL, err)
		}
		return result
	}

	// Cache the result
	a.setCache(targetURL, result)

	if a.logger != nil {
		a.logger.Printf("analyze_done url=%q total_ms=%d internal=%d external=%d inaccessible=%d headings=%d login_form=%t html_version=%q title_len=%d", 
			result.URL, time.Since(overallStart).Milliseconds(), result.InternalLinks, 
			result.ExternalLinks, result.InaccessibleLinks, len(result.HeadingCounts), 
			result.HasLoginForm, result.HTMLVersion, len(result.PageTitle))
	}

	return result
}

func (a *Analyzer) analyzeDocument(doc *html.Node, result *AnalysisResult, baseURL *url.URL, htmlContent string) {
	result.HTMLVersion = a.detectHTMLVersion(htmlContent)
	if a.logger != nil {
		a.logger.Printf("html_version_detected url=%q version=%q", result.URL, result.HTMLVersion)
	}

	var traverse func(*html.Node)
	var links []string

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					result.PageTitle = strings.TrimSpace(n.FirstChild.Data)
				}
			case "h1", "h2", "h3", "h4", "h5", "h6":
				result.HeadingCounts[n.Data]++
			case "a":
				for _, attr := range n.Attr {
					if attr.Key == "href" && attr.Val != "" {
						links = append(links, attr.Val)
					}
				}
			case "form":
				if a.isLoginForm(n) {
					result.HasLoginForm = true
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	a.analyzeLinksConcurrent(links, baseURL, result)
}

func (a *Analyzer) analyzeDocumentWithContext(ctx context.Context, doc *html.Node, result *AnalysisResult, baseURL *url.URL, htmlContent string) {
	result.HTMLVersion = a.detectHTMLVersion(htmlContent)
	if a.logger != nil {
		a.logger.Printf("html_version_detected url=%q version=%q", result.URL, result.HTMLVersion)
	}

	var traverse func(*html.Node)
	var links []string

	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					result.PageTitle = strings.TrimSpace(n.FirstChild.Data)
				}
			case "h1", "h2", "h3", "h4", "h5", "h6":
				result.HeadingCounts[n.Data]++
			case "a":
				for _, attr := range n.Attr {
					if attr.Key == "href" && attr.Val != "" {
						links = append(links, attr.Val)
					}
				}
			case "form":
				if a.isLoginForm(n) {
					result.HasLoginForm = true
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	a.analyzeLinksConcurrent(links, baseURL, result)
}

func (a *Analyzer) detectHTMLVersion(htmlContent string) string {
	htmlContent = strings.TrimSpace(strings.ToLower(htmlContent))
	
	if strings.Contains(htmlContent, "<!doctype html>") {
		return "HTML5"
	}
	
	if strings.Contains(htmlContent, `"-//w3c//dtd xhtml 1.1//en"`) {
		return "XHTML 1.1"
	}
	if strings.Contains(htmlContent, `"-//w3c//dtd xhtml 1.0 strict//en"`) {
		return "XHTML 1.0 Strict"
	}
	if strings.Contains(htmlContent, `"-//w3c//dtd xhtml 1.0 transitional//en"`) {
		return "XHTML 1.0 Transitional"
	}
	if strings.Contains(htmlContent, `"-//w3c//dtd xhtml 1.0 frameset//en"`) {
		return "XHTML 1.0 Frameset"
	}
	
	if strings.Contains(htmlContent, `"-//w3c//dtd html 4.01//en"`) {
		return "HTML 4.01 Strict"
	}
	if strings.Contains(htmlContent, `"-//w3c//dtd html 4.01 transitional//en"`) {
		return "HTML 4.01 Transitional"
	}
	if strings.Contains(htmlContent, `"-//w3c//dtd html 4.01 frameset//en"`) {
		return "HTML 4.01 Frameset"
	}
	
	if strings.Contains(htmlContent, `"-//w3c//dtd html 3.2 final//en"`) {
		return "HTML 3.2"
	}
	
	if strings.Contains(htmlContent, `"-//ietf//dtd html 2.0//en"`) {
		return "HTML 2.0"
	}
	
	return "Unknown"
}

// LinkResult represents the result of analyzing a single link
type LinkResult struct {
	URL         string
	IsInternal  bool
	IsAccessible bool
	Error       error
}

// AnalysisJob represents a job for the worker pool
type AnalysisJob struct {
	URL     string
	BaseURL *url.URL
}

// AnalysisWorkerPool manages concurrent link analysis
type AnalysisWorkerPool struct {
	workers    int
	jobQueue   chan AnalysisJob
	resultChan chan LinkResult
	wg         sync.WaitGroup
	analyzer   *Analyzer
}

// NewAnalysisWorkerPool creates a new worker pool for link analysis
func NewAnalysisWorkerPool(workers int, analyzer *Analyzer) *AnalysisWorkerPool {
	return &AnalysisWorkerPool{
		workers:    workers,
		jobQueue:   make(chan AnalysisJob, workers*2),
		resultChan: make(chan LinkResult, workers*2),
		analyzer:   analyzer,
	}
}

// Start starts the worker pool
func (wp *AnalysisWorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

// Stop stops the worker pool
func (wp *AnalysisWorkerPool) Stop() {
	close(wp.jobQueue)
	wp.wg.Wait()
	close(wp.resultChan)
}

// SubmitJob submits a job to the worker pool
func (wp *AnalysisWorkerPool) SubmitJob(job AnalysisJob) {
	wp.jobQueue <- job
}

// GetResults returns a channel to receive results
func (wp *AnalysisWorkerPool) GetResults() <-chan LinkResult {
	return wp.resultChan
}

// worker processes jobs from the queue
func (wp *AnalysisWorkerPool) worker() {
	defer wp.wg.Done()
	
	for job := range wp.jobQueue {
		result := wp.analyzer.analyzeSingleLink(job.URL, job.BaseURL)
		wp.resultChan <- result
	}
}

// analyzeSingleLink analyzes a single link (used by workers)
func (a *Analyzer) analyzeSingleLink(link string, baseURL *url.URL) LinkResult {
	link = strings.TrimSpace(link)
	if link == "" || strings.HasPrefix(link, "#") || strings.HasPrefix(link, "javascript:") || 
	   strings.HasPrefix(link, "mailto:") || strings.HasPrefix(link, "tel:") {
		return LinkResult{URL: link, IsInternal: false, IsAccessible: false}
	}

	parsedLink, err := url.Parse(link)
	if err != nil {
		return LinkResult{URL: link, IsInternal: false, IsAccessible: false, Error: err}
	}

	resolvedLink := baseURL.ResolveReference(parsedLink)
	isInternal := resolvedLink.Host == baseURL.Host
	isAccessible := a.isLinkAccessible(resolvedLink.String())

	return LinkResult{
		URL:         link,
		IsInternal:  isInternal,
		IsAccessible: isAccessible,
	}
}

// analyzeLinksConcurrent analyzes links using concurrent workers
func (a *Analyzer) analyzeLinksConcurrent(links []string, baseURL *url.URL, result *AnalysisResult) {
	if len(links) == 0 {
		return
	}

	start := time.Now()
	
	// Determine optimal number of workers based on link count
	workers := 10 // Default
	if len(links) < workers {
		workers = len(links)
	}
	if workers > 20 {
		workers = 20 // Cap at 20 workers
	}

	// Create and start worker pool
	pool := NewAnalysisWorkerPool(workers, a)
	pool.Start()
	defer pool.Stop()

	// Submit all jobs
	go func() {
		for _, link := range links {
			pool.SubmitJob(AnalysisJob{URL: link, BaseURL: baseURL})
		}
	}()

	// Collect results
	total := 0
	skipped := 0
	
	// Use a timeout context for link analysis
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Process results with timeout
	resultsReceived := 0
	for resultsReceived < len(links) {
		select {
		case linkResult := <-pool.GetResults():
			resultsReceived++
			total++
			
			if linkResult.IsInternal {
				result.InternalLinks++
			} else {
				result.ExternalLinks++
			}
			
			if !linkResult.IsAccessible {
				result.InaccessibleLinks++
			}
			
		case <-ctx.Done():
			if a.logger != nil {
				a.logger.Printf("link_analysis_timeout url=%q processed=%d total=%d", 
					result.URL, resultsReceived, len(links))
			}
			goto done
		}
	}

done:
	if a.logger != nil {
		a.logger.Printf("links_analyzed_concurrent url=%q total=%d skipped=%d internal=%d external=%d inaccessible=%d ms=%d workers=%d", 
			result.URL, total, skipped, result.InternalLinks, result.ExternalLinks, 
			result.InaccessibleLinks, time.Since(start).Milliseconds(), workers)
	}
}

func (a *Analyzer) isLinkAccessible(link string) bool {
	req, err := http.NewRequest("HEAD", link, nil)
	if err != nil {
		if a.logger != nil {
			a.logger.Printf("link_head_build_error link=%q err=%v", link, err)
		}
		return false
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		if a.logger != nil {
			a.logger.Printf("link_head_error link=%q err=%v", link, err)
		}
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		if a.logger != nil {
			a.logger.Printf("link_inaccessible link=%q status=%d", link, resp.StatusCode)
		}
		return false
	}
	return true
}

func (a *Analyzer) isLoginForm(formNode *html.Node) bool {
	hasPasswordField := false
	hasUsernameField := false
	
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "input" {
			var inputType, inputName, inputId, inputPlaceholder string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "type":
					inputType = strings.ToLower(attr.Val)
				case "name":
					inputName = strings.ToLower(attr.Val)
				case "id":
					inputId = strings.ToLower(attr.Val)
				case "placeholder":
					inputPlaceholder = strings.ToLower(attr.Val)
				}
			}
			
			// Check for password field
			if inputType == "password" {
				hasPasswordField = true
			}
			
			// Check for username/email field with more flexible detection
			if inputType == "email" || inputType == "text" || inputType == "" || inputType == "tel" {
				// Check name attribute
				if strings.Contains(inputName, "user") || 
				   strings.Contains(inputName, "login") || 
				   strings.Contains(inputName, "email") ||
				   strings.Contains(inputName, "account") ||
				   strings.Contains(inputName, "phone") {
					hasUsernameField = true
				}
				
				// Check id attribute
				if strings.Contains(inputId, "user") || 
				   strings.Contains(inputId, "login") || 
				   strings.Contains(inputId, "email") ||
				   strings.Contains(inputId, "account") ||
				   strings.Contains(inputId, "phone") {
					hasUsernameField = true
				}
				
				// Check placeholder attribute
				if strings.Contains(inputPlaceholder, "user") || 
				   strings.Contains(inputPlaceholder, "login") || 
				   strings.Contains(inputPlaceholder, "email") ||
				   strings.Contains(inputPlaceholder, "account") ||
				   strings.Contains(inputPlaceholder, "phone") ||
				   strings.Contains(inputPlaceholder, "username") {
					hasUsernameField = true
				}
			}
		}
		
		// Also check for common login-related elements
		if n.Type == html.ElementNode && n.Data == "button" {
			// Check button text content
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				buttonText := strings.ToLower(strings.TrimSpace(n.FirstChild.Data))
				if strings.Contains(buttonText, "login") || 
				   strings.Contains(buttonText, "sign in") ||
				   strings.Contains(buttonText, "signin") ||
				   strings.Contains(buttonText, "log in") {
					// If we find a login button, be more lenient about username field detection
					if hasPasswordField {
						hasUsernameField = true
					}
				}
			}
		}
		
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	
	traverse(formNode)
	
	// More flexible detection: if we have a password field and some indication of username/email
	// or if we have both fields explicitly
	return hasPasswordField && hasUsernameField
}