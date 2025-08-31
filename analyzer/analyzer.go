package analyzer

import (
	"context"
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

// Analyzer is the main analyzer that orchestrates web page analysis
type Analyzer struct {
	httpClient     *http.Client
	timeout        time.Duration
	logger         *log.Logger
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

	// Create logger
	logger := log.New(os.Stdout, "analyzer ", log.LstdFlags)

	analyzer := &Analyzer{
		httpClient:     httpClient,
		timeout:        timeout,
		logger:         logger,
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second, 2),
		httpClientPool: httpClientPool,
		cacheManager:   NewCacheManager(5*time.Minute, logger),
		metricsManager: NewMetricsManager(),
	}

	return analyzer
}

// SetLogger sets a custom logger for the analyzer
func (a *Analyzer) SetLogger(logger *log.Logger) {
	a.logger = logger
	a.cacheManager.logger = logger
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
	a.logger.Printf("analyzer analyze_done url=%q total_ms=%d internal=%d external=%d inaccessible=%d headings=%d login_form=%t html_version=%q title_len=%d",
		targetURL, time.Since(startTime).Milliseconds(), result.InternalLinks, result.ExternalLinks, result.InaccessibleLinks, len(result.HeadingCounts), result.HasLoginForm, result.HTMLVersion, len(result.PageTitle))
	
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
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	
	// Get HTTP client from pool
	client := a.getHTTPClient()
	defer a.putHTTPClient(client)
	
	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode >= 400 {
		result.StatusCode = resp.StatusCode
		result.Error = NewAnalysisError(ErrCodeHTTPError, "HTTP request failed").WithStatusCode(resp.StatusCode)
		return nil
	}
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	// Parse HTML
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return err
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
