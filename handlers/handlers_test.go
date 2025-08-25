package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"web-page-analyzer/analyzer"
)

func TestNewServer(t *testing.T) {
	server := NewServer()
	
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
	
	if server.analyzer == nil {
		t.Error("Server analyzer is nil")
	}
	
	if server.template == nil {
		t.Error("Server template is nil")
	}
}

func TestIndexHandler_GET(t *testing.T) {
	server := NewServer()
	
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	server.IndexHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
	
	contentType := rr.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Expected Content-Type text/html, got %s", contentType)
	}
	
	body := rr.Body.String()
	if !strings.Contains(body, "Web Page Analyzer") {
		t.Error("Expected 'Web Page Analyzer' in response body")
	}
	
	if !strings.Contains(body, `<form id="analyzeForm">`) {
		t.Error("Expected form element in response body")
	}
}

func TestIndexHandler_POST(t *testing.T) {
	server := NewServer()
	
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	server.IndexHandler(rr, req)
	
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, status)
	}
}

func TestAnalyzeHandler_GET(t *testing.T) {
	server := NewServer()
	
	req, err := http.NewRequest("GET", "/analyze", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	server.AnalyzeHandler(rr, req)
	
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, status)
	}
}

func TestAnalyzeHandler_EmptyURL(t *testing.T) {
	server := NewServer()
	
	form := url.Values{}
	form.Add("url", "")
	
	req, err := http.NewRequest("POST", "/analyze", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	
	rr := httptest.NewRecorder()
	server.AnalyzeHandler(rr, req)
	
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, status)
	}
}

func TestAnalyzeHandler_ValidURL(t *testing.T) {
	// Create a test server that serves HTML content
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Test Page</title>
</head>
<body>
	<h1>Test Heading</h1>
	<a href="https://example.com">External Link</a>
	<a href="/internal">Internal Link</a>
</body>
</html>`))
	}))
	defer testServer.Close()
	
	server := NewServer()
	
	form := url.Values{}
	form.Add("url", testServer.URL)
	
	req, err := http.NewRequest("POST", "/analyze", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	
	rr := httptest.NewRecorder()
	server.AnalyzeHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
	
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
	
	var result analyzer.AnalysisResult
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON response: %v", err)
	}
	
	if result.Error != "" {
		t.Errorf("Unexpected error in result: %s", result.Error)
	}
	
	if result.HTMLVersion != "HTML5" {
		t.Errorf("Expected HTML5, got %s", result.HTMLVersion)
	}
	
	if result.PageTitle != "Test Page" {
		t.Errorf("Expected 'Test Page', got '%s'", result.PageTitle)
	}
	
	if result.HeadingCounts["h1"] != 1 {
		t.Errorf("Expected 1 h1 heading, got %d", result.HeadingCounts["h1"])
	}
	
	if result.ExternalLinks != 1 {
		t.Errorf("Expected 1 external link, got %d", result.ExternalLinks)
	}
	
	if result.InternalLinks != 1 {
		t.Errorf("Expected 1 internal link, got %d", result.InternalLinks)
	}
}

func TestAnalyzeHandler_InvalidURL(t *testing.T) {
	server := NewServer()
	
	form := url.Values{}
	form.Add("url", "invalid-url")
	
	req, err := http.NewRequest("POST", "/analyze", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	
	rr := httptest.NewRecorder()
	server.AnalyzeHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
	
	var result analyzer.AnalysisResult
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON response: %v", err)
	}
	
	if result.Error == "" {
		t.Error("Expected error for invalid URL")
	}
	
	if !strings.Contains(result.Error, "Invalid URL") {
		t.Errorf("Expected 'Invalid URL' in error message, got: %s", result.Error)
	}
}

func TestAnalyzeHandler_HTTPError(t *testing.T) {
	// Create a test server that returns 404
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer testServer.Close()
	
	server := NewServer()
	
	form := url.Values{}
	form.Add("url", testServer.URL)
	
	req, err := http.NewRequest("POST", "/analyze", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	
	rr := httptest.NewRecorder()
	server.AnalyzeHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
	
	var result analyzer.AnalysisResult
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON response: %v", err)
	}
	
	if result.Error == "" {
		t.Error("Expected error for HTTP 404")
	}
	
	if result.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", result.StatusCode)
	}
	
	if !strings.Contains(result.Error, "404") {
		t.Errorf("Expected '404' in error message, got: %s", result.Error)
	}
}

func TestAnalyzeHandler_LoginFormDetection(t *testing.T) {
	// Create a test server that serves HTML with login form
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Login Page</title>
</head>
<body>
	<form>
		<input type="text" name="username" placeholder="Username">
		<input type="password" name="password" placeholder="Password">
		<button type="submit">Login</button>
	</form>
</body>
</html>`))
	}))
	defer testServer.Close()
	
	server := NewServer()
	
	form := url.Values{}
	form.Add("url", testServer.URL)
	
	req, err := http.NewRequest("POST", "/analyze", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	
	rr := httptest.NewRecorder()
	server.AnalyzeHandler(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
	
	var result analyzer.AnalysisResult
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON response: %v", err)
	}
	
	if result.Error != "" {
		t.Errorf("Unexpected error: %s", result.Error)
	}
	
	if !result.HasLoginForm {
		t.Error("Expected login form to be detected")
	}
}