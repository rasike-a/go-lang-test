# Web Page Analyzer

A sophisticated Go web application that analyzes web pages and provides detailed information about their HTML structure, content, and characteristics.

## Features

- **HTML Version Detection**: Automatically detects HTML version from DOCTYPE declarations
- **Page Analysis**: Extracts page title and analyzes heading structure (H1-H6)
- **Link Analysis**: Counts internal vs external links and checks link accessibility
- **Login Form Detection**: Identifies pages containing login forms
- **Error Handling**: Provides detailed HTTP status codes and error descriptions
- **Modern Web Interface**: Clean, responsive UI with real-time analysis results

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
│   └── analyzer_test.go    # Unit tests for analyzer
├── handlers/
│   ├── handlers.go         # HTTP handlers and web interface
│   └── handlers_test.go    # Integration tests for handlers
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
└── README.md              # Project documentation
```

## Architecture

The application follows a clean architecture pattern:

1. **Main Package**: Entry point, server setup, and routing
2. **Analyzer Package**: Core business logic for web page analysis
3. **Handlers Package**: HTTP request handling and web interface

### Key Components

- **Analyzer**: Performs the actual web page analysis including HTML parsing, link checking, and form detection
- **Server**: HTTP server with handlers for the web interface and API endpoints
- **HTML Template**: Embedded responsive web interface with JavaScript for dynamic interaction

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

## Performance Considerations

- **Timeout**: HTTP requests timeout after 30 seconds to prevent hanging
- **Link Checking**: Uses HEAD requests for efficient link accessibility testing
- **Concurrent Safety**: Thread-safe design suitable for concurrent requests
- **Memory Efficient**: Streams HTML parsing without loading entire documents into memory

## Security Features

- **Input Validation**: Validates and sanitizes URL inputs
- **Request Timeouts**: Prevents resource exhaustion from slow responses
- **Error Handling**: Secure error messages without exposing internal details
- **HTTPS Preference**: Automatically upgrades HTTP URLs to HTTPS when possible