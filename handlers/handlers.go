package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"time"

	"web-page-analyzer/analyzer"
	"web-page-analyzer/logger"
)

type Server struct {
	analyzer *analyzer.Analyzer
	template *template.Template
}

// NewServer creates a new server instance
func NewServer() *Server {
	// Create analyzer instance
	analyzer := analyzer.NewAnalyzer(60 * time.Second) // Increased timeout for complex sites

	// Control cache logging verbosity based on environment
	if os.Getenv("CACHE_VERBOSE") == "true" || os.Getenv("ENV") == "development" {
		analyzer.SetCacheVerbose(true)
	} else {
		analyzer.SetCacheVerbose(false)
	}

	tmpl := template.Must(template.New("index").Parse(indexHTML))

	return &Server{
		analyzer: analyzer,
		template: tmpl,
	}
}

// GetAnalyzer returns the analyzer instance for metrics collection
func (s *Server) GetAnalyzer() *analyzer.Analyzer {
	return s.analyzer
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := s.template.Execute(w, nil); err != nil {
		logger.Sugar.Errorw("Template execution error", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (s *Server) AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		http.Error(w, "URL parameter is required", http.StatusBadRequest)
		return
	}

	// Use context-aware analyzer
	result := s.analyzer.AnalyzeURLWithContext(r.Context(), url)

	// Set appropriate HTTP status code based on result
	statusCode := http.StatusOK
	if result.Error != nil {
		switch result.Error.Code {
		case analyzer.ErrCodeInvalidURL:
			statusCode = http.StatusBadRequest
		case analyzer.ErrCodeHTTPError:
			if result.StatusCode >= 400 && result.StatusCode < 500 {
				statusCode = http.StatusBadRequest
			} else if result.StatusCode >= 500 {
				statusCode = http.StatusBadGateway
			}
		case analyzer.ErrCodeNetworkError:
			statusCode = http.StatusBadGateway
		case analyzer.ErrCodeParseError:
			statusCode = http.StatusUnprocessableEntity
		case analyzer.ErrCodeTimeoutError:
			statusCode = http.StatusRequestTimeout
		default:
			statusCode = http.StatusInternalServerError
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Sugar.Errorw("JSON encoding error", "error", err)
		// Don't change status code here as we've already written it
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

const indexHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Web Page Analyzer</title>
    <link rel="stylesheet" href="/static/css/styles.css">
</head>
<body>
    <div class="container">
        <div class="main-content">
            <div class="header">
                <h1 class="title">Web Page Analyzer</h1>
                <p class="subtitle">Analyze web pages for HTML structure, content, and accessibility</p>
            </div>
            
            <div class="card">
                <form id="analyzeForm" role="form" aria-label="URL Analysis Form">
                    <div class="form-group">
                        <label for="url" class="form-label" id="url-label">Enter URL to analyze</label>
                        <input type="url" id="url" name="url" class="form-input" required 
                               placeholder="https://example.com" 
                               aria-labelledby="url-label"
                               aria-describedby="url-help"
                               data-validation="url">
                        <div id="url-help" class="form-help">Enter a valid web address to analyze</div>
                    </div>
                    <button type="submit" id="submitBtn" class="btn btn-primary" 
                            aria-live="polite"
                            data-loading-text="Analyzing..."
                            data-default-text="Analyze Page">Analyze Page</button>
                </form>
                
                <div id="results" class="results" role="region" aria-live="polite" aria-label="Analysis Results"></div>
            </div>
        </div>
    </div>

    <!-- HTML Templates -->
    <div id="templates" style="display: none;">
        <template id="resultsTemplate">
            <h2 class="results-header">Analysis Results</h2>
            
            <div class="result-item">
                <div class="result-label">URL</div>
                <div class="result-value" data-field="url"></div>
            </div>
            
            <div class="result-item">
                <div class="result-label">HTML Version</div>
                <div class="result-value" data-field="html_version"></div>
            </div>
            
            <div class="result-item">
                <div class="result-label">Page Title</div>
                <div class="result-value" data-field="page_title"></div>
            </div>
            
            <div class="result-item">
                <div class="result-label">Headings</div>
                <div class="result-value" data-field="headings"></div>
            </div>
            
            <div class="result-item">
                <div class="result-label">Links</div>
                <div class="result-value" data-field="links"></div>
            </div>
            
            <div class="result-item">
                <div class="result-label">Login Form</div>
                <div class="result-value" data-field="login_form"></div>
            </div>
        </template>

        <template id="headingsTemplate">
            <ul class="headings-list">
                <li data-template="heading-item">
                    <strong data-field="level"></strong>: <span data-field="count"></span>
                </li>
            </ul>
        </template>

        <template id="loadingTemplate">
            <div class="loading-state">
                <div class="loading-spinner"></div>
                <div class="loading-message">Analyzing web page, please wait...</div>
            </div>
        </template>

        <template id="errorTemplate">
            <div class="error-state">
                <div class="error-icon">⚠️</div>
                <div class="error-message" data-field="message"></div>
            </div>
        </template>
    </div>

    <!-- JavaScript Files -->
    <script src="/static/js/resultsRenderer.js"></script>
    <script src="/static/js/app.js"></script>
</body>
</html>`
