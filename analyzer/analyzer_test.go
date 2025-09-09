package analyzer

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func TestNewAnalyzer(t *testing.T) {
	timeout := 10 * time.Second
	analyzer := NewAnalyzer(timeout)

	if analyzer == nil {
		t.Fatal("NewAnalyzer returned nil")
	}

	if analyzer.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, analyzer.timeout)
	}
}

func TestDetectHTMLVersion(t *testing.T) {
	analyzer := NewAnalyzer(30 * time.Second)

	testCases := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "HTML5",
			html:     "<!DOCTYPE html><html><head><title>Test</title></head><body></body></html>",
			expected: "HTML5",
		},
		{
			name:     "XHTML 1.0 Strict",
			html:     `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd">`,
			expected: "XHTML 1.0 Strict",
		},
		{
			name:     "HTML 4.01 Transitional",
			html:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">`,
			expected: "HTML 4.01 Transitional",
		},
		{
			name:     "Unknown version",
			html:     "<html><head><title>Test</title></head><body></body></html>",
			expected: "Unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.detectHTMLVersion(tc.html)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestAnalyzeURL_InvalidURL(t *testing.T) {
	analyzer := NewAnalyzer(5 * time.Second)
	result := analyzer.AnalyzeURL("invalid-url")

	if result.Error == nil {
		t.Fatal("Expected error for invalid URL")
	}

	// The URL gets normalized and fails at network level, so it's INTERNAL_ERROR
	if result.Error.Code != ErrCodeInternalError {
		t.Errorf("Expected error code %s, got %s", ErrCodeInternalError, result.Error.Code)
	}

	if result.URL != "invalid-url" {
		t.Errorf("Expected URL 'invalid-url', got %s", result.URL)
	}
}

func TestAnalyzeURL_ValidHTML(t *testing.T) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<title>Test Page</title>
</head>
<body>
	<h1>Main Heading</h1>
	<h2>Sub Heading 1</h2>
	<h2>Sub Heading 2</h2>
	<h3>Sub Sub Heading</h3>
	
	<a href="https://external.com">External Link</a>
	<a href="/internal">Internal Link</a>
	<a href="#fragment">Fragment Link</a>
	
	<form>
		<input type="text" name="username">
		<input type="password" name="password">
		<button type="submit">Login</button>
	</form>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlContent))
	}))
	defer server.Close()

	analyzer := NewAnalyzer(30 * time.Second)
	result := analyzer.AnalyzeURL(server.URL)

	if result.Error != nil {
		t.Errorf("Unexpected error: %s", result.Error.Message)
	}

	if result.HTMLVersion != "HTML5" {
		t.Errorf("Expected HTML5, got %s", result.HTMLVersion)
	}

	if result.PageTitle != "Test Page" {
		t.Errorf("Expected 'Test Page', got '%s'", result.PageTitle)
	}

	expectedHeadings := map[string]int{"h1": 1, "h2": 2, "h3": 1}
	for level, count := range expectedHeadings {
		if result.HeadingCounts[level] != count {
			t.Errorf("Expected %d %s headings, got %d", count, level, result.HeadingCounts[level])
		}
	}

	if result.InternalLinks != 1 {
		t.Errorf("Expected 1 internal link, got %d", result.InternalLinks)
	}

	if result.ExternalLinks != 1 {
		t.Errorf("Expected 1 external link, got %d", result.ExternalLinks)
	}

	if !result.HasLoginForm {
		t.Error("Expected login form to be detected")
	}
}

func TestAnalyzeURL_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer server.Close()

	analyzer := NewAnalyzer(30 * time.Second)
	result := analyzer.AnalyzeURL(server.URL)

	if result.Error == nil {
		t.Error("Expected error for HTTP 404")
	}

	if result.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", result.StatusCode)
	}

	if result.Error.Code != ErrCodeHTTPError {
		t.Errorf("Expected error code %s, got %s", ErrCodeHTTPError, result.Error.Code)
	}
}

