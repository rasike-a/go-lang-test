# Web Page Analyzer

A sophisticated Go web application that analyzes web pages and provides detailed information about their HTML structure, content, and characteristics.

## Features

- **HTML Version Detection**: Automatically detects HTML version from DOCTYPE declarations
- **Page Analysis**: Extracts page title and analyzes heading structure (H1-H6)
- **Link Analysis**: Counts internal vs external links and checks link accessibility
- **Login Form Detection**: Identifies pages containing login forms
- **Error Handling**: Provides detailed HTTP status codes and error descriptions
- **Modern Web Interface**: Clean, responsive UI with real-time analysis results
- **Template-Based Rendering**: Fast, maintainable HTML generation using templates
- **Enhanced User Experience**: Loading states, animations, and interactive feedback
- **Accessibility Features**: ARIA attributes, keyboard shortcuts, and screen reader support
- **Responsive Design**: Mobile-friendly layout with CSS custom properties

## Error Handling & Resilience

The application implements enterprise-grade error handling and resilience patterns:

### 🚨 Error Types & Codes

The system categorizes errors into structured types with specific codes:

| Error Code | Description | HTTP Status | Example Cause |
|------------|-------------|--------------|---------------|
| `INVALID_URL` | Malformed URL format | 400 Bad Request | Invalid URL syntax |
| `HTTP_ERROR` | HTTP response errors | 400/502 | 4xx/5xx status codes |
| `NETWORK_ERROR` | Network connectivity issues | 502 Bad Gateway | DNS failure, timeout |
| `PARSE_ERROR` | HTML parsing failures | 422 Unprocessable Entity | Malformed HTML |
| `TIMEOUT_ERROR` | Request timeout | 408 Request Timeout | Slow response |
| `INTERNAL_ERROR` | Application errors | 500 Internal Server Error | Internal failures |

### 🛡️ Resilience Features

#### Circuit Breaker Pattern
- **Automatic failure detection** after 5 consecutive failures
- **Recovery timeout** of 30 seconds before retry attempts
- **Graceful degradation** during service outages
- **Automatic recovery** after successful requests

#### Request Context & Timeouts
- **Request cancellation** support for client disconnections
- **Configurable timeouts** (default: 30 seconds)
- **Context-aware operations** throughout the request lifecycle
- **Resource cleanup** on timeout or cancellation

#### Panic Recovery
- **Automatic panic recovery** with detailed stack traces
- **Graceful error responses** instead of server crashes
- **Request isolation** - one panic doesn't affect other requests
- **Comprehensive logging** for debugging

### 📊 Error Response Format

All errors return structured JSON responses:

```json
{
  "url": "https://example.com",
  "error": {
    "code": "NETWORK_ERROR",
    "message": "Failed to fetch URL",
    "details": "dial tcp: lookup example.com: no such host",
    "url": "https://example.com",
    "timestamp": "2024-08-26T14:38:00Z",
    "status_code": 502
  },
  "status_code": 502
}
```

### 🔧 Error Handling Configuration

```go
// Circuit Breaker Configuration
circuitBreaker: NewCircuitBreaker(
    5,                    // Failure threshold
    30*time.Second,      // Recovery timeout
    2                     // Success threshold for recovery
)

// HTTP Client Configuration
httpClient: &http.Client{
    Timeout: 30 * time.Second,  // Request timeout
}
```

### 📝 Logging & Monitoring

- **Structured logging** with consistent key-value pairs
- **Performance metrics** (timing, bytes, status codes)
- **Error categorization** for monitoring and alerting
- **Request tracing** with unique identifiers

### 🚦 HTTP Status Code Mapping

The application maps internal error types to appropriate HTTP status codes:

| Internal Error | HTTP Status | Description |
|----------------|--------------|-------------|
| `INVALID_URL` | 400 | Client provided invalid URL |
| `HTTP_ERROR` (4xx) | 400 | Client-side HTTP errors |
| `HTTP_ERROR` (5xx) | 502 | Server-side HTTP errors |
| `NETWORK_ERROR` | 502 | Network connectivity issues |
| `PARSE_ERROR` | 422 | Content parsing failures |
| `TIMEOUT_ERROR` | 408 | Request timeout |
| `INTERNAL_ERROR` | 500 | Application errors |

### 🧪 Error Testing

The test suite includes comprehensive error scenarios:

```bash
# Test error handling
go test -v ./analyzer -run TestAnalyzeURL_InvalidURL
go test -v ./handlers -run TestAnalyzeHandler_HTTPError

# Test circuit breaker
go test -v ./analyzer -run TestCircuitBreaker

# Test timeout scenarios
go test -v ./analyzer -run TestAnalyzeURL_Timeout
```

## Frontend Architecture & User Experience

The application features a modern, maintainable frontend architecture with enhanced user experience:

### 🎨 Modern UI Components

