package analyzer

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
	"web-page-analyzer/logger"
)

// Analyzer is the main analyzer that orchestrates web page analysis
type Analyzer struct {
	httpClient     *http.Client
	timeout        time.Duration
	circuitBreaker *CircuitBreaker
	
	// Modular components
	cacheManager   *CacheManager
	metricsManager *MetricsManager
	httpClientPool *sync.Pool
}

// NewAnalyzer creates a new analyzer instance with optimized settings
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
		circuitBreaker: NewCircuitBreaker(5, 60*time.Second, 2), // Increased timeout for complex sites
		httpClientPool: httpClientPool,
		cacheManager:   NewCacheManager(5*time.Minute),
		metricsManager: NewMetricsManager(),
	}

	return analyzer
}

// SetLogger sets a custom logger for the analyzer (deprecated - now uses structured logging)
func (a *Analyzer) SetLogger(_ interface{}) {
	// This method is deprecated as we now use structured logging
	// The logger parameter is ignored
}

// SetCacheVerbose enables or disables verbose cache logging
func (a *Analyzer) SetCacheVerbose(verbose bool) {
	a.cacheManager.SetVerbose(verbose)
}

// GetMetrics returns current performance metrics
func (a *Analyzer) GetMetrics() MetricsManager {
	return a.metricsManager.GetMetrics()
}

// AnalyzeURL analyzes a URL without context (legacy method)
func (a *Analyzer) AnalyzeURL(targetURL string) *AnalysisResult {
	return a.AnalyzeURLWithContext(context.Background(), targetURL)
}

// AnalyzeURLWithContext analyzes a URL with context support
func (a *Analyzer) AnalyzeURLWithContext(ctx context.Context, targetURL string) *AnalysisResult {
	startTime := time.Now()
	
	// Track active requests
	a.metricsManager.incrementActiveRequests()
	defer a.metricsManager.decrementActiveRequests()
	
	// Check cache first
	if cachedResult, found := a.cacheManager.Get(targetURL); found {
		a.metricsManager.RecordCacheHit()
		return cachedResult
	}
	a.metricsManager.RecordCacheMiss()
	
	// Create result
	result := &AnalysisResult{
		URL:          targetURL,
		HeadingCounts: make(map[string]int),
	}
	
	// Validate and normalize URL
	parsedURL, err := a.normalizeURL(targetURL)
	if err != nil {
		result.Error = NewAnalysisError(ErrCodeInvalidURL, "Invalid URL format").WithDetails(err.Error())
		a.updateMetrics(startTime)
		return result
	}
	
	// Check circuit breaker
	if !a.circuitBreaker.CanExecute() {
		result.Error = NewAnalysisError(ErrCodeInternalError, "Service temporarily unavailable")
		a.updateMetrics(startTime)
		return result
	}
	
	// Execute analysis with circuit breaker
	err = a.circuitBreaker.Execute(func() error {
		return a.performAnalysis(ctx, parsedURL, result)
	})
	
	if err != nil {
		if result.Error == nil {
			result.Error = NewAnalysisError(ErrCodeInternalError, "Analysis failed").WithCause(err)
		}
		a.circuitBreaker.OnFailure()
	} else {
		a.circuitBreaker.OnSuccess()
	}
	
	// Cache the result
	a.cacheManager.Set(targetURL, result)
	
	// Update metrics
	a.updateMetrics(startTime)
	
	// Log completion
	logger.WithAnalysis(targetURL).Infow("Analysis completed",
		"total_ms", time.Since(startTime).Milliseconds(),
		"internal_links", result.InternalLinks,
		"external_links", result.ExternalLinks,
		"inaccessible_links", result.InaccessibleLinks,
		"headings", len(result.HeadingCounts),
		"login_form", result.HasLoginForm,
		"html_version", result.HTMLVersion,
		"title_len", len(result.PageTitle),
	)
	
	return result
}

// normalizeURL validates and normalizes the input URL
func (a *Analyzer) normalizeURL(targetURL string) (*url.URL, error) {
	// Add scheme if missing
	if !strings.HasPrefix(targetURL, "http://") && !strings.HasPrefix(targetURL, "https://") {
		targetURL = "https://" + targetURL
	}
	
	// Parse URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	
	// Validate scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, &url.Error{Op: "parse", URL: targetURL, Err: &url.Error{Op: "scheme", URL: targetURL, Err: &url.Error{Op: "unsupported", URL: targetURL}}}
	}
	
	return parsedURL, nil
}

// performAnalysis performs the actual web page analysis
func (a *Analyzer) performAnalysis(ctx context.Context, parsedURL *url.URL, result *AnalysisResult) error {
	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if err != nil {
		return err
	}
	
	// Set headers
	req.Header.Set("User-Agent", "WebPageAnalyzer/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	// Explicitly tell server not to compress
	req.Header.Set("Accept-Encoding", "identity")
	req.Header.Set("Connection", "keep-alive")
	
	// Get HTTP client from pool
	client := a.httpClientPool.Get().(*http.Client)
	defer a.httpClientPool.Put(client)
	
	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Debug: Log response headers
	logger.WithAnalysis(parsedURL.String()).Infow("HTTP response received", 
		"status", resp.StatusCode,
		"content_length", resp.ContentLength,
		"content_encoding", resp.Header.Get("Content-Encoding"),
		"content_type", resp.Header.Get("Content-Type"),
	)
	
	// Check response status
	if resp.StatusCode >= 400 {
		result.StatusCode = resp.StatusCode
		result.Error = NewAnalysisError(ErrCodeHTTPError, "HTTP request failed").WithStatusCode(resp.StatusCode)
		return nil
	}
	
	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	// Parse HTML
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		logger.WithAnalysis(parsedURL.String()).Errorw("HTML parsing failed", "error", err, "body_length", len(body))
		return err
	}
	
	// Check if parsing succeeded
	if doc == nil {
		logger.WithAnalysis(parsedURL.String()).Errorw("HTML parsing returned nil document", "body_length", len(body))
		return fmt.Errorf("HTML parsing returned nil document")
	}
	
	// Analyze document
	a.analyzeDocumentWithContext(ctx, doc, result, parsedURL, string(body))
	
	return nil
}

// updateMetrics updates performance metrics
func (a *Analyzer) updateMetrics(startTime time.Time) {
	duration := time.Since(startTime)
	a.metricsManager.updateMetrics(duration)
}

// Stop stops the analyzer and cleans up resources
func (a *Analyzer) Stop() {
	a.cacheManager.Stop()
}