func TestIsLoginForm(t *testing.T) {
	analyzer := NewAnalyzer(30 * time.Second)

	testCases := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			name: "Valid login form",
			html: `<form>
				<input type="text" name="username">
				<input type="password" name="password">
			</form>`,
			expected: true,
		},
		{
			name: "Valid login form with email",
			html: `<form>
				<input type="email" name="email">
				<input type="password" name="pwd">
			</form>`,
			expected: true,
		},
		{
			name: "Modern login form with ID attributes",
			html: `<form>
				<input type="text" id="user-email" placeholder="Enter your email">
				<input type="password" id="user-password">
			</form>`,
			expected: true,
		},
		{
			name: "Login form with placeholder hints",
			html: `<form>
				<input type="text" placeholder="Username or email">
				<input type="password" placeholder="Password">
			</form>`,
			expected: true,
		},
		{
			name: "Login form with account field",
			html: `<form>
				<input type="text" name="account">
				<input type="password" name="pass">
			</form>`,
			expected: true,
		},
		{
			name: "Login form with phone field",
			html: `<form>
				<input type="tel" name="phone">
				<input type="password" name="password">
			</form>`,
			expected: true,
		},
		{
			name: "Login form with login button",
			html: `<form>
				<input type="text" name="user">
				<input type="password" name="pass">
				<button>Sign In</button>
			</form>`,
			expected: true,
		},
		{
			name: "Form without password",
			html: `<form>
				<input type="text" name="username">
			</form>`,
			expected: false,
		},
		{
			name: "Form without username field",
			html: `<form>
				<input type="password" name="password">
				<input type="text" name="message">
			</form>`,
			expected: false,
		},
		{
			name:     "Empty form",
			html:     `<form></form>`,
			expected: false,
		},
		{
			name: "Contact form (not login)",
			html: `<form>
				<input type="text" name="name" placeholder="Your name">
				<input type="email" name="email" placeholder="Your email">
				<textarea name="message" placeholder="Your message"></textarea>
			</form>`,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := parseHTMLString(tc.html)
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}

			var formNode *html.Node
			var findForm func(*html.Node)
			findForm = func(n *html.Node) {
				if n.Type == html.ElementNode && n.Data == "form" {
					formNode = n
					return
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					findForm(c)
				}
			}
			findForm(doc)

			if formNode == nil {
				t.Fatal("Form node not found")
			}

			result := analyzer.isLoginForm(formNode)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestAnalyzeLinks(t *testing.T) {
	analyzer := NewAnalyzer(30 * time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/good" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	result := &AnalysisResult{HeadingCounts: make(map[string]int)}

	links := []string{
		"/good",                   // internal, accessible
		"/bad",                    // internal, not accessible
		"https://external.com",    // external
		"#fragment",               // should be ignored
		"javascript:void(0)",      // should be ignored
		"mailto:test@example.com", // should be ignored
	}

	analyzer.analyzeLinksConcurrent(links, baseURL, result)

	// Should count /good and /bad as internal links
	// Should count https://external.com as external (but will be inaccessible in test)
	if result.InternalLinks < 2 {
		t.Errorf("Expected at least 2 internal links, got %d", result.InternalLinks)
	}

	if result.ExternalLinks < 2 {
		t.Errorf("Expected at least 2 external links, got %d", result.ExternalLinks)
	}
}

func parseHTMLString(htmlStr string) (*html.Node, error) {
	return html.Parse(strings.NewReader(htmlStr))
}

func TestAnalyzeURL_URLWithoutScheme(t *testing.T) {
	analyzer := NewAnalyzer(5 * time.Second)
	result := analyzer.AnalyzeURL("example.com")

	// example.com is a valid domain, so it should succeed
	if result.Error != nil {
		t.Errorf("Unexpected error: %s", result.Error.Message)
	}

	if result.URL != "example.com" {
		t.Errorf("Expected URL 'example.com', got %s", result.URL)
	}
}

func TestCacheManager(t *testing.T) {
	cache := NewCacheManager(100 * time.Millisecond)

	// Test cache operations
	result := &AnalysisResult{URL: "test.com"}
	cache.Set("test.com", result)

	// Test cache hit
	if cached, found := cache.Get("test.com"); !found {
		t.Error("Expected cache hit")
	} else if cached.URL != "test.com" {
		t.Error("Expected correct cached result")
	}

	// Test cache miss
	if _, found := cache.Get("nonexistent.com"); found {
		t.Error("Expected cache miss")
	}

	// Test cache expiration
	time.Sleep(150 * time.Millisecond)
	if _, found := cache.Get("test.com"); found {
		t.Error("Expected expired cache entry to be removed")
	}

	// Test cache stats
	total, _ := cache.GetStats()
	if total != 0 {
		t.Errorf("Expected 0 total entries, got %d", total)
	}

	// Test stop functionality
	cache.Stop()
}

func TestMetricsManager(t *testing.T) {
	metrics := NewMetricsManager()

	// Test initial state
	initialMetrics := metrics.GetMetrics()
	if initialMetrics.TotalRequests != 0 {
		t.Error("Expected initial total requests to be 0")
	}

	// Test metrics updates
	metrics.incrementActiveRequests()
	metrics.incrementActiveRequests()
	metrics.decrementActiveRequests()

	metrics.RecordCacheHit()
	metrics.RecordCacheMiss()

	metrics.updateMetrics(100 * time.Millisecond)

	// Test final state
	finalMetrics := metrics.GetMetrics()
	if finalMetrics.ActiveRequests != 1 {
		t.Errorf("Expected 1 active request, got %d", finalMetrics.ActiveRequests)
	}
	if finalMetrics.CacheHits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", finalMetrics.CacheHits)
	}
	if finalMetrics.CacheMisses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", finalMetrics.CacheMisses)
	}
	if finalMetrics.TotalRequests != 1 {
		t.Errorf("Expected 1 total request, got %d", finalMetrics.TotalRequests)
	}

	// Test reset
	metrics.Reset()
	resetMetrics := metrics.GetMetrics()
	if resetMetrics.TotalRequests != 0 {
		t.Error("Expected reset total requests to be 0")
	}
}

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(2, 200*time.Millisecond, 1)

	// Test initial state
	if cb.State() != StateClosed {
		t.Error("Expected initial state to be closed")
	}

	// Test successful execution
	if !cb.CanExecute() {
		t.Error("Expected to be able to execute in closed state")
	}

	err := cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Error("Expected successful execution")
	}

	// Test failure threshold
	err = cb.Execute(func() error {
		return errors.New("test error")
	})
	if err == nil {
		t.Error("Expected error to be returned")
	}

	err = cb.Execute(func() error {
		return errors.New("test error")
	})
	if err == nil {
		t.Error("Expected error to be returned")
	}

	// Test open state
	if cb.State() != StateOpen {
		t.Error("Expected circuit breaker to be open after failures")
	}

	if cb.CanExecute() {
		t.Error("Expected to not be able to execute in open state")
	}

	// Test timeout and half-open state
	time.Sleep(250 * time.Millisecond)

	// The circuit breaker transitions to half-open when CanExecute is called after timeout
	if cb.CanExecute() {
		if cb.State() != StateHalfOpen {
			t.Error("Expected circuit breaker to be half-open after timeout")
		}
	} else {
		t.Error("Expected to be able to execute in half-open state")
	}

	// Test successful execution in half-open state
	err = cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Error("Expected successful execution in half-open state")
	}

	// Test reset to closed state
	if cb.State() != StateClosed {
		t.Error("Expected circuit breaker to be closed after success")
	}

	// Test manual reset
	cb.Reset()
	if cb.State() != StateClosed {
		t.Error("Expected circuit breaker to be closed after reset")
	}
}