#### Template-Based Rendering System
- **HTML Templates**: Separated HTML structure from JavaScript logic
- **ResultsRenderer Class**: Dedicated class for handling all rendering operations
- **Dynamic Content**: Fast template cloning instead of string concatenation
- **Maintainable Code**: HTML changes don't require JavaScript modifications

#### Enhanced CSS Design System
- **CSS Custom Properties**: Consistent theming with CSS variables
- **State-Based Styling**: Dynamic styling based on application state
- **Responsive Design**: Mobile-first approach with breakpoint system
- **Dark Mode Support**: Automatic dark mode detection and styling
- **Smooth Animations**: CSS transitions and keyframe animations

#### Interactive JavaScript Features
- **Loading States**: Animated spinner with user feedback
- **Error Handling**: Better error display with icons and styling
- **Keyboard Shortcuts**: Ctrl/Cmd+Enter to submit, Escape to clear
- **Form Validation**: Real-time validation with helpful text
- **Focus Management**: Visual feedback and accessibility improvements

### 🚀 Performance Improvements

#### Template System Benefits
- **Faster Rendering**: Template cloning outperforms string manipulation
- **Memory Efficiency**: Reduced memory allocation during rendering
- **Better Caching**: Browser can cache template elements
- **Cleaner Code**: Separation of concerns improves maintainability

#### User Experience Enhancements
- **Immediate Feedback**: Loading states and progress indicators
- **Error Recovery**: Clear error messages with actionable information
- **Accessibility**: ARIA labels, semantic HTML, and keyboard navigation
- **Responsive Interactions**: Touch-friendly interface for mobile devices

### 🔧 Frontend Architecture

```javascript
// Clean separation of concerns
const resultsRenderer = new ResultsRenderer(resultsDiv);

// Template-based rendering
resultsRenderer.renderResults(result);
resultsRenderer.renderLoading();
resultsRenderer.renderError(error);

// State-based styling
<div data-state="loading" class="results">
  <!-- Content automatically styled based on state -->
</div>
```

### 📱 Accessibility Features

- **ARIA Attributes**: Proper labeling and descriptions
- **Keyboard Navigation**: Full keyboard support for all interactions
- **Screen Reader Support**: Semantic HTML structure
- **Focus Management**: Clear visual focus indicators
- **Form Help Text**: Descriptive text for form inputs

## Middleware & Security

The application includes a comprehensive middleware stack for production readiness:

### 🛡️ Security Middleware

#### Security Headers
- **X-Content-Type-Options**: Prevents MIME type sniffing
- **X-Frame-Options**: Protects against clickjacking
- **X-XSS-Protection**: Enables browser XSS filtering
- **Referrer-Policy**: Controls referrer information

#### CORS Support
- **Cross-origin requests** support for API integration
- **Configurable origins** and methods
- **Preflight request handling** for complex requests

### 📊 Monitoring & Observability

#### Request Logging
- **Structured logging** with consistent format
- **Performance metrics** (duration, status codes)
- **Request details** (method, path, user agent)
- **Remote address tracking** for security

#### Panic Recovery
- **Automatic panic handling** with stack traces
- **Graceful error responses** instead of crashes
- **Request isolation** for stability
- **Comprehensive error logging**

### ⏱️ Performance & Reliability

#### Request Timeouts
- **Configurable timeouts** per request
- **Context cancellation** support
- **Resource cleanup** on timeout
- **Graceful degradation** under load

#### Middleware Chain
```go
middleware.Chain(
    handler,
    middleware.PanicRecovery,    // Panic recovery
    middleware.Logging,          // Request logging
    middleware.CORS,             // CORS support
    middleware.SecurityHeaders,  // Security headers
    middleware.Timeout(30*time.Second), // Request timeout
)
```

### 🔧 Middleware Configuration

```go
// Custom timeout configuration
timeout := 30 * time.Second
timeoutMiddleware := middleware.Timeout(timeout)

// Custom logging configuration
loggingMiddleware := middleware.Logging

// Security headers
securityMiddleware := middleware.SecurityHeaders
```

## Requirements

- Go 1.21 or later
- Internet connection for analyzing external web pages

## Installation and Setup

### 1. Clone the Repository

```bash
git clone <repository-url>
cd web-page-analyzer
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Build the Application

```bash
go build -o bin/web-page-analyzer .
```

### 4. Run the Application

```bash
# Run directly with go
go run main.go

# Or run the compiled binary
./bin/web-page-analyzer
```

The application will start on port 8080 by default. You can specify a different port using the `PORT` environment variable:

```bash
PORT=3000 go run main.go
```

### 5. Access the Application

Open your web browser and navigate to:
```
http://localhost:8080
```

## Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./analyzer
go test ./handlers
```

## API Endpoints

### GET /
Returns the main HTML interface for entering URLs to analyze.

### POST /analyze
Analyzes a web page URL and returns JSON results.

**Request Parameters:**
- `url` (form parameter): The URL to analyze

