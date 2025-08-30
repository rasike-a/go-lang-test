package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"web-page-analyzer/analyzer"
)

type Server struct {
	analyzer *analyzer.Analyzer
	template *template.Template
}

func NewServer() *Server {
	tmpl := template.Must(template.New("index").Parse(indexHTML))
	
	return &Server{
		analyzer: analyzer.NewAnalyzer(30 * time.Second),
		template: tmpl,
	}
}

func (s *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "text/html")
	if err := s.template.Execute(w, nil); err != nil {
		log.Printf("Template execution error: %v", err)
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
		log.Printf("JSON encoding error: %v", err)
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
                <form id="analyzeForm">
                    <div class="form-group">
                        <label for="url" class="form-label">Enter URL to analyze</label>
                        <input type="url" id="url" name="url" class="form-input" required placeholder="https://example.go">
                    </div>
                    <button type="submit" id="submitBtn" class="btn btn-primary">Analyze Page</button>
                </form>
                
                <div id="results" class="results"></div>
            </div>
        </div>
    </div>

    <script src="/static/js/app.js"></script>
</body>
</html>`