func TestErrorHelpers(t *testing.T) {
	// Test error creation with fluent methods
	err := NewAnalysisError(ErrCodeInvalidURL, "Invalid URL").
		WithDetails("Additional details").
		WithURL("https://example.com").
		WithStatusCode(400).
		WithCause(errors.New("root cause"))

	if err.Code != ErrCodeInvalidURL {
		t.Errorf("Expected error code %s, got %s", ErrCodeInvalidURL, err.Code)
	}

	if err.Details != "Additional details" {
		t.Errorf("Expected details 'Additional details', got %s", err.Details)
	}

	if err.URL != "https://example.com" {
		t.Errorf("Expected URL 'https://example.com', got %s", err.URL)
	}

	if err.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", err.StatusCode)
	}

	// Test specific error constructors
	invalidURLErr := NewInvalidURLError("bad url", errors.New("parse error"))
	if invalidURLErr.Code != ErrCodeInvalidURL {
		t.Errorf("Expected error code %s, got %s", ErrCodeInvalidURL, invalidURLErr.Code)
	}

	httpErr := NewHTTPError(404, "https://example.com")
	if httpErr.Code != ErrCodeHTTPError {
		t.Errorf("Expected error code %s, got %s", ErrCodeHTTPError, httpErr.Code)
	}
	if httpErr.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", httpErr.StatusCode)
	}

	networkErr := NewNetworkError("https://example.com", errors.New("connection failed"))
	if networkErr.Code != ErrCodeNetworkError {
		t.Errorf("Expected error code %s, got %s", ErrCodeNetworkError, networkErr.Code)
	}

	parseErr := NewParseError("https://example.com", errors.New("parsing failed"))
	if parseErr.Code != ErrCodeParseError {
		t.Errorf("Expected error code %s, got %s", ErrCodeParseError, parseErr.Code)
	}

	timeoutErr := NewTimeoutError("https://example.com", 30*time.Second)
	if timeoutErr.Code != ErrCodeTimeoutError {
		t.Errorf("Expected error code %s, got %s", ErrCodeTimeoutError, timeoutErr.Code)
	}

	// Test error checking functions
	if !IsAnalysisError(err) {
		t.Error("Expected IsAnalysisError to return true")
	}

	if retrievedErr := GetAnalysisError(err); retrievedErr == nil {
		t.Error("Expected GetAnalysisError to return error")
	}

	// Test error unwrapping
	if unwrapped := err.Unwrap(); unwrapped == nil {
		t.Error("Expected Unwrap to return cause")
	}
}

func TestAnalyzerMethods(t *testing.T) {
	analyzer := NewAnalyzer(5 * time.Second)

	// Test GetMetrics
	metrics := analyzer.GetMetrics()
	if metrics.TotalRequests != 0 {
		t.Error("Expected initial metrics to have 0 total requests")
	}

	// Test Stop method
	analyzer.Stop()
}