**Response Format:**
```json
{
  "url": "https://example.com",
  "html_version": "HTML5",
  "page_title": "Example Domain",
  "heading_counts": {
    "h1": 1,
    "h2": 3,
    "h3": 2
  },
  "internal_links": 5,
  "external_links": 3,
  "inaccessible_links": 1,
  "has_login_form": false,
  "status_code": 200
}
```

**Error Response:**
```json
{
  "url": "https://example.com/404",
  "error": "HTTP 404: Not Found",
  "status_code": 404
}
```

## Project Structure

```
├── main.go                 # Application entry point
├── analyzer/
│   ├── analyzer.go         # Core analysis logic
│   ├── analyzer_test.go    # Unit tests for analyzer
│   ├── errors.go           # Error types and handling
│   └── circuit_breaker.go  # Circuit breaker pattern
├── handlers/
│   ├── handlers.go         # HTTP handlers and web interface
│   └── handlers_test.go    # Integration tests for handlers
├── middleware/
│   └── middleware.go       # HTTP middleware stack
├── static/
│   ├── css/
│   │   └── styles.css      # Modern CSS with custom properties
│   ├── js/
│   │   ├── app.js          # Main application logic
│   │   └── resultsRenderer.js # Template-based rendering
│   └── templates/
│       └── results.html    # HTML templates for dynamic content
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── .gitignore             # Git ignore patterns
└── README.md              # Project documentation
```

## Architecture

The application follows a clean, layered architecture pattern with enterprise-grade features:

1. **Main Package**: Entry point, server setup, middleware configuration, and routing
2. **Analyzer Package**: Core business logic for web page analysis with error handling and resilience
3. **Handlers Package**: HTTP request handling and web interface with proper status codes
4. **Middleware Package**: Cross-cutting concerns including security, logging, and recovery

### Key Components

- **Analyzer**: Performs web page analysis with structured error handling, circuit breaker pattern, and context support
- **Server**: HTTP server with comprehensive middleware stack for security, monitoring, and reliability
- **Error Handling**: Structured error types with appropriate HTTP status codes and detailed context
- **Circuit Breaker**: Automatic failure detection and recovery for external HTTP calls
- **Middleware Stack**: Panic recovery, request logging, CORS, security headers, and timeout handling
- **Frontend**: Modern, responsive web interface with template-based rendering and enhanced user experience
- **ResultsRenderer**: Dedicated class for handling HTML rendering using templates
- **Template System**: HTML templates for maintainable and fast content generation
- **CSS Design System**: Modern styling with custom properties and state-based classes

## Deployment

### Local Development
```bash
go run main.go
```

### Production Build
```bash
# Build for current platform
go build -o web-page-analyzer .

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o web-page-analyzer-linux .

# Cross-compile for Windows
GOOS=windows GOARCH=amd64 go build -o web-page-analyzer.exe .
```

### Docker Deployment
Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o web-page-analyzer .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/web-page-analyzer .
EXPOSE 8080
CMD ["./web-page-analyzer"]
```

Build and run:
```bash
docker build -t web-page-analyzer .
docker run -p 8080:8080 web-page-analyzer
```

## Environment Variables

- `PORT`: Server port (default: 8080)

## Usage Examples

1. **Basic Analysis**: Enter any URL (e.g., `https://example.com`) in the web interface
2. **URL without Schema**: The application automatically adds `https://` if missing
3. **Error Handling**: Invalid URLs or inaccessible pages will show appropriate error messages
4. **Detailed Results**: View comprehensive analysis including HTML version, headings breakdown, and link statistics
5. **Interactive Features**: Use keyboard shortcuts (Ctrl/Cmd+Enter to submit, Escape to clear)
6. **Loading States**: Watch animated spinner during analysis with real-time feedback
7. **Responsive Design**: Use on any device with touch-friendly mobile interface
8. **Accessibility**: Full keyboard navigation and screen reader support

## Performance Considerations

- **Timeout**: HTTP requests timeout after 30 seconds to prevent hanging
- **Link Checking**: Uses HEAD requests for efficient link accessibility testing
- **Concurrent Safety**: Thread-safe design suitable for concurrent requests
- **Memory Efficient**: Streams HTML parsing without loading entire documents into memory
- **Circuit Breaker**: Prevents cascading failures and improves system resilience
- **Context Cancellation**: Efficient resource cleanup on client disconnection
- **Structured Logging**: Minimal performance impact with comprehensive observability
- **Template Rendering**: Fast HTML generation using DOM templates instead of string manipulation
- **CSS Optimization**: Efficient styling with CSS custom properties and minimal reflows
- **JavaScript Performance**: Clean separation of concerns with optimized rendering pipeline

## Security Features

- **Input Validation**: Validates and sanitizes URL inputs
- **Request Timeouts**: Prevents resource exhaustion from slow responses
- **Error Handling**: Secure error messages without exposing internal details
- **HTTPS Preference**: Automatically upgrades HTTP URLs to HTTPS when possible