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
	
	result := s.analyzer.AnalyzeURL(url)
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("JSON encoding error: %v", err)
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
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            text-align: center;
            margin-bottom: 30px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            font-weight: 600;
            color: #555;
        }
        input[type="url"] {
            width: 100%;
            padding: 12px;
            border: 2px solid #ddd;
            border-radius: 5px;
            font-size: 16px;
            transition: border-color 0.3s;
        }
        input[type="url"]:focus {
            outline: none;
            border-color: #007bff;
        }
        button {
            background-color: #007bff;
            color: white;
            padding: 12px 24px;
            border: none;
            border-radius: 5px;
            font-size: 16px;
            cursor: pointer;
            transition: background-color 0.3s;
        }
        button:hover {
            background-color: #0056b3;
        }
        button:disabled {
            background-color: #ccc;
            cursor: not-allowed;
        }
        .results {
            margin-top: 30px;
            padding: 20px;
            border: 1px solid #ddd;
            border-radius: 5px;
            background-color: #f9f9f9;
            display: none;
        }
        .error {
            color: #dc3545;
            background-color: #f8d7da;
            border: 1px solid #f5c6cb;
            padding: 15px;
            border-radius: 5px;
            margin-top: 20px;
        }
        .result-item {
            margin-bottom: 15px;
            padding: 10px;
            background-color: white;
            border-left: 4px solid #007bff;
        }
        .result-label {
            font-weight: 600;
            color: #333;
        }
        .result-value {
            margin-top: 5px;
            color: #666;
        }
        .headings-list {
            list-style: none;
            padding: 0;
        }
        .headings-list li {
            background-color: #e9ecef;
            margin: 5px 0;
            padding: 5px 10px;
            border-radius: 3px;
        }
        .loading {
            text-align: center;
            color: #666;
            font-style: italic;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Web Page Analyzer</h1>
        <form id="analyzeForm">
            <div class="form-group">
                <label for="url">Enter URL to analyze:</label>
                <input type="url" id="url" name="url" required placeholder="https://example.com">
            </div>
            <button type="submit" id="submitBtn">Analyze Page</button>
        </form>
        
        <div id="results" class="results"></div>
    </div>

    <script>
        document.getElementById('analyzeForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const url = document.getElementById('url').value;
            const submitBtn = document.getElementById('submitBtn');
            const resultsDiv = document.getElementById('results');
            
            submitBtn.disabled = true;
            submitBtn.textContent = 'Analyzing...';
            resultsDiv.style.display = 'block';
            resultsDiv.innerHTML = '<div class="loading">Analyzing web page, please wait...</div>';
            
            try {
                const formData = new FormData();
                formData.append('url', url);
                
                const response = await fetch('/analyze', {
                    method: 'POST',
                    body: formData
                });
                
                const result = await response.json();
                displayResults(result);
            } catch (error) {
                resultsDiv.innerHTML = '<div class="error">Error: Failed to analyze the page. Please try again.</div>';
            } finally {
                submitBtn.disabled = false;
                submitBtn.textContent = 'Analyze Page';
            }
        });
        
        function displayResults(result) {
            const resultsDiv = document.getElementById('results');
            
            if (result.error) {
                let errorMsg = result.error;
                if (result.status_code) {
                    errorMsg = 'HTTP ' + result.status_code + ': ' + result.error;
                }
                resultsDiv.innerHTML = '<div class="error">' + errorMsg + '</div>';
                return;
            }
            
            let headingsList = '';
            if (Object.keys(result.heading_counts).length > 0) {
                headingsList = '<ul class="headings-list">';
                for (const [level, count] of Object.entries(result.heading_counts)) {
                    headingsList += '<li><strong>' + level.toUpperCase() + ':</strong> ' + count + '</li>';
                }
                headingsList += '</ul>';
            } else {
                headingsList = '<em>No headings found</em>';
            }
            
            resultsDiv.innerHTML = '
                <h2>Analysis Results</h2>
                <div class="result-item">
                    <div class="result-label">URL:</div>
                    <div class="result-value">' + result.url + '</div>
                </div>
                <div class="result-item">
                    <div class="result-label">HTML Version:</div>
                    <div class="result-value">' + result.html_version + '</div>
                </div>
                <div class="result-item">
                    <div class="result-label">Page Title:</div>
                    <div class="result-value">' + (result.page_title || '<em>No title found</em>') + '</div>
                </div>
                <div class="result-item">
                    <div class="result-label">Headings:</div>
                    <div class="result-value">' + headingsList + '</div>
                </div>
                <div class="result-item">
                    <div class="result-label">Links:</div>
                    <div class="result-value">
                        <strong>Internal:</strong> ' + result.internal_links + '<br>
                        <strong>External:</strong> ' + result.external_links + '<br>
                        <strong>Inaccessible:</strong> ' + result.inaccessible_links + '
                    </div>
                </div>
                <div class="result-item">
                    <div class="result-label">Login Form:</div>
                    <div class="result-value">' + (result.has_login_form ? 'Yes' : 'No') + '</div>
                </div>
            ';
        }
    </script>
</body>
</html>
`